package token

import (
	"context"

	jwt "github.com/dgrijalva/jwt-go"
)

type Token jwt.Token

const (
	// ContextTokenKey defines key for storing token in the Context.
	ContextTokenKey = iota
)

func ContextWithToken(ctx context.Context, token *Token) context.Context {
	return context.WithValue(ctx, ContextTokenKey, token)
}

func FromContext(ctx context.Context) *Token {
	ti := ctx.Value(ContextTokenKey)
	if t, ok := ti.(*Token); ok {
		return t
	}
	return nil
}
