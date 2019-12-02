package jaeger

import (
	"io"
	"os"

	jaeger_client "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	jaeger_zap "github.com/uber/jaeger-client-go/log/zap"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/expvar"
	"go.uber.org/zap"
)

var (
	serviceName   = "loong"
	gCloser       io.Closer
	metricFactory metrics.Factory
)

func Init(name string, logger *zap.Logger) error {
	if value := os.Getenv("OPENTRACE_ENABLED"); value == "false" || value == "disabled" {
		return nil
	}

	if name == "" {
		name = serviceName
	}
	cfg, err := config.FromEnv()
	if err != nil {
		panic(err)
	}

	cfg.Sampler.Type = jaeger_client.SamplerTypeConst
	cfg.Sampler.Param = 1

	gCloser.Close()

	if metricFactory == nil {
		metricFactory = expvar.NewFactory(10)
	}

	var opts = make([]config.Option, 0, 2)
	opts = append(opts, config.Metrics(metricFactory))
	if logger != nil {
		opts = append(opts, config.Logger(jaeger_zap.NewLogger(logger)))
	}

	closer, err := cfg.InitGlobalTracer(
		name,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	gCloser = closer
	return nil
}
