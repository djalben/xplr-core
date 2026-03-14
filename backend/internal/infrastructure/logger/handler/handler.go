package handler

import (
	"context"
	"io"
	"log/slog"
)

type jsonHandler struct {
	*commonHandler
}

// New - создание JSON хандлера.
// writer - io.Writer для записи логов.
// opts - опции для хандлера, если nil, то используются значения по умолчанию.
// opts.Level - уровень логирования, по умолчанию slog.LevelInfo.
// opts.AddSource - добавлять ли информацию о файле и строке вызова, по умолчанию false.
// opts.ReplaceAttr - функция для замены атрибутов, по умолчанию nil.
// opts.AddGroup - добавлять ли информацию о группах, по умолчанию false.
// opts.TimeFormat - формат времени, по умолчанию "2006-01-02T15:04:05.000Z07:00".
func New(writer io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &jsonHandler{
		commonHandler: &commonHandler{
			writer: writer,
			opts:   opts,
			groups: nil,
		},
	}
}

func (handler *jsonHandler) Enabled(_ context.Context, level slog.Level) bool {
	return handler.enabled(level)
}

func (handler *jsonHandler) Handle(ctx context.Context, record slog.Record) error {
	return handler.handle(ctx, record)
}

func (handler *jsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler.copy(handler.withAttrs(attrs))
}

func (handler *jsonHandler) WithGroup(name string) slog.Handler {
	return handler.copy(handler.withGroup(name))
}

func (handler *jsonHandler) copy(comHandler *commonHandler) slog.Handler {
	return &jsonHandler{
		commonHandler: comHandler,
	}
}
