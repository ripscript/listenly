package repository

import (
	"context"
	"errors"
	models "listenly-backend/internal/auth/model"
	"time"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	FindUserByID(ctx context.Context, id uint) (*models.User, error)
	FindUserByUUID(ctx context.Context, uuid string) (*models.User, error)

	SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error
	FindRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uint) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID uint) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *authRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) FindUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "uuid = ?", uuid).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) SaveRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *authRepository) FindRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token = ? AND revoked = false AND expires_at > ?", hash, time.Now()).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *authRepository) RevokeRefreshToken(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked", true).Error
}

func (r *authRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}
