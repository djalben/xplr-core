package logger

import (
	"context"
	"slices"
)

type KeyContext string // KeyContext - ключ к контексту

const KeyContextGroups KeyContext = "loggerGroups" // KeyContext - ключ для групп контекста

// GetGroupsFromContext получение групп из контекста.
func GetGroupsFromContext(ctx context.Context) []string {
	contextGroups, _ := ctx.Value(KeyContextGroups).([]string)

	return contextGroups
}

// ContextWithGroups - добавление групп в контекст.
func ContextWithGroups(ctx context.Context, groups ...string) context.Context {
	contextGroups := slices.Clone(GetGroupsFromContext(ctx))

	for _, name := range groups {
		if !slices.Contains(contextGroups, name) {
			contextGroups = append(contextGroups, name)
		}
	}

	return context.WithValue(ctx, KeyContextGroups, contextGroups)
}
