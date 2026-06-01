package cli

import "context"

func contextWithOptions(ctx context.Context, opts *options) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, optionsKey{}, opts)
}

func optionsFromContext(ctx context.Context) (*options, bool) {
	if ctx == nil {
		return nil, false
	}
	opts, ok := ctx.Value(optionsKey{}).(*options)
	return opts, ok
}
