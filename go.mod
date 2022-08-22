module github.com/runner-mei/loong

go 1.13

require (
	emperror.dev/emperror v0.33.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang/mock v1.6.0 // indirect
	github.com/labstack/echo/v4 v4.6.1
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/runner-mei/errors v0.0.0-20220212075214-841cd0837bdc
	github.com/runner-mei/log v1.0.8
	github.com/swaggo/echo-swagger v1.3.0
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace github.com/labstack/echo/v4 v4.6.1 => github.com/runner-mei/echo/v4 v4.6.2-0.20211219084718-001961dd5394
