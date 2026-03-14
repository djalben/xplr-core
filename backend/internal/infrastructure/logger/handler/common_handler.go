package handler

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/djalben/xplr-core/internal/infrastructure/logger"
)

type commonHandler struct {
	writer io.Writer
	opts   *slog.HandlerOptions
	groups []string
	attrs  []slog.Attr
}

func (ch *commonHandler) clone() *commonHandler {
	// We can't use assignment because we can't copy the mutex.
	return &commonHandler{
		writer: ch.writer,
		opts:   ch.opts,
		groups: slices.Clip(ch.groups),
		attrs:  slices.Clip(ch.attrs),
	}
}

func (ch *commonHandler) enabled(lvl slog.Level) bool {
	if ch.opts.Level != nil {
		return lvl >= ch.opts.Level.Level()
	}

	return lvl >= logger.Log
}

func (ch *commonHandler) withAttrs(attrs []slog.Attr) *commonHandler {
	h2 := ch.clone()
	h2.attrs = append(h2.attrs, attrs...)

	return h2
}

func (ch *commonHandler) withGroup(name string) *commonHandler {
	ch2 := ch.clone()

	if !slices.Contains(ch2.groups, name) {
		ch2.groups = append(ch2.groups, name)
	}

	return ch2
}

func (ch *commonHandler) handle(ctx context.Context, record slog.Record) error {
	writer := bufio.NewWriter(ch.writer)

	defer func() {
		err := writer.Flush()
		if err != nil {
			return
		}
	}()

	// Открываем JSON
	err := writer.WriteByte(123)
	if err != nil {
		return fmt.Errorf("writer.WriteByte(123) error=%w", err)
	}

	ch.writeLevel(writer, record.Level)

	// Запись поля time
	ch.writeTime(writer, record.Time.Local())

	// Запись поля t
	ch.writeTag(ctx, writer)

	// Запись поля msg
	ch.writeMSG(writer, record.Message)

	ch.writeD(writer, record)

	// Закрываем JSON
	_ = writer.WriteByte(125)
	// Переход на новую строку
	_ = writer.WriteByte(10)

	return nil
}

func (ch *commonHandler) writeMSG(builder *bufio.Writer, msg string) {
	_, _ = builder.WriteString("\"msg\":\"")
	_, _ = builder.WriteString(msg)
	_, _ = builder.WriteString("\",")
}

func (ch *commonHandler) writeTime(builder *bufio.Writer, time time.Time) {
	_, _ = builder.WriteString("\"time\":\"")
	_, _ = builder.WriteString(time.Format("2006-01-02 15:04:05.000"))
	_, _ = builder.WriteString("\",")
}

func (ch *commonHandler) writeLevel(builder *bufio.Writer, lvl slog.Level) {
	_, _ = builder.WriteString("\"l\":\"")
	_, _ = builder.WriteString(logger.ToString(lvl))
	_, _ = builder.WriteString("\",")
}

func (ch *commonHandler) writeTag(ctx context.Context, builder *bufio.Writer) {
	// Открываем теги
	_, _ = builder.WriteString("\"t\":[")

	for index, group := range ch.groups {
		if index != 0 {
			_ = builder.WriteByte(44)
		}

		_ = builder.WriteByte(34)
		_, _ = builder.WriteString(group)
		_ = builder.WriteByte(34)
	}

	for index, group := range logger.GetGroupsFromContext(ctx) {
		if !slices.Contains(ch.groups, group) {
			if len(ch.groups)+index != 0 {
				_ = builder.WriteByte(44)
			}

			_ = builder.WriteByte(34)
			_, _ = builder.WriteString(group)
			_ = builder.WriteByte(34)
		}
	}

	// закрываем теги
	_, _ = builder.WriteString("],")
}

func (ch *commonHandler) writeD(builder *bufio.Writer, record slog.Record) {
	// запускаем тег D
	_, _ = builder.WriteString("\"d\":")

	ch.writePC(builder, record.PC)

	ch.writeAttr(builder, record)

	// Закрываем тег D
	_ = builder.WriteByte(125)
}

func (ch *commonHandler) writePC(builder *bufio.Writer, pointer uintptr) {
	// Открываем caller
	_, _ = builder.WriteString("{\"$caller\":\"")

	funcForPC := runtime.FuncForPC(pointer)

	file, line := funcForPC.FileLine(pointer)

	_, fileName := path.Split(file)

	funcNameForPC := funcForPC.Name()

	ix := strings.LastIndex(funcNameForPC, ".")

	_, _ = builder.WriteString(funcNameForPC[0:ix] + "/" + fileName + ":" + strconv.Itoa(line))

	// Закрываем caller
	_ = builder.WriteByte(34)
}

func (ch *commonHandler) writeAttr(builder *bufio.Writer, record slog.Record) {
	record.Attrs(func(attr slog.Attr) bool {
		_ = builder.WriteByte(44)
		_ = builder.WriteByte(34)
		_, _ = builder.WriteString(attr.Key)
		_, _ = builder.WriteString("\":")

		switch attr.Value.Kind() {
		case
			slog.KindBool,
			slog.KindFloat64,
			slog.KindInt64,
			slog.KindUint64:
			_, _ = builder.WriteString(attr.Value.String())

		case
			slog.KindAny,
			slog.KindDuration,
			slog.KindString,
			slog.KindTime,
			slog.KindGroup,
			slog.KindLogValuer:
			_, _ = builder.WriteString(strconv.Quote(attr.Value.String()))
		}

		return true
	})
}
