// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// Factory is the default logging wrapper that can create
// logger instances either for a given Context or context-less.
type Factory struct {
	targets []Target
	logger  Logger
}

// NewFactory creates a new Factory.
func NewFactory(logger *zap.Logger) Factory {
	return Factory{logger: NewLogger(logger)}
}

// FactoryFrom creates a new Factory.
func FactoryFrom(logger Logger) Factory {
	return Factory{logger: logger}
}

// Bg creates a context-unaware logger.
func (b Factory) New() Logger {
	return b.logger.WithTargets(b.targets...)
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func (b Factory) For(ctx context.Context) Factory {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return b.Span(span)
	}
	return b
}

// Span returns a span Logger, all logging calls are also
// echo-ed into the span.
func (b Factory) Span(span opentracing.Span) Factory {
	if span != nil {
		return Factory{logger: b.logger, targets: append(b.targets, OutputToTracer(DefaultSpanLevel, span))}
	}
	return b
}

// Span returns a span Logger, all logging calls are also
// echo-ed into the span.
func (b Factory) OutputToStrings(target *[]string) Factory {
	return Factory{logger: b.logger, targets: append(b.targets, OutputToStrings(InfoLevel, target))}
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b Factory) With(keyAndValues ...interface{}) Factory {
	return Factory{logger: b.logger.With(keyAndValues...), targets: b.targets}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (b Factory) Named(name string) Factory {
	return Factory{logger: b.logger.Named(name), targets: b.targets}
}

type contextKey struct{}

var activeFactoryKey = contextKey{}

// ContextWithFactory returns a new `context.Context` that holds a reference to
// `Factory`'s FactoryContext.
func ContextWithFactory(ctx context.Context, factory *Factory) context.Context {
	return context.WithValue(ctx, activeFactoryKey, factory)
}

// FactoryFromContext returns the `Factory` previously associated with `ctx`, or
// `nil` if no such `Factory` could be found.
//
// NOTE: context.Context != SpanContext: the former is Go's intra-process
// context propagation mechanism, and the latter houses OpenTracing's per-Factory
// identity and baggage information.
func FactoryFromContext(ctx context.Context) *Factory {
	val := ctx.Value(activeFactoryKey)
	if sp, ok := val.(*Factory); ok {
		return sp
	}
	return nil
}

func Span(logger Logger, span opentracing.Span, enabledLevel ...Level) Logger {
	if span == nil {
		return logger
	}

	if len(enabledLevel) > 0 {
		return logger.WithTargets(OutputToTracer(enabledLevel[0], span))
	}

	return logger.WithTargets(OutputToTracer(DefaultSpanLevel, span))
}

func SpanContext(logger Logger, spanContext opentracing.SpanContext, method string, enabledLevel ...Level) (Logger, func()) {
	if spanContext == nil {
		return logger, func() {}
	}

	span := opentracing.StartSpan(method, opentracing.ChildOf(spanContext))
	finish := func() {
		span.Finish()
	}

	if len(enabledLevel) > 0 {
		return logger.WithTargets(OutputToTracer(enabledLevel[0], span)), finish
	}
	return logger.WithTargets(OutputToTracer(DefaultSpanLevel, span)), finish
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func For(ctx context.Context, args ...interface{}) (Logger, func()) {
	var logger Logger
	var span opentracing.Span
	var spanContext opentracing.SpanContext
	var method string
	var level = DefaultSpanLevel
	var fields []interface{}

	for _, arg := range args {
		switch value := arg.(type) {
		case Logger:
			logger = value
		case opentracing.Span:
			span = value
		case opentracing.SpanContext:
			spanContext = value
		case string:
			method = value
		case Level:
			level = value
		case Field:
			fields = append(fields, value)
		}
	}

	if logger == nil {
		logger = LoggerOrEmptyFromContext(ctx)
	}
	if len(fields) > 0 {
		logger = logger.With(fields...)
	}

	if span != nil {
		return Span(logger, span, level), noop
	}

	if spanContext != nil {
		return SpanContext(logger, spanContext, method, level)
	}

	if span := opentracing.SpanFromContext(ctx); span != nil {
		return Span(logger, span, level), noop
	}
	return logger, noop
}

var noop = func() {}

func IsEmpty(logger Logger) bool {
	return logger == Empty
}
