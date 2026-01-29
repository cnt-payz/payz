package interceptors

import (
	"context"

	"github.com/cnt-payz/payz/sso-service/internal/domain/session"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type PackInterceptors struct {
	jwtMngr     session.JWTManager
	authRequire map[string]bool
}

func New(jwtMngr session.JWTManager, authRequire map[string]bool) *PackInterceptors {
	return &PackInterceptors{
		jwtMngr:     jwtMngr,
		authRequire: authRequire,
	}
}

func (pi *PackInterceptors) BaseInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid metadata")
		}

		clientIP := md.Get("x-client-ip")
		if len(clientIP) < 1 {
			return nil, status.Error(codes.InvalidArgument, "client ip is required")
		}

		userAgent := md.Get("x-client-user-agent")
		if len(userAgent) < 1 {
			return nil, status.Error(codes.InvalidArgument, "user-agent is required")
		}

		ctx = context.WithValue(ctx, session.CtxKey("x-client-ip"), clientIP[0])
		ctx = context.WithValue(ctx, session.CtxKey("user-agent"), userAgent[0])

		return handler(ctx, req)
	}
}

func (pi *PackInterceptors) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !pi.authRequire[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid metadata")
		}

		token := md.Get("authorization")
		if len(token) < 1 {
			return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
		}

		claims, err := pi.jwtMngr.Validate(token[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidToken.Error())
		}

		ctx = context.WithValue(ctx, session.CtxKey("user-id"), claims.UserID)

		return handler(ctx, req)
	}
}
