package oops

import "context"

type contextKey string

const contextKeyOops = contextKey("oops")

func getBuilderFromContext(ctx context.Context) (OopsErrorBuilder, bool) {
	b, ok := ctx.Value(contextKeyOops).(OopsErrorBuilder)
	return b, ok
}

// WithBuilder set the error builder in the context, to be retrieved later with FromContext.
func WithBuilder(ctx context.Context, builder OopsErrorBuilder) context.Context {
	return context.WithValue(ctx, contextKeyOops, builder)
}
