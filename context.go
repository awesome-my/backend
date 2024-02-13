package awesomemy

import (
	"context"
	"fmt"
)

type ctxKey struct {
	name string
}

var (
	CtxKeyLogger = ctxKey{"awesomemy.logger"}
	CtxKeyConfig = ctxKey{"awesomemy.config"}
)

func MustContextValue[T any](ctx context.Context, key ctxKey) T {
	v, ok := ctx.Value(key).(T)
	if !ok {
		panic(fmt.Sprintf("turtle: could not assert context (%v) value", key))
	}

	return v
}