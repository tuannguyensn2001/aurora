package fx

import (
	"api/config"
	"context"
	"fmt"
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
	client, err := sdk.NewClient(sdk.ClientOptions{
		S3BucketName: params.Cfg.S3.BucketName,
		EndpointURL:  fmt.Sprintf("http://localhost:%d", params.Cfg.Service.Port),
	},
		sdk.WithPath("sdk-dump"),
		sdk.WithLogLevel(slog.LevelError),
		sdk.WithRefreshRate(1*time.Minute),
		sdk.WithEnableS3(true),
		sdk.WithOnEvaluate(func(source string, parameterName string, attribute *sdk.Attribute, rolloutValueRaw *string, err error) {
			// slog.Info("SDK evaluated parameter", "source", source, "parameterName", parameterName, "attribute", attribute, "rolloutValueRaw", *rolloutValueRaw, "error", err)
			slog.Info("SDK evaluated parameter", "source", source, "parameterName", parameterName, "attribute", attribute, "rolloutValueRaw", rolloutValueRaw, "error", err)
		}),
	)
	if err != nil {
		panic(err) // In production, handle this more gracefully
	}

	slog.Info("before start")
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			client.Stop()
			return nil
		},
		OnStart: func(ctx context.Context) error {
			err := client.Start(context.WithoutCancel(ctx))
			if err != nil {
				slog.Error("Failed to start SDK client", "error", err)
				return err
			}

			result := client.EvaluateParameter(context.TODO(), "enableViewAgent", sdk.NewAttribute().SetBool("is_active", true).SetNumber("age", 14))
			if result.HasError() {
				slog.Error("Failed to evaluate parameter", "error", result.Error())
			} else {
				str := result.AsBool(false)
				slog.Info("parameter evaluation result", "value", str)
			}

			return nil
		},
	})
	return client
}

var SDKModule = fx.Module("sdk",
	fx.Provide(ProvideSDK),
)
