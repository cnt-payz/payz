package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	ssopb "github.com/cnt-payz/payz/sso-service/api/sso/v1"
	"github.com/cnt-payz/payz/sso-service/internal/domain/filemanager"
	"github.com/cnt-payz/payz/sso-service/internal/domain/session"
	"github.com/cnt-payz/payz/sso-service/internal/domain/user"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-service/pkg/consts"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SSOService interface {
	Register(ctx context.Context, req *ssopb.RegisterRequest) (*ssopb.Token, codes.Code, error)
	Login(ctx context.Context, req *ssopb.LoginRequest) (*ssopb.Token, codes.Code, error)
	Refresh(ctx context.Context, req *ssopb.RefreshRequest) (*ssopb.Token, codes.Code, error)
	LogoutAll(ctx context.Context) (codes.Code, error)
	GetUserByEmail(ctx context.Context, req *ssopb.GetByEmailRequest) (*ssopb.User, codes.Code, error)
	GetUserByID(ctx context.Context, req *ssopb.GetByIDRequest) (*ssopb.User, codes.Code, error)
	GetSelfUser(ctx context.Context) (*ssopb.User, codes.Code, error)
	GetPublicKey(ctx context.Context) (*ssopb.PublicKey, codes.Code, error)
}

type ssoService struct {
	cfg         *config.Config
	log         *slog.Logger
	userRepo    user.UserRepository
	userCache   user.UserCache
	sessionRepo session.SessionRepository
	jwtMngr     session.JWTManager
	fileMngr    filemanager.FileManager
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	userRepo user.UserRepository,
	userCache user.UserCache,
	sessionRepo session.SessionRepository,
	jwtMngr session.JWTManager,
	fileMngr filemanager.FileManager,
) SSOService {
	return &ssoService{
		cfg:         cfg,
		log:         log,
		userRepo:    userRepo,
		userCache:   userCache,
		sessionRepo: sessionRepo,
		jwtMngr:     jwtMngr,
		fileMngr:    fileMngr,
	}
}

func (s *ssoService) Register(ctx context.Context, req *ssopb.RegisterRequest) (*ssopb.Token, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	s.log.Debug("creating domain user", slog.String("email", req.Email))
	email, err := user.NewEmail(req.GetEmail())
	if err != nil {
		return nil, codes.InvalidArgument, err
	}

	passwordHash, err := user.NewPasswordHash(req.GetPassword())
	if err != nil {
		return nil, codes.InvalidArgument, err
	}

	user := user.New(
		email,
		passwordHash,
	)

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("saving user", slog.String("user_id", user.ID().String()))
	savedUser, err := s.userRepo.Save(ctxTimeout, user)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, codes.DeadlineExceeded, err
		}

		if errors.Is(err, consts.ErrUserAlreadyExists) {
			return nil, codes.AlreadyExists, err
		}

		s.log.Error("failed to save user", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	access, refresh, err := s.generatePairTokens(savedUser)
	if err != nil {
		s.log.Error("failed to generate a pair of tokens", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	session, err := s.createSession(ctx, savedUser)
	if err != nil {
		return nil, codes.InvalidArgument, err
	}

	s.log.Debug("setting session")
	if err := s.sessionRepo.Set(ctxTimeout, refresh, session); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, codes.DeadlineExceeded, err
		}

		s.log.Error("failed to set session into cache", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	now := time.Now().UTC()
	return &ssopb.Token{
		Access:           access,
		Refresh:          refresh,
		AccessExpiresAt:  timestamppb.New(now.Add(s.cfg.Secrets.JWT.AccessTTL)),
		RefreshExpiresAt: timestamppb.New(now.Add(s.cfg.Secrets.Redis.SessionTTL)),
	}, codes.OK, nil
}

func (s *ssoService) Login(ctx context.Context, req *ssopb.LoginRequest) (*ssopb.Token, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	email, err := user.NewEmail(req.GetEmail())
	if err != nil {
		return nil, codes.InvalidArgument, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("fetching user by email", slog.String("email", req.Email))
	fetchedUser, err := s.userRepo.GetByEmail(ctxTimeout, email)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, codes.DeadlineExceeded, err
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, codes.NotFound, err
		}

		s.log.Error("failed to fetch user by email", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	access, refresh, err := s.generatePairTokens(fetchedUser)
	if err != nil {
		s.log.Error("failed to generate a pair of tokens", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	session, err := s.createSession(ctx, fetchedUser)
	if err != nil {
		return nil, codes.InvalidArgument, err
	}

	s.log.Debug("setting session")
	if err := s.sessionRepo.Set(ctxTimeout, refresh, session); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, codes.DeadlineExceeded, err
		}

		s.log.Error("failed to set session into cache", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	now := time.Now().UTC()
	return &ssopb.Token{
		Access:           access,
		Refresh:          refresh,
		AccessExpiresAt:  timestamppb.New(now.Add(s.cfg.Secrets.JWT.AccessTTL)),
		RefreshExpiresAt: timestamppb.New(now.Add(s.cfg.Secrets.Redis.SessionTTL)),
	}, codes.OK, nil
}

func (s *ssoService) Refresh(ctx context.Context, req *ssopb.RefreshRequest) (*ssopb.Token, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("checking session")
	oldSession, err := s.sessionRepo.Get(ctxTimeout, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, codes.DeadlineExceeded, err
		}

		if errors.Is(err, consts.ErrSessionDoesntExist) {
			return nil, codes.NotFound, err
		}

		s.log.Error("failed to get session", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	user := user.ForSession(
		oldSession.UserID(),
		user.Email(oldSession.Email()),
	)

	s.log.Debug("create new session")
	newSession, err := s.createSession(
		ctx,
		user,
	)
	if err != nil {
		s.log.Error("failed to generate new session", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	s.log.Debug("checking fingerprint")
	if newSession.FingerPrint().Value() != oldSession.FingerPrint().Value() {
		return nil, codes.PermissionDenied, consts.ErrInvalidFingerprint
	}

	access, refresh, err := s.generatePairTokens(user)
	if err != nil {
		s.log.Error("failed to generate a pair of tokens", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	if err := s.dropSession(ctxTimeout, req.GetRefreshToken()); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, codes.DeadlineExceeded, err
		}

		s.log.Error("failed to delete old session from cache", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	if err := s.sessionRepo.Set(ctxTimeout, refresh, newSession); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, codes.DeadlineExceeded, err
		}

		s.log.Error("failed to save session into cache", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	now := time.Now().UTC()
	return &ssopb.Token{
		Access:           access,
		Refresh:          refresh,
		AccessExpiresAt:  timestamppb.New(now.Add(s.cfg.Secrets.JWT.AccessTTL)),
		RefreshExpiresAt: timestamppb.New(now.Add(s.cfg.Secrets.Redis.SessionTTL)),
	}, codes.OK, nil
}

func (s *ssoService) LogoutAll(ctx context.Context) (codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return codes.Canceled, err
	}

	s.log.Debug("get user's id")
	userID, err := s.getUserID(ctx)
	if err != nil {
		return codes.Unauthenticated, err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	s.log.Debug("logouting")
	if err := s.sessionRepo.LogoutAll(ctxTimeout, userID); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return codes.DeadlineExceeded, err
		}

		s.log.Error("failed to logout all sessions", slog.String("error", err.Error()))
		return codes.Internal, consts.ErrInternalServer
	}

	return codes.OK, nil
}

func (s *ssoService) GetUserByID(ctx context.Context, req *ssopb.GetByIDRequest) (*ssopb.User, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	s.log.Debug("checking user's id")
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, codes.InvalidArgument, consts.ErrInvalidUserID
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	if user, err := s.userCache.GetByID(ctxTimeout, userID); err == nil {
		s.log.Debug("user fetched from cache")
		return &ssopb.User{
			Id:        user.ID().String(),
			Email:     user.Email().Value(),
			CreatedAt: timestamppb.New(user.CreatedAt()),
		}, codes.OK, nil
	} else if !errors.Is(err, consts.ErrUserDoesntExist) {
		s.log.Error("failed to get user by id from cache", slog.String("error", err.Error()))
	}

	s.log.Debug("fetching user", slog.String("user_id", userID.String()))
	user, err := s.userRepo.GetByID(ctxTimeout, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, codes.DeadlineExceeded, err
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, codes.NotFound, err
		}

		s.log.Error("failed to fetch user", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	go s.saveCacheID(context.Background(), user)
	return &ssopb.User{
		Id:        user.ID().String(),
		Email:     user.Email().Value(),
		CreatedAt: timestamppb.New(user.CreatedAt()),
	}, codes.OK, nil
}

func (s *ssoService) GetUserByEmail(ctx context.Context, req *ssopb.GetByEmailRequest) (*ssopb.User, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	s.log.Debug("checking user's email")
	email, err := user.NewEmail(req.GetEmail())
	if err != nil {
		return nil, codes.InvalidArgument, consts.ErrInvalidEmail
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	if user, err := s.userCache.GetByEmail(ctxTimeout, email); err == nil {
		s.log.Debug("user fetched from cache")
		return &ssopb.User{
			Id:        user.ID().String(),
			Email:     user.Email().Value(),
			CreatedAt: timestamppb.New(user.CreatedAt()),
		}, codes.OK, nil
	} else if !errors.Is(err, consts.ErrUserDoesntExist) {
		s.log.Error("failed to get user from cache by email", slog.String("error", err.Error()))
	}

	s.log.Debug("fetching user", slog.String("email", email.Value()))
	user, err := s.userRepo.GetByEmail(ctxTimeout, email)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, codes.DeadlineExceeded, err
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, codes.NotFound, err
		}

		s.log.Error("failed to fetch user", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	go s.saveCacheEmail(context.Background(), user)
	return &ssopb.User{
		Id:        user.ID().String(),
		Email:     user.Email().Value(),
		CreatedAt: timestamppb.New(user.CreatedAt()),
	}, codes.OK, nil
}

func (s *ssoService) GetPublicKey(ctx context.Context) (*ssopb.PublicKey, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	s.log.Debug("load public key path")
	bytes, err := s.fileMngr.LoadFile(s.cfg.Secrets.JWT.PublicKeyPath)
	if err != nil {
		s.log.Error("failed to load file", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	return &ssopb.PublicKey{
		Body: bytes,
	}, codes.OK, nil
}

func (s *ssoService) GetSelfUser(ctx context.Context) (*ssopb.User, codes.Code, error) {
	if err := ctx.Err(); err != nil {
		return nil, codes.Canceled, err
	}

	s.log.Debug("checking user's id")
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, codes.Unauthenticated, consts.ErrInvalidToken
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	if user, err := s.userCache.GetByID(ctxTimeout, userID); err == nil {
		return &ssopb.User{
			Id:        user.ID().String(),
			Email:     user.Email().Value(),
			CreatedAt: timestamppb.New(user.CreatedAt()),
		}, codes.OK, nil
	} else if !errors.Is(err, consts.ErrUserDoesntExist) {
		s.log.Error("failed to get user by id from cache", slog.String("error", err.Error()))
	}

	s.log.Debug("fetching user", slog.String("user_id", userID.String()))
	user, err := s.userRepo.GetByID(ctxTimeout, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, codes.DeadlineExceeded, err
		}

		if errors.Is(err, consts.ErrUserDoesntExist) {
			return nil, codes.NotFound, err
		}

		s.log.Error("failed to fetch user", slog.String("error", err.Error()))
		return nil, codes.Internal, consts.ErrInternalServer
	}

	go s.saveCacheID(context.Background(), user)
	return &ssopb.User{
		Id:        user.ID().String(),
		Email:     user.Email().Value(),
		CreatedAt: timestamppb.New(user.CreatedAt()),
	}, codes.OK, nil
}

func (s *ssoService) dropSession(ctx context.Context, refresh string) error {
	if err := s.sessionRepo.Del(ctx, refresh); err != nil {
		return err
	}

	return nil
}

func (s *ssoService) createSession(ctx context.Context, user *user.User) (*session.Session, error) {
	ip, err := s.getIP(ctx)
	if err != nil {
		return nil, err
	}

	userAgent, err := s.getUserAgent(ctx)
	if err != nil {
		return nil, err
	}

	fingerPrint, err := session.NewFingerPrint(ip, userAgent)
	if err != nil {
		return nil, err
	}

	return session.New(
		user.ID(),
		user.Email().Value(),
		fingerPrint,
	), nil
}

func (s *ssoService) getUserID(ctx context.Context) (uuid.UUID, error) {
	rawID := ctx.Value(session.CtxKey("user-id"))
	if rawID == nil {
		return uuid.Nil, consts.ErrInvalidToken
	}

	if id, ok := rawID.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, consts.ErrInvalidToken
}

func (s *ssoService) generatePairTokens(user *user.User) (string, string, error) {
	if user == nil {
		return "", "", consts.ErrNilArgs
	}

	access, err := s.jwtMngr.GetAccess(user.ID(), user.Email())
	if err != nil {
		return "", "", err
	}

	refresh, err := s.jwtMngr.GetRefresh()
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *ssoService) getIP(ctx context.Context) (string, error) {
	rawIP := ctx.Value(session.CtxKey("x-client-ip"))
	if rawIP == nil {
		return "", consts.ErrInvalidIP
	}

	if ip, ok := rawIP.(string); ok {
		return ip, nil
	}

	return "", consts.ErrInvalidIP
}

func (s *ssoService) getUserAgent(ctx context.Context) (string, error) {
	rawUserAgent := ctx.Value(session.CtxKey("user-agent"))
	if rawUserAgent == nil {
		return "", consts.ErrInvalidUserAgent
	}

	if userAgent, ok := rawUserAgent.(string); ok {
		return userAgent, nil
	}

	return "", consts.ErrInvalidUserAgent
}

func (s *ssoService) saveCacheID(ctx context.Context, user *user.User) {
	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	if err := s.userCache.SetByID(ctxTimeout, user); err != nil {
		s.log.Error("failed to set user by id into cache", slog.String("error", err.Error()))
	}
}

func (s *ssoService) saveCacheEmail(ctx context.Context, user *user.User) {
	ctxTimeout, cancel := context.WithTimeout(ctx, s.cfg.Server.Timeout)
	defer cancel()

	if err := s.userCache.SetByEmail(ctxTimeout, user); err != nil {
		s.log.Error("failed to set user by id into cache", slog.String("error", err.Error()))
	}
}
