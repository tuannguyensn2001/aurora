package fx

import (
	"context"

	config2 "api/config"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/fx"
)

type S3Params struct {
	fx.In
	Config *config2.Config
}

func ProvideS3(params S3Params) *s3.Client {
	if !params.Config.S3.Enable {
		return nil
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	return s3.NewFromConfig(cfg)
}

var S3Module = fx.Module("s3",
	fx.Provide(ProvideS3),
)
