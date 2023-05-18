package oops

import "github.com/samber/lo"

// AsOops checks if an error is an `oops.OopsError` object.
// Alias to errors.As.
func AsOops(err error) (OopsError, bool) {
	return lo.ErrorsAs[OopsError](err)
}
