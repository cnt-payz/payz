package interceptors

import (
	"context"

	"github.com/cnt-payz/payz/payment-service/internal/domain/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type InterceptorPack struct {
	jwtMngr            session.JWTManager
	authRequire        map[string]bool
	idempotencyRequire map[string]bool
}

func New(
	jwtMngr session.JWTManager,
	authRequire, idempotencyRequire map[string]bool,
) *InterceptorPack {
	return &InterceptorPack{
		jwtMngr:            jwtMngr,
		authRequire:        authRequire,
		idempotencyRequire: idempotencyRequire,
	}
}

func (ip *InterceptorPack) IdempotencyInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !ip.idempotencyRequire[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "metadata is required")
		}

		idempotencyKey := md.Get("idempotency-key")
		if len(idempotencyKey) < 1 {
			return nil, status.Error(codes.InvalidArgument, "idempotency key is required")
		}

		ctx = context.WithValue(ctx, session.CtxKey("idempotency-key"), idempotencyKey[0])

		return handler(ctx, req)
	}
}

func (ip *InterceptorPack) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !ip.authRequire[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "metadata is required")
		}

		token := md.Get("authorization")
		if len(token) < 1 {
			return nil, status.Error(codes.Unauthenticated, "token is required")
		}

		claims, err := ip.jwtMngr.Validate(token[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		ctx = context.WithValue(ctx, session.CtxKey("user-id"), claims.UserID)
		ctx = context.WithValue(ctx, session.CtxKey("user-email"), claims.Email)

		return handler(ctx, req)
	}
}
