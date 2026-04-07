package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	paymentpb "github.com/cnt-payz/payz/payment-service/api/payment/v1"
	"github.com/cnt-payz/payz/payment-service/internal/domain/extransaction"
	"github.com/cnt-payz/payz/payment-service/internal/domain/idempotency"
	"github.com/cnt-payz/payz/payment-service/internal/domain/notification"
	"github.com/cnt-payz/payz/payment-service/internal/domain/session"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/security/signature"
	"github.com/cnt-payz/payz/payment-service/pkg/consts"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentService interface {
	MakeExTransaction(ctx context.Context, req *paymentpb.MakeExRequest) (*paymentpb.ExTransaction, error)
	ConfirmExTransaction(ctx context.Context, req *paymentpb.ActionExRequest) (*paymentpb.ExTransaction, error)
	CancelExTransaction(ctx context.Context, req *paymentpb.ActionExRequest) (*paymentpb.ExTransaction, error)
	GetExHistory(ctx context.Context, req *paymentpb.GetHistoryRequest) (*paymentpb.ExHistory, error)
	GetExTransaction(ctx context.Context, req *paymentpb.GetExRequest) (*paymentpb.ExTransaction, error)
}

type paymentService struct {
	cfg               *config.Config
	log               *slog.Logger
	idempotencyRepo   idempotency.IdempotencyRepository
	extransactionRepo extransaction.ExTransactionRepository
	notificationRepo  notification.NotificationRepository
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	idempotencyKey idempotency.IdempotencyRepository,
	extransactionRepo extransaction.ExTransactionRepository,
	notificationRepo notification.NotificationRepository,
) PaymentService {
	return &paymentService{
		cfg:               cfg,
		log:               log,
		idempotencyRepo:   idempotencyKey,
		extransactionRepo: extransactionRepo,
		notificationRepo:  notificationRepo,
	}
}

func (ps *paymentService) MakeExTransaction(ctx context.Context, req *paymentpb.MakeExRequest) (*paymentpb.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	shopID, err := uuid.Parse(req.ShopId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidShopID.Error())
	}

	serverSignature := signature.GetSignature(
		req.Timestamp.AsTime(),
		shopID,
		ps.cfg.Service.Private,
	)

	if serverSignature != req.GetSingature() {
		return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidSignature.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ps.cfg.Server.Timeout)
	defer cancel()

	if code, err := ps.checkValidity(ctxTimeout, req.Timestamp.AsTime(), shopID); err != nil {
		return nil, status.Error(code, err.Error())
	}

	userID, err := ps.getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	transaction, err := ps.createExTransaction(req, userID, shopID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	savedTransaction, err := ps.extransactionRepo.Save(ctxTimeout, transaction)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrTransactionAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		ps.log.Error("failed to save ex-transaction",
			slog.String("transaction_id", transaction.ID().String()),
			slog.String("error", err.Error()),
		)
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	// TODO: add: create payment in crypto

	responseUserMetadata, err := ps.jsonToStruct(savedTransaction.UserMetadata())
	if err != nil {
		ps.log.Error("failed to get struct of user's metadata", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return &paymentpb.ExTransaction{
		Id:     savedTransaction.ID().String(),
		ShopId: savedTransaction.ShopID().String(),
		UserId: savedTransaction.UserID().String(),
		Status: paymentpb.Status(savedTransaction.Status().Value()),
		Typ:    paymentpb.Typ(savedTransaction.Typ().Value()),
		Amount: &paymentpb.Amount{
			Amount: savedTransaction.Amount().Value(),
		},
		CallbackUrl:  savedTransaction.CallbackURL().Value(),
		UserMetadata: responseUserMetadata,
		CreatedAt:    timestamppb.New(savedTransaction.CreatedAt()),
		UpdatedAt:    timestamppb.New(savedTransaction.UpdatedAt()),
	}, nil
}

func (ps *paymentService) ConfirmExTransaction(ctx context.Context, req *paymentpb.ActionExRequest) (*paymentpb.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	transactionID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidTransactionID.Error())
	}

	serverSignature := signature.GetActionSignature(
		req.Timestamp.AsTime(),
		transactionID,
		ps.cfg.Service.Private,
	)

	if serverSignature != req.GetSignature() {
		return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidSignature.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ps.cfg.Server.Timeout)
	defer cancel()

	if code, err := ps.checkValidity(ctxTimeout, req.Timestamp.AsTime(), transactionID); err != nil {
		return nil, status.Error(code, err.Error())
	}

	confirmedTransaction, err := ps.extransactionRepo.Confirm(ctxTimeout, transactionID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrTransactionIsntPending) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		if errors.Is(err, consts.ErrTransactionDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		ps.log.Error("failed to confirm ex-transaction",
			slog.String("transaction_id", transactionID.String()),
			slog.String("error", err.Error()),
		)
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go ps.sendMsg(context.Background(), confirmedTransaction)

	responseUserMetadata, err := ps.jsonToStruct(confirmedTransaction.UserMetadata())
	if err != nil {
		ps.log.Error("failed to get struct of user's metadata", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return &paymentpb.ExTransaction{
		Id:     confirmedTransaction.ID().String(),
		ShopId: confirmedTransaction.ShopID().String(),
		UserId: confirmedTransaction.UserID().String(),
		Status: paymentpb.Status(confirmedTransaction.Status().Value()),
		Typ:    paymentpb.Typ(confirmedTransaction.Typ().Value()),
		Amount: &paymentpb.Amount{
			Amount: confirmedTransaction.Amount().Value(),
		},
		CallbackUrl:  confirmedTransaction.CallbackURL().Value(),
		UserMetadata: responseUserMetadata,
		CreatedAt:    timestamppb.New(confirmedTransaction.CreatedAt()),
		UpdatedAt:    timestamppb.New(confirmedTransaction.UpdatedAt()),
	}, nil
}

func (ps *paymentService) CancelExTransaction(ctx context.Context, req *paymentpb.ActionExRequest) (*paymentpb.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	transactionID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidTransactionID.Error())
	}

	serverSignature := signature.GetActionSignature(
		req.Timestamp.AsTime(),
		transactionID,
		ps.cfg.Service.Private,
	)

	if serverSignature != req.GetSignature() {
		return nil, status.Error(codes.Unauthenticated, consts.ErrInvalidSignature.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ps.cfg.Server.Timeout)
	defer cancel()

	if code, err := ps.checkValidity(ctxTimeout, req.Timestamp.AsTime(), transactionID); err != nil {
		return nil, status.Error(code, err.Error())
	}

	canceledTransaction, err := ps.extransactionRepo.Cancel(ctxTimeout, transactionID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrTransactionIsntPending) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		if errors.Is(err, consts.ErrTransactionDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		ps.log.Error("failed to cancel ex-transaction",
			slog.String("transaction_id", transactionID.String()),
			slog.String("error", err.Error()),
		)
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	go ps.sendMsg(context.Background(), canceledTransaction)

	responseUserMetadata, err := ps.jsonToStruct(canceledTransaction.UserMetadata())
	if err != nil {
		ps.log.Error("failed to get struct of user's metadata", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return &paymentpb.ExTransaction{
		Id:     canceledTransaction.ID().String(),
		ShopId: canceledTransaction.ShopID().String(),
		UserId: canceledTransaction.UserID().String(),
		Status: paymentpb.Status(canceledTransaction.Status().Value()),
		Typ:    paymentpb.Typ(canceledTransaction.Typ().Value()),
		Amount: &paymentpb.Amount{
			Amount: canceledTransaction.Amount().Value(),
		},
		CallbackUrl:  canceledTransaction.CallbackURL().Value(),
		UserMetadata: responseUserMetadata,
		CreatedAt:    timestamppb.New(canceledTransaction.CreatedAt()),
		UpdatedAt:    timestamppb.New(canceledTransaction.UpdatedAt()),
	}, nil
}

func (ps *paymentService) GetExHistory(ctx context.Context, req *paymentpb.GetHistoryRequest) (*paymentpb.ExHistory, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.Error(codes.Canceled, err.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ps.cfg.Server.Timeout)
	defer cancel()

	shopID, err := uuid.Parse(req.GetShopId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := ps.extransactionRepo.GetExHistory(ctxTimeout, shopID, req.GetPage(), req.GetSize())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrInvalidArgs) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		ps.log.Error("failed to get ex-history",
			slog.String("error", err.Error()),
			slog.String("shop_id", shopID.String()),
		)
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	response := paymentpb.ExHistory{
		Timestamp:    timestamppb.New(time.Now().UTC()),
		Transactions: make([]*paymentpb.ExTransaction, len(resp)),
	}
	for idx, transaction := range resp {
		userMetadata, err := ps.jsonToStruct(transaction.UserMetadata())
		if err != nil {
			ps.log.Error("failed to convert user's metadata",
				slog.String("transaction_id", transaction.ID().String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		response.Transactions[idx] = &paymentpb.ExTransaction{
			Id:     transaction.ID().String(),
			ShopId: transaction.ShopID().String(),
			UserId: transaction.UserID().String(),
			Status: paymentpb.Status(transaction.Status().Value()),
			Typ:    paymentpb.Typ(transaction.Typ().Value()),
			Amount: &paymentpb.Amount{
				Amount: transaction.Amount().Value(),
			},
			CallbackUrl:  transaction.CallbackURL().Value(),
			UserMetadata: userMetadata,
			CreatedAt:    timestamppb.New(transaction.CreatedAt()),
			UpdatedAt:    timestamppb.New(transaction.UpdatedAt()),
		}
	}

	return &response, nil
}

func (ps *paymentService) GetExTransaction(ctx context.Context, req *paymentpb.GetExRequest) (*paymentpb.ExTransaction, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	transactionID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, consts.ErrInvalidShopID.Error())
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, ps.cfg.Server.Timeout)
	defer cancel()

	resp, err := ps.extransactionRepo.GetExTransaction(ctxTimeout, transactionID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		if errors.Is(err, consts.ErrTransactionDoesntExist) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		ps.log.Error("failed to get ex-transaction",
			slog.String("transaction_id", transactionID.String()),
			slog.String("error", err.Error()),
		)
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	responseUserMetadata, err := ps.jsonToStruct(resp.UserMetadata())
	if err != nil {
		ps.log.Error("failed to get struct of user's metadata", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, consts.ErrInternalServer.Error())
	}

	return &paymentpb.ExTransaction{
		Id:     resp.ID().String(),
		ShopId: resp.ShopID().String(),
		UserId: resp.UserID().String(),
		Status: paymentpb.Status(resp.Status().Value()),
		Typ:    paymentpb.Typ(resp.Typ().Value()),
		Amount: &paymentpb.Amount{
			Amount: resp.Amount().Value(),
		},
		CallbackUrl:  resp.CallbackURL().Value(),
		UserMetadata: responseUserMetadata,
		CreatedAt:    timestamppb.New(resp.CreatedAt()),
		UpdatedAt:    timestamppb.New(resp.UpdatedAt()),
	}, nil
}

func (ps *paymentService) checkValidity(ctx context.Context, timestamp time.Time, id uuid.UUID) (codes.Code, error) {
	now := time.Now().UTC()
	if now.Sub(timestamp) > ps.cfg.Service.Timecfg.Window {
		return codes.InvalidArgument, consts.ErrTooOldRequest
	}

	idempotencyKey, err := ps.getIdempotencyKey(ctx)
	if err != nil {
		return codes.InvalidArgument, err
	}

	if err := ps.checkIdempotency(ctx, idempotencyKey, id); err != nil {
		return codes.AlreadyExists, err
	}

	return codes.OK, nil
}

func (ps *paymentService) checkIdempotency(ctx context.Context, idempotencyKey string, id uuid.UUID) error {
	if err := ps.idempotencyRepo.Get(ctx, idempotencyKey, id); err == nil {
		return consts.ErrIdempotencyCheck
	} else if !errors.Is(err, consts.ErrIdempotencyNotFound) {
		ps.log.Error("failed to get idempotency key", slog.String("error", err.Error()))
	}

	if err := ps.idempotencyRepo.Save(ctx, idempotencyKey, id); err != nil {
		ps.log.Error("failed to save idempotency key", slog.String("error", err.Error()))
	}

	return nil
}

func (ps *paymentService) createExTransaction(req *paymentpb.MakeExRequest, userID, shopID uuid.UUID) (*extransaction.ExTransaction, error) {
	typ, err := extransaction.NewTypID(int(req.Typ))
	if err != nil {
		return nil, err
	}

	callbackURL, err := extransaction.NewCallbackURL(req.CallbackUrl)
	if err != nil {
		return nil, err
	}

	amount, err := extransaction.NewAmount(req.Amount.GetAmount())
	if err != nil {
		return nil, err
	}

	userMetadata, err := ps.structToJSON(req.UserMetadata)
	if err != nil {
		return nil, consts.ErrInvalidUserMetadata
	}

	return extransaction.New(
		shopID,
		userID,
		typ,
		amount,
		callbackURL,
		userMetadata,
	), nil
}

func (ps *paymentService) structToJSON(str *structpb.Struct) ([]byte, error) {
	if str == nil {
		return nil, consts.ErrInvalidUserMetadata
	}

	return protojson.Marshal(str)
}

func (ps *paymentService) jsonToStruct(bytes []byte) (*structpb.Struct, error) {
	var str structpb.Struct
	if err := protojson.Unmarshal(bytes, &str); err != nil {
		return nil, consts.ErrInvalidUserMetadata
	}

	return &str, nil
}

func (ps *paymentService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(session.CtxKey("user-id"))
	if raw == nil {
		return uuid.Nil, consts.ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, consts.ErrInvalidToken
}

func (ps *paymentService) getIdempotencyKey(ctx context.Context) (string, error) {
	raw := ctx.Value(session.CtxKey("idempotency-key"))
	if raw == nil {
		return "", consts.ErrIdempotencyNotFound
	}

	if key, ok := raw.(string); ok {
		return key, nil
	}

	return "", consts.ErrIdempotencyNotFound
}

func (ps *paymentService) sendMsg(ctx context.Context, transaction *extransaction.ExTransaction) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ps.cfg.Server.Timeout)
	defer cancel()

	msg := notification.Message{
		TransactionID: transaction.ID(),
		CallbackURL:   transaction.CallbackURL().Value(),
		UserMetadata:  transaction.UserMetadata(),
		Status:        transaction.Status().String(),
	}

	if err := ps.notificationRepo.SendMsg(ctxTimeout, &msg); err != nil {
		ps.log.Error("failed to send msg", slog.String("error", err.Error()))
	}
}
