package logger_test

import (
	"context"
	"slices"
	"testing"

	"github.com/djalben/xplr-core/backend/internal/infrastructure/logger"
)

func TestContextWithGroups(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx    context.Context
		groups []string
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "",
			args: args{
				ctx:    t.Context(),
				groups: []string{"one", "two"},
			},
			want: []string{"one", "two"},
		},
		{
			name: "ContextWithGroups",
			args: args{
				ctx:    logger.ContextWithGroups(t.Context(), "one", "two"),
				groups: []string{"two", "three", "four"},
			},
			want: []string{"one", "two", "three", "four"},
		},
	}

	for _, tStruct := range tests {
		t.Run(tStruct.name, func(t *testing.T) {
			t.Parallel()

			ctx := logger.ContextWithGroups(tStruct.args.ctx, tStruct.args.groups...)
			got := logger.GetGroupsFromContext(ctx)

			if !slices.Equal(got, tStruct.want) {
				t.Errorf("ContextWithGroups() = %v, want %v", got, tStruct.want)
			}
		})
	}
}
