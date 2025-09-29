package fx

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/fx"
)

type S3Params struct {
	fx.In
}

func ProvideS3(params S3Params) *s3.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	return s3.NewFromConfig(cfg)
}

var S3Module = fx.Module("s3",
	fx.Provide(ProvideS3),
)
