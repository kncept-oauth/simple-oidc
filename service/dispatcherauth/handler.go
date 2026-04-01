package dispatcherauth

import (
	"context"
	"fmt"

	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

type BearerAuthContextKey struct{}
type LoginCookieContextKey struct{}
type AnyAuthContextKey struct{}

type Handler struct {
}

func (obj *Handler) HandleBearerAuth(ctx context.Context, operationName api.OperationName, t api.BearerAuth) (context.Context, error) {
	if t.Token != "" {
		ctx = context.WithValue(ctx, AnyAuthContextKey{}, t.Token)
		ctx = context.WithValue(ctx, BearerAuthContextKey{}, t.Token)
	}
	return ctx, nil
}
func (obj *Handler) HandleLoginCookie(ctx context.Context, operationName api.OperationName, t api.LoginCookie) (context.Context, error) {
	fmt.Printf("HandleLoginCookie %v\n", t.APIKey)
	if t.APIKey != "" {
		ctx = context.WithValue(ctx, AnyAuthContextKey{}, t.APIKey)
		ctx = context.WithValue(ctx, LoginCookieContextKey{}, t.APIKey)
	}
	return ctx, nil
}

func GetBearerAuth(ctx context.Context) string {
	return valueAsString(ctx, BearerAuthContextKey{})
}

func GetLoginCookie(ctx context.Context) string {
	return valueAsString(ctx, LoginCookieContextKey{})
}

func GetAnyAuth(ctx context.Context) string {
	return valueAsString(ctx, AnyAuthContextKey{})
}

func valueAsString(ctx context.Context, key any) string {
	val := ctx.Value(key)
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}
