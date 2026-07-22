package middlewares

import (
	"context"
	"strings"

	"goph_keeper/internal/server/utils"
	"goph_keeper/internal/shared/models"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthUnaryInterceptor(secretKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if strings.HasSuffix(info.FullMethod, "Login") || strings.HasSuffix(info.FullMethod, "Register") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is missing")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is missing")
		}

		tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		userName, err := utils.ParseToken(tokenStr, secretKey)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		authenticatedCtx := models.WithUserName(ctx, userName)

		return handler(authenticatedCtx, req)
	}
}
