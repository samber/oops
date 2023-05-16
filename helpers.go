package oops

import "github.com/samber/lo"

func AsOops(err error) (OopsError, bool) {
	return lo.ErrorsAs[OopsError](err)
}
