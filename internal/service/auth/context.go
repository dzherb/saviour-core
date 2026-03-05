package auth

import "context"

type tokenContextKey struct{}

func ContextWithToken(
	ctx context.Context,
	token *TokenParsed,
) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

func TokenFromContext(ctx context.Context) (*TokenParsed, bool) {
	token, ok := ctx.Value(tokenContextKey{}).(*TokenParsed)
	return token, ok
}

func MustGetTokenFromContext(ctx context.Context) *TokenParsed {
	token, ok := TokenFromContext(ctx)
	if !ok {
		panic("auth token not found in context")
	}

	return token
}
