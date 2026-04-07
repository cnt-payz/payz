package handlers

import (
	"context"

	paymentpb "github.com/cnt-payz/payz/payment-service/api/payment/v1"
	"github.com/cnt-payz/payz/payment-service/internal/application/services"
	"github.com/cnt-payz/payz/payment-service/pkg/consts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerAPI struct {
	paymentpb.UnimplementedPaymentServer
	service services.PaymentService
}

func New(service services.PaymentService) *ServerAPI {
	return &ServerAPI{
		service: service,
	}
}

func (sapi *ServerAPI) MakeExTransaction(ctx context.Context, req *paymentpb.MakeExRequest) (*paymentpb.ExTransaction, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrNilRequest.Error())
	}

	resp, err := sapi.service.MakeExTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) ConfirmExTransaction(ctx context.Context, req *paymentpb.ActionExRequest) (*paymentpb.ExTransaction, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrNilRequest.Error())
	}

	resp, err := sapi.service.ConfirmExTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) CancelExTransaction(ctx context.Context, req *paymentpb.ActionExRequest) (*paymentpb.ExTransaction, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrNilRequest.Error())
	}

	resp, err := sapi.service.CancelExTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetExHistory(ctx context.Context, req *paymentpb.GetHistoryRequest) (*paymentpb.ExHistory, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrNilRequest.Error())
	}

	resp, err := sapi.service.GetExHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sapi *ServerAPI) GetExTransaction(ctx context.Context, req *paymentpb.GetExRequest) (*paymentpb.ExTransaction, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrNilRequest.Error())
	}

	resp, err := sapi.service.GetExTransaction(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
