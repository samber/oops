package oops

func Wrap2[A any](a A, err error) (A, error) {
	return a, Wrap(err)
}

func Wrap3[A any, B any](a A, b B, err error) (A, B, error) {
	return a, b, Wrap(err)
}

func Wrap4[A any, B any, C any](a A, b B, c C, err error) (A, B, C, error) {
	return a, b, c, Wrap(err)
}

func Wrap5[A any, B any, C any, D any](a A, b B, c C, d D, err error) (A, B, C, D, error) {
	return a, b, c, d, Wrap(err)
}

func Wrap6[A any, B any, C any, D any, E any](a A, b B, c C, d D, e E, err error) (A, B, C, D, E, error) {
	return a, b, c, d, e, Wrap(err)
}

func Wrap7[A any, B any, C any, D any, E any, F any](a A, b B, c C, d D, e E, f F, err error) (A, B, C, D, E, F, error) {
	return a, b, c, d, e, f, Wrap(err)
}

func Wrap8[A any, B any, C any, D any, E any, F any, G any](a A, b B, c C, d D, e E, f F, g G, err error) (A, B, C, D, E, F, G, error) {
	return a, b, c, d, e, f, g, Wrap(err)
}

func Wrap9[A any, B any, C any, D any, E any, F any, G any, H any](a A, b B, c C, d D, e E, f F, g G, h H, err error) (A, B, C, D, E, F, G, H, error) {
	return a, b, c, d, e, f, g, h, Wrap(err)
}

func Wrap10[A any, B any, C any, D any, E any, F any, G any, H any, I any](a A, b B, c C, d D, e E, f F, g G, h H, i I, err error) (A, B, C, D, E, F, G, H, I, error) {
	return a, b, c, d, e, f, g, h, i, Wrap(err)
}
