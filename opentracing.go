package loong

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/runner-mei/log"
)

func Tracing(comp string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			var req = c.Request()
			var span opentracing.Span

			// 监测Header中是否有Trace信息
			wireContext, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Request().Header))
			if err != nil {
				if isDebug := c.QueryParam("opentracing"); isDebug != "true" {
					return next(c)
				}

				span = opentracing.StartSpan(comp + ":" + req.URL.Path)
			} else {
				span = opentracing.StartSpan(comp+":"+req.URL.Path, opentracing.ChildOf(wireContext))
			}
			defer span.Finish()

			ext.Component.Set(span, comp)
			ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
			ext.HTTPUrl.Set(span, c.Request().Host+c.Request().RequestURI)
			ext.HTTPMethod.Set(span, c.Request().Method)

			if c.CtxLogger != nil {
				c.CtxLogger = c.CtxLogger.WithTargets(log.OutputToTracer(log.DefaultSpanLevel, span))
				c.StdContext = log.ContextWithLogger(c.StdContext, c.CtxLogger)
			}

			err = next(c)
			if err != nil {
				ext.Error.Set(span, true)
			} else {
				ext.Error.Set(span, false)
			}

			ext.HTTPStatusCode.Set(span, uint16(c.Response().Status))
			return err
		}
	}
}
