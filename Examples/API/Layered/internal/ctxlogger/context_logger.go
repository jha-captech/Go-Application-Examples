package ctxlogger

import (
	"context"
	"log/slog"
)

type (
	Option   func(options *ctxHandlerOptions)
	AttrFunc func(context.Context) slog.Attr
)

type ctxHandlerOptions struct {
	slogAttrFuncs []AttrFunc
}

type CtxHandler struct {
	slog.Handler
	options ctxHandlerOptions
}

func WithAtterFunc(f AttrFunc) Option {
	return func(options *ctxHandlerOptions) {
		options.slogAttrFuncs = append(options.slogAttrFuncs, f)
	}
}

func (c *CtxHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, f := range c.options.slogAttrFuncs {
		record.AddAttrs(f(ctx))
	}

	return c.Handler.Handle(ctx, record)
}

func WrapSlogHandler(handler slog.Handler, options ...Option) *CtxHandler {
	opts := &ctxHandlerOptions{
		slogAttrFuncs: make([]AttrFunc, 0),
	}

	for _, f := range options {
		f(opts)
	}

	return &CtxHandler{
		Handler: handler,
		options: *opts,
	}
}
