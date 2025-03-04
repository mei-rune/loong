module github.com/runner-mei/loong

go 1.20

require (
	gitee.com/Trisia/gotlcp v1.3.21
	github.com/golang-jwt/jwt/v4 v4.5.1-0.20230219130118-4fd5621d8dd0
	github.com/labstack/echo/v4 v4.11.2-0.20230919052447-4bc3e475e313
	github.com/mei-rune/csvutil v0.0.0-20221230090625-d3b9c650225d
	github.com/mei-rune/ipfilter v1.0.2
	github.com/opentracing/opentracing-go v1.2.0
	github.com/runner-mei/errors v0.0.0-20220725054952-d7c9c10762ea
	github.com/runner-mei/log v1.0.11
	github.com/swaggo/echo-swagger v1.4.0
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible
	go.uber.org/zap v1.25.0
)

require (
	emperror.dev/emperror v0.33.0 // indirect
	emperror.dev/errors v0.8.1 // indirect
	github.com/GeertJohan/go.rice v1.0.3 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/daaku/go.zipexe v1.0.2 // indirect
	github.com/emmansun/gmsm v0.27.2 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/spec v0.20.9 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jszwec/csvutil v1.10.0 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/swaggo/files/v2 v2.0.2 // indirect
	github.com/swaggo/swag v1.16.1 // indirect
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20191030013958-a1ab85dbe136 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/swaggo/echo-swagger v1.4.0 => github.com/mei-rune/echo-swagger v0.0.0-20250304024037-fe7f6354b014

replace github.com/swaggo/files/v2 v2.0.2 => github.com/mei-rune/swaggofiles/v2 v2.0.0-20240321041418-dd385d891b92

replace github.com/swaggo/swag v1.16.1 => github.com/runner-mei/swag v1.8.2-0.20231226075722-f02eee2df576

replace golang.org/x/exp => github.com/mei-rune/golang_exp_for_go120 v0.0.0-20250303053821-1e7433e4f2f2
