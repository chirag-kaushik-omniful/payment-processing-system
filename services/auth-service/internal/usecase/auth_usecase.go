package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/omniful/payment-platform/shared/auth"
	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/services/auth-service/internal/model"
	"github.com/omniful/payment-platform/services/auth-service/internal/repository"
)

type AuthUsecase struct {
	repo   *repository.UserRepository
	jwtMgr *auth.JWTManager
}

func NewAuthUsecase(repo *repository.UserRepository, jwtMgr *auth.JWTManager) *AuthUsecase {
	return &AuthUsecase{repo: repo, jwtMgr: jwtMgr}
}

func (u *AuthUsecase) Signup(ctx context.Context, req model.SignupRequest) (*model.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal("hash password failed", err)
	}
	user, err := u.repo.Create(ctx, req.Email, string(hash))
	if err != nil {
		return nil, apperrors.Conflict("email already registered")
	}
	return u.issueTokens(ctx, user.ID, user.Email, user.Roles)
}

func (u *AuthUsecase) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	user, err := u.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.Unauthorized("invalid credentials")
	}
	return u.issueTokens(ctx, user.ID, user.Email, user.Roles)
}

func (u *AuthUsecase) Refresh(ctx context.Context, req model.RefreshRequest) (*model.AuthResponse, error) {
	claims, err := u.jwtMgr.Validate(req.RefreshToken)
	if err != nil {
		return nil, apperrors.Unauthorized("invalid refresh token")
	}
	_ = u.repo.RevokeRefreshTokens(ctx, claims.UserID)
	return u.issueTokens(ctx, claims.UserID, claims.Email, claims.Roles)
}

func (u *AuthUsecase) issueTokens(ctx context.Context, userID, email string, roles []string) (*model.AuthResponse, error) {
	pair, err := u.jwtMgr.GeneratePair(userID, email, roles)
	if err != nil {
		return nil, apperrors.Internal("token generation failed", err)
	}
	hash := hashToken(pair.RefreshToken)
	expires := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	_ = u.repo.SaveRefreshToken(ctx, userID, hash, expires)
	return &model.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		UserID:       userID,
		Email:        email,
	}, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
