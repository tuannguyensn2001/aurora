package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dgraph-io/badger/v4"
)

const (
	defaultRefreshRate = 1 * time.Minute
	defaultLogLevel    = slog.LevelDebug
	defaultPath        = "/sdk-dump"
)

type Client interface {
	Start(ctx context.Context)
	Stop()
	EvaluateParameter(ctx context.Context, parameterName string, attribute *attribute) rolloutValue
}

type client struct {
	s3BucketName string
	refreshRate  time.Duration
	logLevel     slog.Level
	logger       *slog.Logger
	quit         chan struct{}
	s3Client     *s3.Client
	inMemoryOnly bool
	path         string
	db           *badger.DB
	engine       *engine
}

type ClientOptions struct {
	S3BucketName string
}

type Option func(*client)

func WithRefreshRate(refreshRate time.Duration) Option {
	return func(c *client) {
		c.refreshRate = refreshRate
	}
}

func WithLogLevel(logLevel slog.Level) Option {
	return func(c *client) {
		c.logLevel = logLevel
	}
}

func WithS3Client(s3Client *s3.Client) Option {
	return func(c *client) {
		c.s3Client = s3Client
	}
}

func WithInMemoryOnly(inMemoryOnly bool) Option {
	return func(c *client) {
		c.inMemoryOnly = inMemoryOnly
	}
}

func WithPath(path string) Option {
	return func(c *client) {
		c.path = path
	}
}

func (c *client) applyDefaults() {
	c.refreshRate = defaultRefreshRate
	c.logLevel = defaultLogLevel
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		c.logger.ErrorContext(context.Background(), "failed to load default config", "error", err)
	}
	c.s3Client = s3.NewFromConfig(cfg)
	c.inMemoryOnly = false
	c.path = defaultPath
	c.quit = make(chan struct{})
}

func NewClient(clientOptions ClientOptions, options ...Option) *client {
	c := &client{
		s3BucketName: clientOptions.S3BucketName,
	}
	c.applyDefaults()
	for _, option := range options {
		option(c)
	}
	opts := &slog.HandlerOptions{
		Level: c.logLevel,
	}
	c.logger = slog.New(slog.NewTextHandler(os.Stdout, opts))

	opt := badger.DefaultOptions(c.path)
	if c.inMemoryOnly {
		opt = opt.WithInMemory(true)
	}
	db, err := badger.Open(opt)
	if err != nil {
		panic(fmt.Errorf("failed to open badger db: %w", err))
	}
	c.db = db
	c.engine = newEngine(c.logger)
	return c
}

func (c *client) Start(ctx context.Context) {
	c.logger.Info("starting")
	if ctx.Err() != nil {
		c.logger.ErrorContext(ctx, "context is done")
	}
	err := c.persist(ctx)
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to persist parameters", "error", err)
	}
	go c.dispatch(ctx)
}

func (c *client) Stop() {
	c.db.Close()
	close(c.quit)
}

func (c *client) dispatch(ctx context.Context) {
	c.logger.Info("dispatching")
	ticket := time.NewTicker(c.refreshRate)

	for {
		select {
		case <-ticket.C:
			c.logger.Info("dispatching")
			err := c.persist(ctx)
			if err != nil {
				c.logger.ErrorContext(ctx, "failed to get parameters", "error", err)
			}

		case <-ctx.Done():
			ticket.Stop()
			return
		case <-c.quit:
			ticket.Stop()
			return
		}

	}
}

func (c *client) persist(ctx context.Context) error {

	parameters, err := c.getParameters(ctx)
	if err != nil {
		return err
	}
	err = c.persistParameters(ctx, parameters)
	if err != nil {
		return err
	}
	c.logger.Info("parameters persisted")
	return nil
}

func (c *client) getParameters(ctx context.Context) ([]Parameter, error) {
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(c.s3BucketName),
		Key:    aws.String("parameters.json"),
	}
	getObjectOutput, err := c.s3Client.GetObject(ctx, getObjectInput)
	if err != nil {
		return nil, err
	}

	parameters := []Parameter{}
	err = json.NewDecoder(getObjectOutput.Body).Decode(&parameters)
	if err != nil {
		return nil, err
	}
	return parameters, nil
}

func (c *client) persistParameters(ctx context.Context, parameters []Parameter) error {

	for _, parameter := range parameters {
		jsonParameters, err := json.Marshal(parameter)
		if err != nil {
			return err
		}
		err = c.db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(parameter.Name), jsonParameters)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *client) EvaluateParameter(ctx context.Context, parameterName string, attribute *attribute) rolloutValue {

	return c.resolveFromParameter(ctx, parameterName, attribute)
}

func (c *client) resolveFromParameter(ctx context.Context, parameterName string, attribute *attribute) rolloutValue {

	c.logger.InfoContext(ctx, "resolving parameter", "parameterName", parameterName)
	var parameter Parameter
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(parameterName))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &parameter)
		})
	})
	if err != nil {
		c.logger.ErrorContext(ctx, "failed to get parameter", "error", err)
		return rolloutValue{}
	}

	rolloutValueStr := c.engine.evaluateParameter(&parameter, attribute)
	c.logger.InfoContext(ctx, "resolved parameter", "parameterName", parameterName, "rolloutValue", rolloutValueStr, "dataType", parameter.DataType, "rules count", len(parameter.Rules))
	return NewRolloutValue(&rolloutValueStr, parameter.DataType)

}
