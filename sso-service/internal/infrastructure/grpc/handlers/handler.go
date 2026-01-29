package handlers

import (
	"context"

	ssopb "github.com/cnt-payz/payz/sso-service/api/sso/v1"
	"github.com/cnt-payz/payz/sso-service/internal/application/services"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServerAPI struct {
	ssopb.UnimplementedSSOServer
	service services.SSOService
}

func New(service services.SSOService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (sapi *ServerAPI) Register(ctx context.Context, req *ssopb.RegisterRequest) (*ssopb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidRequest.Error())
	}

	resp, code, err := sapi.service.Register(ctx, req)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}

func (sapi *ServerAPI) Refresh(ctx context.Context, req *ssopb.RefreshRequest) (*ssopb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidRequest.Error())
	}

	resp, code, err := sapi.service.Refresh(ctx, req)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}

func (sapi *ServerAPI) LogoutAll(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if code, err := sapi.service.LogoutAll(ctx); err != nil {
		return nil, status.Error(code, err.Error())
	}

	return nil, nil
}

func (sapi *ServerAPI) Login(ctx context.Context, req *ssopb.LoginRequest) (*ssopb.Token, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidRequest.Error())
	}

	resp, code, err := sapi.service.Login(ctx, req)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}

func (sapi *ServerAPI) GetUserByID(ctx context.Context, req *ssopb.GetByIDRequest) (*ssopb.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidRequest.Error())
	}

	resp, code, err := sapi.service.GetUserByID(ctx, req)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}

func (sapi *ServerAPI) GetUserByEmail(ctx context.Context, req *ssopb.GetByEmailRequest) (*ssopb.User, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidRequest.Error())
	}

	resp, code, err := sapi.service.GetUserByEmail(ctx, req)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}

func (sapi *ServerAPI) GetSelfUser(ctx context.Context, _ *emptypb.Empty) (*ssopb.User, error) {
	resp, code, err := sapi.service.GetSelfUser(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}

func (sapi *ServerAPI) GetPublicKey(ctx context.Context, _ *emptypb.Empty) (*ssopb.PublicKey, error) {
	resp, code, err := sapi.service.GetPublicKey(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return resp, nil
}
