package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

// UserRepository defines data operations for Users
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}

// AccountRepository defines data operations for Accounts
type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error)
	Update(ctx context.Context, account *domain.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// TransactionRepository defines data operations for Transactions
type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error)
	GetByMonth(ctx context.Context, userID uuid.UUID, year int, month int) ([]domain.Transaction, error)
	Update(ctx context.Context, tx *domain.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CategoryRepository defines data operations for Categories
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
}

// CreditCardRepository defines data operations for Credit Cards and Invoices
type CreditCardRepository interface {
	Create(ctx context.Context, card *domain.CreditCard) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.CreditCard, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CreditCard, error)
	CreateTransactions(ctx context.Context, txs []domain.Transaction) error
	GetInvoiceTransactions(ctx context.Context, cardID uuid.UUID, invoiceMonth string) ([]domain.Transaction, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// BudgetRepository defines data operations for Budgets
type BudgetRepository interface {
	CreateOrUpdate(ctx context.Context, budget *domain.Budget) error
	GetByMonth(ctx context.Context, userID uuid.UUID, monthYear string) ([]domain.Budget, error)
}

// PortfolioRepository defines data operations for Portfolios and Positions
type PortfolioRepository interface {
	Create(ctx context.Context, portfolio *domain.Portfolio) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Portfolio, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Portfolio, error)
	GetPosition(ctx context.Context, portfolioID, assetID uuid.UUID) (*domain.Position, error)
	SavePosition(ctx context.Context, pos *domain.Position) error
	GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]domain.Position, error)
}

// AssetRepository defines data operations for Assets
type AssetRepository interface {
	CreateOrUpdate(ctx context.Context, asset *domain.Asset) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Asset, error)
	GetByTicker(ctx context.Context, ticker string) (*domain.Asset, error)
	Search(ctx context.Context, query string) ([]domain.Asset, error)
	UpdatePrice(ctx context.Context, assetID uuid.UUID, price decimal.Decimal) error
}

// InvestmentTxRepository defines operations for Investment Orders
type InvestmentTxRepository interface {
	Create(ctx context.Context, tx *domain.InvestmentTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentTransaction, error)
	GetByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]domain.InvestmentTransaction, error)
	GetByPortfolioAndAsset(ctx context.Context, portfolioID, assetID uuid.UUID) ([]domain.InvestmentTransaction, error)
	Update(ctx context.Context, tx *domain.InvestmentTransaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// EarningRepository defines operations for Earnings & Dividends
type EarningRepository interface {
	Create(ctx context.Context, earning *domain.Earning) error
	GetByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]domain.Earning, error)
	GetByDateRange(ctx context.Context, portfolioID uuid.UUID, start, end time.Time) ([]domain.Earning, error)
}

// MarketDataProvider defines external quotes and exchange rate services
type MarketDataProvider interface {
	GetQuote(ctx context.Context, ticker string) (decimal.Decimal, error)
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error)
}
