package auth

import (
	"context"
)

type contextKey string

const userKey = contextKey("user_id")
const userIdKey = contextKey("id")

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIdKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userIdKey).(string)
	return uid, ok
}
