package auth

import (
	"context"
)

type contextKey string

const userKey = contextKey("user_id")

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userKey).(string)
	return uid, ok
}
