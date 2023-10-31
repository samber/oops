package oops

import (
	"context"

	"github.com/samber/lo"
)

func coalesceOrEmpty[T comparable](v ...T) T {
	result, _ := lo.Coalesce(v...)
	return result
}

// convert (interface{})(nil) to nil
func contextValueOrNil(ctx context.Context, k any) any {
	v := ctx.Value(k)
	if v == nil {
		return nil
	}

	return v
}
