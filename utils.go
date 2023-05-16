package oops

import (
	"github.com/samber/lo"
)

func coalesceOrEmpty[T comparable](v ...T) T {
	result, _ := lo.Coalesce(v...)
	return result
}
