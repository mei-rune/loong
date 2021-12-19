module github.com/runner-mei/loong

go 1.13

require (
	emperror.dev/emperror v0.33.0 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/go-kit/kit v0.9.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/labstack/echo/v4 v4.6.1
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/runner-mei/GoBatis v1.2.0 // indirect
	github.com/runner-mei/errors v0.0.0-20211208021129-603ff364d8b8
	github.com/runner-mei/log v1.0.8
	github.com/uber/jaeger-client-go v2.19.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/tools v0.1.8 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace github.com/labstack/echo/v4 v4.6.1 => github.com/runner-mei/echo/v4 v4.6.2-0.20211219084718-001961dd5394
