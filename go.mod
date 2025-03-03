module github.com/runner-mei/loong

go 1.20

require (
	gitee.com/Trisia/gotlcp v1.3.21
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/spec v0.20.9 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1-0.20230219130118-4fd5621d8dd0
	github.com/golang/mock v1.6.0 // indirect
	github.com/jszwec/csvutil v1.10.0 // indirect
	github.com/labstack/echo/v4 v4.11.2-0.20230919052447-4bc3e475e313
	github.com/mei-rune/csvutil v0.0.0-20221230090625-d3b9c650225d
	github.com/mei-rune/ipfilter v1.0.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/runner-mei/errors v0.0.0-20220725054952-d7c9c10762ea
	github.com/runner-mei/log v1.0.11
	github.com/swaggo/echo-swagger v1.4.0
	github.com/swaggo/swag v1.16.1 // indirect
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.25.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

exclude github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2

replace golang.org/x/exp => github.com/mei-rune/golang_exp_for_go120 v0.0.0-20250303053821-1e7433e4f2f2
