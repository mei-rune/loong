package loong

import (
	"context"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/runner-mei/log"
)

func Tracing(comp string, traceAll bool) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			var req = c.Request()
			var span opentracing.Span

			// 监测Header中是否有Trace信息
			wireContext, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Request().Header))
			if err != nil {
				if !traceAll {
					if isDebug := c.QueryParam("opentracing"); isDebug != "true" {
						return next(c)
					}
				}
				span = opentracing.StartSpan(comp + ":" + req.URL.Path)
			} else {
				span = opentracing.StartSpan(comp+":"+req.URL.Path, opentracing.ChildOf(wireContext))
			}
			defer span.Finish()

			c.StdContext = opentracing.ContextWithSpan(c.StdContext, span)

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

func RawTracing(comp string, traceAll bool) func(ContextHandlerFunc) ContextHandlerFunc {
	return func(next ContextHandlerFunc) ContextHandlerFunc {
		hfn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
			var span opentracing.Span

			// 监测Header中是否有Trace信息
			wireContext, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(req.Header))
			if err != nil {
				if !traceAll {
					if isDebug := req.URL.Query().Get("opentracing"); isDebug != "true" {
						next(ctx, w, req)
						return
					}
				}
				span = opentracing.StartSpan(comp + ":" + req.URL.Path)
			} else {
				span = opentracing.StartSpan(comp+":"+req.URL.Path, opentracing.ChildOf(wireContext))
			}
			defer span.Finish()

			ctx = opentracing.ContextWithSpan(ctx, span)

			ext.Component.Set(span, comp)
			ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
			ext.HTTPUrl.Set(span, req.Host+req.RequestURI)
			ext.HTTPMethod.Set(span, req.Method)

			resp := response{span, w}
			next(ctx, resp, req)
		}
		return ContextHandlerFunc(hfn)
	}
}

type response struct {
	span opentracing.Span
	http.ResponseWriter
}

func (r response) WriteHeader(code int) {
	if code >= 200 && code < 300 {
		ext.Error.Set(r.span, true)
	} else {
		ext.Error.Set(r.span, false)
	}

	ext.HTTPStatusCode.Set(r.span, uint16(code))
	r.ResponseWriter.WriteHeader(code)
}
