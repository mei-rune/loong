package loong

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Tracing(tracer trace.Tracer, comp string, traceAll bool) MiddlewareFunc {
	if tracer == nil {
		tracer = otel.Tracer(comp)
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			var req = c.Request()

			// 监测Header中是否有Trace信息
			ctx, span := tracer.Start(
				c.StdContext,
				comp+":"+req.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("component", comp),
					attribute.String("http.url", c.Request().Host+c.Request().RequestURI),
					attribute.String("http.method", c.Request().Method),
				),
			)
			defer span.End()

			c.StdContext = ctx

			err := next(c)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "")
			}

			span.SetAttributes(attribute.Int("http.status_code", c.Response().Status))
			return err
		}
	}
}

func RawTracing(tracer trace.Tracer, comp string, traceAll bool) func(ContextHandlerFunc) ContextHandlerFunc {
	if tracer == nil {
		tracer = otel.Tracer(comp)
	}

	return func(next ContextHandlerFunc) ContextHandlerFunc {
		hfn := func(ctx context.Context, w http.ResponseWriter, req *http.Request) {

			ctx, span := tracer.Start(
				ctx,
				comp+":"+req.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("component", comp),
					attribute.String("http.url", req.Host+req.RequestURI),
					attribute.String("http.method", req.Method),
				),
			)
			defer span.End()

			resp := responseOTel{span, w}
			next(ctx, resp, req)
		}
		return ContextHandlerFunc(hfn)
	}
}

type responseOTel struct {
	span trace.Span
	http.ResponseWriter
}

func (r responseOTel) WriteHeader(code int) {
	if code >= 400 {
		r.span.SetStatus(codes.Error, "")
	} else {
		r.span.SetStatus(codes.Ok, "")
	}

	r.span.SetAttributes(attribute.Int("http.status_code", code))
	r.ResponseWriter.WriteHeader(code)
}

