package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "listenly-backend/gen/go/auth/v1"
	"listenly-backend/internal/auth/dto"
	"listenly-backend/internal/auth/service"
)

type AuthGRPCHandler struct {
	authv1.UnimplementedAuthServiceServer
	service service.AuthService
}

func NewAuthGRPCHandler(service service.AuthService) *AuthGRPCHandler {
	return &AuthGRPCHandler{service: service}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	result, err := h.service.Register(ctx, dto.RegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		FullName: req.FullName,
	})
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoAuthResponse(result), nil
}

func (h *AuthGRPCHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	result, err := h.service.Login(ctx, dto.LoginRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoAuthResponse(result), nil
}

func (h *AuthGRPCHandler) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.AuthResponse, error) {
	result, err := h.service.Refresh(ctx, dto.RefreshRequest{
		RefreshToken: req.GetRefreshToken(),
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProtoAuthResponse(result), nil
}

func (h *AuthGRPCHandler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := h.service.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.LogoutResponse{Success: true}, nil
}

func (h *AuthGRPCHandler) LogoutAll(ctx context.Context, req *authv1.LogoutAllRequest) (*authv1.LogoutResponse, error) {
	if err := h.service.LogoutAll(ctx, req.GetUserUuid()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &authv1.LogoutResponse{Success: true}, nil
}

func (h *AuthGRPCHandler) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	claims, err := h.service.VerifyAccessToken(req.GetAccessToken())
	if err != nil {
		return &authv1.VerifyTokenResponse{Valid: false}, nil
	}
	return &authv1.VerifyTokenResponse{
		Valid:    true,
		UserUuid: claims.UserUUID,
		Email:    claims.Email,
	}, nil
}

func (h *AuthGRPCHandler) GetUserByUUID(ctx context.Context, req *authv1.GetUserByUUIDRequest) (*authv1.GetUserByUUIDResponse, error) {
	user, err := h.service.GetUserByUUID(ctx, req.GetUuid())
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &authv1.GetUserByUUIDResponse{
		Id:    int64(user.ID),
		Uuid:  user.UUID.String(),
		Email: user.Email,
	}, nil
}

func (h *AuthGRPCHandler) GetUserByID(ctx context.Context, req *authv1.GetUserByIDRequest) (*authv1.GetUserByUUIDResponse, error) {
	user, err := h.service.GetUserByID(ctx, uint(req.GetId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return &authv1.GetUserByUUIDResponse{
		Id:    int64(user.ID),
		Uuid:  user.UUID.String(),
		Email: user.Email,
	}, nil
}

func toProtoAuthResponse(r *dto.AuthResponse) *authv1.AuthResponse {
	return &authv1.AuthResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		User: &authv1.UserSummary{
			Uuid:     r.User.UUID,
			Email:    r.User.Email,
			FullName: r.User.FullName,
		},
	}
}
