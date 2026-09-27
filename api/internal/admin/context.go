package admin

import (
	"context"

	"github.com/wolfandwings/api/internal/model"
)

type ctxKey string

const userKey ctxKey = "user"

func withUser(ctx context.Context, u *model.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func userFromCtx(ctx context.Context) *model.User {
	u, _ := ctx.Value(userKey).(*model.User)
	return u
}
