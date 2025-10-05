package fx

import (
	"api/config"
	"context"
	"log/slog"
	"sdk"
	"time"

	"go.uber.org/fx"
)

type SDKParams struct {
	fx.In
	Cfg *config.Config
}

func ProvideSDK(lc fx.Lifecycle, params SDKParams) sdk.Client {
	client := sdk.NewClient(sdk.ClientOptions{
		S3BucketName: params.Cfg.S3.BucketName,
		EndpointUrl:  "http://localhost:9000",
	}, sdk.WithPath("sdk-dump"), sdk.WithLogLevel(slog.LevelError), sdk.WithRefreshRate(10*time.Second), sdk.WithEnableS3(false))
	slog.Info("before start")
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			client.Stop()
			return nil
		},
		OnStart: func(ctx context.Context) error {
			client.Start(context.WithoutCancel(ctx))

			str := client.EvaluateParameter(context.TODO(), "enableViewAgent", sdk.NewAttribute().SetBool("is_active", true).SetNumber("age", 14)).AsBool(false)
			slog.Info("str", "str", str)

			return nil
		},
	})
	return client
}

var SDKModule = fx.Module("sdk",
	fx.Provide(ProvideSDK),
)
