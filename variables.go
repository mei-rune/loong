package loong

import "context"

type userKey struct{}

var UserKey = &userKey{}

func ContextWithUser(ctx context.Context, u interface{}) context.Context {
	return context.WithValue(ctx, UserKey, u)
}

func UserFromContext(ctx context.Context) interface{} {
	return ctx.Value(UserKey)
}

type tokenKey struct{}

var TokenKey = &tokenKey{}

func ContextWithToken(ctx context.Context, u interface{}) context.Context {
	return context.WithValue(ctx, TokenKey, u)
}

func TokenFromContext(ctx context.Context) interface{} {
	return ctx.Value(TokenKey)
}
