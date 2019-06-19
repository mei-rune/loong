package jaeger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/expvar"
)

var (
	serviceName = "loong"
	gTracer     opentracing.Tracer
	gCloser     io.Closer
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

	metricsFactory := expvar.NewFactory(10)
	tracer, closer, err := config.Configuration{
		ServiceName: serviceName,
	}.NewTracer(
		config.Metrics(metricsFactory),
	)
	if err != nil {
		panic(err)
	}

	opentracing.SetGlobalTracer(tracer)

	gTracer = tracer
	gCloser = closer
}

func Init(name string) error {
	if name == "" {
		name = serviceName
	}

	metricsFactory := expvar.NewFactory(10)
	tracer, closer, err := config.Configuration{
		ServiceName: name,
	}.NewTracer(
		config.Metrics(metricsFactory),
	)
	if err != nil {
		return err
	}

	gCloser.Close()
	opentracing.SetGlobalTracer(tracer)
	gTracer = tracer
	gCloser = closer
	return nil
}
