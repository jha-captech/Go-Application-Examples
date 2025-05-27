package ctxhandler

import (
	"context"
	"log/slog"
)

type (
	Option   func(options *handlerOptions)
	AttrFunc func(context.Context) slog.Attr
)

type handlerOptions struct {
	slogAttrFuncs []AttrFunc
}

type Handler struct {
	slog.Handler
	options handlerOptions
}

func WithAtterFunc(f AttrFunc) Option {
	return func(options *handlerOptions) {
		options.slogAttrFuncs = append(options.slogAttrFuncs, f)
	}
}

func (c *Handler) Handle(ctx context.Context, record slog.Record) error {
	for _, f := range c.options.slogAttrFuncs {
		record.AddAttrs(f(ctx))
	}

	return c.Handler.Handle(ctx, record)
}

func WrapSlogHandler(handler slog.Handler, options ...Option) *Handler {
	opts := &handlerOptions{
		slogAttrFuncs: make([]AttrFunc, 0),
	}

	for _, f := range options {
		f(opts)
	}

	return &Handler{
		Handler: handler,
		options: *opts,
	}
}
