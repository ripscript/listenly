package service

import (
	"context"
	"errors"
	"listenly-backend/internal/auth/dto"
	models "listenly-backend/internal/auth/model"
	"listenly-backend/internal/auth/repository"
	"listenly-backend/pkg/hash"
	"listenly-backend/pkg/jwt"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired refresh token")
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userUUID string) error
	VerifyAccessToken(token string) (*jwt.Claims, error)
	GetUserByUUID(ctx context.Context, userUUID string) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
}

type authService struct {
	repo       repository.AuthRepository
	jwtManager *jwt.Manager
}

func NewAuthService(repo repository.AuthRepository, jwtManager *jwt.Manager) AuthService {
	return &authService{repo: repo, jwtManager: jwtManager}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existing, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
		FullName: req.FullName,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !hash.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	tokenHash := jwt.HashToken(req.RefreshToken)

	stored, err := s.repo.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.FindUserByID(ctx, stored.User.ID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if err := s.repo.RevokeRefreshToken(ctx, stored.ID); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := jwt.HashToken(refreshToken)
	stored, err := s.repo.FindRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // sudah invalid/logout, anggap sukses
		}
		return err
	}
	return s.repo.RevokeRefreshToken(ctx, stored.ID)
}

func (s *authService) LogoutAll(ctx context.Context, userUUID string) error {
	user, err := s.repo.FindUserByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // user tidak ada / sudah tidak punya sesi, anggap sukses (idempoten)
		}
		return err
	}
	return s.repo.RevokeAllUserRefreshTokens(ctx, user.ID)
}

func (s *authService) issueTokens(ctx context.Context, user *models.User) (*dto.AuthResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.UUID.String(), user.Email)
	if err != nil {
		return nil, err
	}

	plainRefresh, hashedRefresh, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshRecord := &models.RefreshToken{
		UserID:    user.ID,
		Token:     hashedRefresh,
		ExpiresAt: timeNowAdd(s.jwtManager.RefreshTTL()),
	}
	if err := s.repo.SaveRefreshToken(ctx, refreshRecord); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: plainRefresh,
		User: dto.UserSummary{
			UUID:     user.UUID.String(),
			Email:    user.Email,
			FullName: user.FullName,
		},
	}, nil
}

func (s *authService) VerifyAccessToken(token string) (*jwt.Claims, error) {
	return s.jwtManager.VerifyAccessToken(token)
}

func (s *authService) GetUserByUUID(ctx context.Context, userUUID string) (*models.User, error) {
	u, err := uuid.Parse(userUUID)
	if err != nil {
		return nil, errors.New("invalid uuid format")
	}
	user, err := s.repo.FindUserByUUID(ctx, u.String())
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
