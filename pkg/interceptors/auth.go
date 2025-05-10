package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/AleksZelenchuk/vault-server/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// Extract and validate token
func authenticate(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata not provided")
	}

	authHeader := md["authorization"]
	if len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token missing")
	}

	token := strings.TrimPrefix(authHeader[0], "Bearer ")
	claims, err := auth.ValidateToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	uid, ok := claims["user_id"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user ID missing in token")
	}

	ctx = auth.WithUserID(ctx, uid)
	return ctx, nil
}

func UnaryAuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	fmt.Println(ctx)
	ctx, err := authenticate(ctx)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func StreamAuthInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx, err := authenticate(ss.Context())
	if err != nil {
		return err
	}
	wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
	return handler(srv, wrapped)
}
