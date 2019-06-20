package jaeger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
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

func init() {
	if name, _ := osext.Executable(); name != "" {
		name = filepath.Base(name)
		if name != "" {
			serviceName = name
		}
	}

	if name := os.Getenv("LOONG_TRACING_NAME"); name != "" {
		serviceName = name
	}

	cfg, err := config.FromEnv()
	if err != nil {
		panic(err)
	}

	cfg.Sampler.Type = jaeger_client.SamplerTypeConst
	cfg.Sampler.Param = 1

	metricFactory = expvar.NewFactory(10)
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		config.Metrics(metricFactory),
	)
	if err != nil {
		panic(err)
	}

	gCloser = closer
}

func Init(name string, logger *zap.Logger) error {
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
