package services

import (
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	ssopb "github.com/cnt-payz/payz/sso-gateway/api/sso/v1"
	"github.com/cnt-payz/payz/sso-gateway/internal/application/dtos"
	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-gateway/pkg/consts"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SSOGatewayService interface {
	Login(ctx *gin.Context, req *dtos.LoginRequest) (*dtos.Token, int, error)
	Register(ctx *gin.Context, req *dtos.RegisterRequest) (*dtos.Token, int, error)
	Refresh(ctx *gin.Context, req *dtos.RefreshRequest) (*dtos.Token, int, error)
	LogoutAll(ctx *gin.Context) (int, error)
	Delete(ctx *gin.Context) (int, error)
	GetUserByID(ctx *gin.Context, id string) (*dtos.User, int, error)
	GetUserByEmail(ctx *gin.Context, email string) (*dtos.User, int, error)
	GetSelfUser(ctx *gin.Context) (*dtos.User, int, error)
}

type ssoGatewayService struct {
	cfg       *config.Config
	log       *slog.Logger
	ssoClient ssopb.SSOClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	ssoClient ssopb.SSOClient,
) SSOGatewayService {
	return &ssoGatewayService{
		cfg:       cfg,
		log:       log,
		ssoClient: ssoClient,
	}
}

func (sgs *ssoGatewayService) Login(ctx *gin.Context, req *dtos.LoginRequest) (*dtos.Token, int, error) {
	req.Email = strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, http.StatusBadRequest, consts.ErrInvalidEmail
	}

	req.Password = strings.TrimSpace(req.Password)

	sgs.log.Debug("start to login user")
	md := sgs.baseMD(ctx)
	resp, err := sgs.ssoClient.Login(metadata.NewOutgoingContext(ctx, md), &ssopb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to login user",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return nil, http.StatusInternalServerError, consts.ErrInternalServer
		}

		return nil, code, errors.New(st.Message())
	}

	return &dtos.Token{
		Access:           resp.Access,
		Refresh:          resp.Refresh,
		AccessExpiresAt:  resp.AccessExpiresAt.AsTime().UnixMilli(),
		RefreshExpiresAt: resp.RefreshExpiresAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (sgs *ssoGatewayService) Register(ctx *gin.Context, req *dtos.RegisterRequest) (*dtos.Token, int, error) {
	req.Email = strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, http.StatusBadRequest, consts.ErrInvalidEmail
	}

	req.Password = strings.TrimSpace(req.Password)

	sgs.log.Debug("start to register user")
	md := sgs.baseMD(ctx)
	resp, err := sgs.ssoClient.Register(metadata.NewOutgoingContext(ctx, md), &ssopb.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to register user",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return nil, http.StatusInternalServerError, consts.ErrInternalServer
		}

		return nil, code, errors.New(st.Message())
	}

	return &dtos.Token{
		Access:           resp.Access,
		Refresh:          resp.Refresh,
		AccessExpiresAt:  resp.AccessExpiresAt.AsTime().UnixMilli(),
		RefreshExpiresAt: resp.RefreshExpiresAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (sgs *ssoGatewayService) Refresh(ctx *gin.Context, req *dtos.RefreshRequest) (*dtos.Token, int, error) {
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		return nil, http.StatusBadRequest, consts.ErrInvalidRefreshToken
	}

	sgs.log.Debug("start to refresh token")
	md := sgs.baseMD(ctx)
	resp, err := sgs.ssoClient.Refresh(metadata.NewOutgoingContext(ctx, md), &ssopb.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to refresh session",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return nil, http.StatusInternalServerError, consts.ErrInternalServer
		}

		return nil, code, errors.New(st.Message())
	}

	return &dtos.Token{
		Access:           resp.Access,
		Refresh:          resp.Refresh,
		AccessExpiresAt:  resp.AccessExpiresAt.AsTime().UnixMilli(),
		RefreshExpiresAt: resp.RefreshExpiresAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (sgs *ssoGatewayService) LogoutAll(ctx *gin.Context) (int, error) {
	sgs.log.Debug("start to logout all sessions")
	md, err := sgs.tokenMD(ctx)
	if err != nil {
		return http.StatusUnauthorized, err
	}

	if _, err := sgs.ssoClient.LogoutAll(metadata.NewOutgoingContext(ctx, md), nil); err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to logout all sessions",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return http.StatusInternalServerError, consts.ErrInternalServer
		}

		return code, errors.New(st.Message())
	}

	return http.StatusOK, nil
}

func (sgs *ssoGatewayService) Delete(ctx *gin.Context) (int, error) {
	sgs.log.Debug("start to delete user")
	md, err := sgs.tokenMD(ctx)
	if err != nil {
		return http.StatusUnauthorized, err
	}
	if _, err := sgs.ssoClient.Delete(metadata.NewOutgoingContext(ctx, md), nil); err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to delete user",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return http.StatusInternalServerError, consts.ErrInternalServer
		}

		return code, errors.New(st.Message())
	}

	return http.StatusOK, nil
}

func (sgs *ssoGatewayService) GetUserByID(ctx *gin.Context, id string) (*dtos.User, int, error) {
	sgs.log.Debug("getting user by id", slog.String("user_id", id))
	md := sgs.baseMD(ctx)
	resp, err := sgs.ssoClient.GetUserByID(metadata.NewOutgoingContext(ctx, md), &ssopb.GetByIDRequest{
		UserId: id,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to get user by id",
				slog.String("user_id", id),
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return nil, http.StatusInternalServerError, consts.ErrInternalServer
		}

		return nil, code, errors.New(st.Message())
	}

	return &dtos.User{
		ID:        resp.Id,
		Email:     resp.Email,
		CreatedAt: resp.CreatedAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (sgs *ssoGatewayService) GetUserByEmail(ctx *gin.Context, email string) (*dtos.User, int, error) {
	sgs.log.Info("getting user by email")
	md := sgs.baseMD(ctx)
	resp, err := sgs.ssoClient.GetUserByEmail(metadata.NewOutgoingContext(ctx, md), &ssopb.GetByEmailRequest{
		Email: email,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to get user by email",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return nil, http.StatusInternalServerError, consts.ErrInternalServer
		}

		return nil, code, errors.New(st.Message())
	}

	return &dtos.User{
		ID:        resp.Id,
		Email:     resp.Email,
		CreatedAt: resp.CreatedAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (sgs *ssoGatewayService) GetSelfUser(ctx *gin.Context) (*dtos.User, int, error) {
	sgs.log.Debug("getting self user")
	md, err := sgs.tokenMD(ctx)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	resp, err := sgs.ssoClient.GetSelfUser(metadata.NewOutgoingContext(ctx, md), nil)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return nil, http.StatusBadGateway, consts.ErrBadGateway
		}

		if st.Code() != codes.InvalidArgument {
			sgs.log.Error("failed to get user by jwt-token",
				slog.String("error", st.Message()),
				slog.String("code", st.Code().String()),
			)
		}

		code, ok := consts.CodeMap[st.Code()]
		if !ok {
			return nil, http.StatusInternalServerError, consts.ErrInternalServer
		}

		return nil, code, errors.New(st.Message())
	}

	return &dtos.User{
		ID:        resp.Id,
		Email:     resp.Email,
		CreatedAt: resp.CreatedAt.AsTime().UnixMilli(),
	}, http.StatusOK, nil
}

func (sgs *ssoGatewayService) tokenMD(ctx *gin.Context) (metadata.MD, error) {
	md := metadata.New(
		map[string]string{
			"x-client-ip":         ctx.ClientIP(),
			"x-client-user-agent": ctx.Request.UserAgent(),
		},
	)

	token, err := ctx.Cookie("ac")
	if err != nil {
		return nil, consts.ErrInvalidToken
	}
	md.Set("authorization", token)

	return md, nil
}

func (sgs *ssoGatewayService) baseMD(ctx *gin.Context) metadata.MD {
	md := metadata.New(
		map[string]string{
			"x-client-ip":         ctx.ClientIP(),
			"x-client-user-agent": ctx.Request.UserAgent(),
		},
	)

	return md
}
