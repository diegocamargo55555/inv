package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CreateAccountRequest struct {
	Name           string             `json:"name" binding:"required"`
	Type           domain.AccountType `json:"type" binding:"required"`
	InitialBalance decimal.Decimal    `json:"initial_balance"`
	Currency       string             `json:"currency"`
	Institution    string             `json:"institution"`
	Color          string             `json:"color"`
}

type CreateTransactionRequest struct {
	AccountID   *uuid.UUID             `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID             `json:"category_id"`
	Type        domain.TransactionType `json:"type" binding:"required"`
	Amount      decimal.Decimal        `json:"amount" binding:"required"`
	Date        time.Time              `json:"date" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	Notes       string                 `json:"notes"`
	Tags        string                 `json:"tags"`
}

type SetBudgetRequest struct {
	CategoryID  *uuid.UUID      `json:"category_id"`
	MonthYear   string          `json:"month_year" binding:"required"` // e.g. "2026-03"
	AmountLimit decimal.Decimal `json:"amount_limit" binding:"required"`
}

type ExecuteOrderRequest struct {
	PortfolioID uuid.UUID       `json:"portfolio_id" binding:"required"`
	AssetID     uuid.UUID       `json:"asset_id" binding:"required"`
	AccountID   *uuid.UUID      `json:"account_id"`
	Quantity    decimal.Decimal `json:"quantity" binding:"required"`
	UnitPrice   decimal.Decimal `json:"unit_price" binding:"required"`
	Fees        decimal.Decimal `json:"fees"`
	Date        time.Time       `json:"date" binding:"required"`
	Notes       string          `json:"notes"`
}

type UpdateInvestmentOrderRequest struct {
	AccountID *uuid.UUID      `json:"account_id"`
	Quantity  decimal.Decimal `json:"quantity" binding:"required"`
	UnitPrice decimal.Decimal `json:"unit_price" binding:"required"`
	Fees      decimal.Decimal `json:"fees"`
	Date      time.Time       `json:"date" binding:"required"`
	Notes     string          `json:"notes"`
}

type UpdateTransactionRequest struct {
	AccountID   *uuid.UUID             `json:"account_id" binding:"required"`
	CategoryID  *uuid.UUID             `json:"category_id"`
	Type        domain.TransactionType `json:"type" binding:"required"`
	Amount      decimal.Decimal        `json:"amount" binding:"required"`
	Date        time.Time              `json:"date" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	Notes       string                 `json:"notes"`
	Tags        string                 `json:"tags"`
}
