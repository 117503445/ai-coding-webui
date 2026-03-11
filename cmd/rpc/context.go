package main

import "context"

type ctxKey struct{}

type AppContext struct {
	RequestID string
}

func WithContext(ctx context.Context, appContext AppContext) context.Context {
	return context.WithValue(ctx, ctxKey{}, appContext)
}

func GetAppContext(ctx context.Context) AppContext {
	v := ctx.Value(ctxKey{})
	if v == nil {
		return AppContext{}
	}
	return v.(AppContext)
}
