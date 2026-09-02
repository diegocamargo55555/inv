package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/invest/backend/pkg/hash"
	"github.com/invest/backend/pkg/token"
)

type AuthUseCase struct {
	userRepo      UserRepository
	portfolioRepo PortfolioRepository
	tokenMaker    token.Maker
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewAuthUseCase(
	userRepo UserRepository,
	portfolioRepo PortfolioRepository,
	tokenMaker token.Maker,
	accessExpiry time.Duration,
	refreshExpiry time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      userRepo,
		portfolioRepo: portfolioRepo,
		tokenMaker:    tokenMaker,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

type AuthResult struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *domain.User `json:"user"`
}

func (uc *AuthUseCase) Register(ctx context.Context, name, email, password string) (*AuthResult, error) {
	existing, _ := uc.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		BaseCurrency: "BRL",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Create default investment portfolio for user
	defaultPortfolio := &domain.Portfolio{
		ID:          uuid.New(),
		UserID:      user.ID,
		Name:        "Carteira Principal",
		Description: "Carteira padrão consolidada",
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = uc.portfolioRepo.Create(ctx, defaultPortfolio)

	return uc.generateTokens(ctx, user)
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := hash.CheckPassword(password, user.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return uc.generateTokens(ctx, user)
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshTokenStr string) (*AuthResult, error) {
	payload, err := uc.tokenMaker.VerifyToken(refreshTokenStr)
	if err != nil {
		return nil, err
	}

	storedToken, err := uc.userRepo.GetRefreshToken(ctx, refreshTokenStr)
	if err != nil || storedToken == nil {
		return nil, token.ErrInvalidToken
	}

	user, err := uc.userRepo.GetByID(ctx, payload.UserID)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Invalidate previous refresh token (rotation)
	_ = uc.userRepo.DeleteRefreshToken(ctx, refreshTokenStr)

	return uc.generateTokens(ctx, user)
}

func (uc *AuthUseCase) generateTokens(ctx context.Context, user *domain.User) (*AuthResult, error) {
	accessToken, _, err := uc.tokenMaker.CreateToken(user.ID, user.Email, token.TokenTypeAccess, uc.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := uc.tokenMaker.CreateToken(user.ID, user.Email, token.TokenTypeRefresh, uc.refreshExpiry)
	if err != nil {
		return nil, err
	}

	rf := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(uc.refreshExpiry),
		CreatedAt: time.Now(),
	}
	if err := uc.userRepo.CreateRefreshToken(ctx, rf); err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}
