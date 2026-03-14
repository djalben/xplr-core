package handler //nolint:testpackage

import (
	"bufio"
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/djalben/xplr-core/internal/infrastructure/logger"
)

func TestGroups(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx          context.Context
		loggerGroups []string
		groups       []string
	}

	tests := []struct {
		name string
		args args
		want map[string]struct{}
	}{
		{
			name: "dumb",
			args: args{
				ctx:          t.Context(),
				loggerGroups: []string{},
				groups:       []string{"one", "two"},
			},
			want: map[string]struct{}{"one": {}, "two": {}},
		},
		{
			name: "ContextWithGroups",
			args: args{
				ctx:          logger.ContextWithGroups(t.Context(), "one", "two"),
				loggerGroups: []string{},
				groups:       []string{"one", "two", "three", "four"},
			},
			want: map[string]struct{}{"one": {}, "two": {}, "three": {}, "four": {}},
		},
		{
			name: "LoggerWithGroups",
			args: args{
				ctx:          t.Context(),
				loggerGroups: []string{"one", "two"},
				groups:       []string{"one", "two", "three", "four"},
			},
			want: map[string]struct{}{"one": {}, "two": {}, "three": {}, "four": {}},
		},
		{
			name: "LoggerAndContextWithGroups",
			args: args{
				ctx:          logger.ContextWithGroups(t.Context(), "one", "two"),
				loggerGroups: []string{"two", "one", "three"},
				groups:       []string{"one", "two", "three", "four"},
			},
			want: map[string]struct{}{"one": {}, "two": {}, "three": {}, "four": {}},
		},
	}

	for _, tStruct := range tests {
		t.Run(tStruct.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			writerBuffer := bufio.NewWriter(buf)

			cHandler := &commonHandler{}

			for _, g := range tStruct.args.loggerGroups {
				cHandler = cHandler.withGroup(g)
			}

			ctx := logger.ContextWithGroups(tStruct.args.ctx, tStruct.args.groups...)
			cHandler.writeTag(ctx, writerBuffer)

			_ = writerBuffer.Flush()

			// ch.writeTag output: `"t":["one","two"],`
			str := buf.String()[6:]
			str = strings.TrimRight(str, "\"],")
			strArr := strings.Split(str, `","`)
			got := make(map[string]struct{})

			for _, s := range strArr {
				got[s] = struct{}{}
			}

			if !reflect.DeepEqual(got, tStruct.want) {
				t.Errorf("ContextWithGroups() = %v, want %v", got, tStruct.want)
			}
		})
	}
}

func BenchmarkLoggerGroups(b *testing.B) {
	cHandler := &commonHandler{}

	groups := []string{"one", "two", "three", "four", "five", "six"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := range b.N {
		cHandler.withGroup(groups[i%6])
	}
}

func BenchmarkCtxGroups(b *testing.B) {
	groups := []string{"one", "two", "three", "four", "five", "six"}

	b.ReportAllocs()
	b.ResetTimer()

	ctx := b.Context()

	for i := range b.N {
		ctx = logger.ContextWithGroups(ctx, groups[i%6]) //nolint:fatcontext
	}
}

func BenchmarkWriteTags(b *testing.B) {
	groups1 := []string{"one", "two", "three"}
	groups2 := []string{"two", "five", "six"}

	buf := &bytes.Buffer{}
	writerBuffer := bufio.NewWriter(buf)
	cHandler := &commonHandler{}

	for _, g := range groups1 {
		cHandler = cHandler.withGroup(g)
	}

	ctx := logger.ContextWithGroups(b.Context(), groups2...)

	b.ReportAllocs()
	b.ResetTimer()

	for range b.N {
		cHandler.writeTag(ctx, writerBuffer)
	}
}
