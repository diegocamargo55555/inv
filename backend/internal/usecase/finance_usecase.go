package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type FinanceUseCase struct {
	accRepo    AccountRepository
	txRepo     TransactionRepository
	budgetRepo BudgetRepository
	marketData MarketDataProvider
}

func NewFinanceUseCase(
	accRepo AccountRepository,
	txRepo TransactionRepository,
	budgetRepo BudgetRepository,
) *FinanceUseCase {
	return &FinanceUseCase{
		accRepo:    accRepo,
		txRepo:     txRepo,
		budgetRepo: budgetRepo,
	}
}

func (uc *FinanceUseCase) WithMarketData(marketData MarketDataProvider) *FinanceUseCase {
	uc.marketData = marketData
	return uc
}

// CreateAccount creates a new bank account or wallet
func (uc *FinanceUseCase) CreateAccount(ctx context.Context, userID uuid.UUID, name string, accType domain.AccountType, initialBalance decimal.Decimal, currency, institution, color string) (*domain.Account, error) {
	if color == "" {
		color = "#10B981"
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "BRL"
	}
	acc := &domain.Account{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        name,
		Type:        accType,
		Balance:     initialBalance,
		Currency:    currency,
		Institution: institution,
		Color:       color,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := uc.accRepo.Create(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

// GetUserAccounts retrieves all accounts belonging to a user
func (uc *FinanceUseCase) GetUserAccounts(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	return uc.accRepo.GetByUserID(ctx, userID)
}

// CreateTransaction creates a single transaction and updates account balance atomically
func (uc *FinanceUseCase) CreateTransaction(
	ctx context.Context,
	userID uuid.UUID,
	accountID *uuid.UUID,
	categoryID *uuid.UUID,
	txType domain.TransactionType,
	amount decimal.Decimal,
	date time.Time,
	desc, notes, tags string,
) (*domain.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrInvalidAmount
	}

	if accountID == nil || *accountID == uuid.Nil {
		return nil, domain.ErrAccountRequired
	}

	acc, err := uc.accRepo.GetByID(ctx, *accountID)
	if err != nil || acc == nil || acc.UserID != userID {
		return nil, domain.ErrAccountNotFound
	}

	// Update balance
	if err := acc.ApplyTransaction(amount, txType); err != nil {
		return nil, err
	}
	if err := uc.accRepo.Update(ctx, acc); err != nil {
		return nil, err
	}

	tx := &domain.Transaction{
		ID:                uuid.New(),
		UserID:            userID,
		AccountID:         accountID,
		CategoryID:        categoryID,
		Type:              txType,
		Status:            domain.TxStatusCompleted,
		Amount:            amount,
		Date:              date,
		Description:       desc,
		Notes:             notes,
		Tags:              tags,
		InstallmentNumber: 1,
		TotalInstallments: 1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// GetTransactions gets user transactions with pagination
func (uc *FinanceUseCase) GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	return uc.txRepo.GetByUserID(ctx, userID, limit, offset)
}

// MonthlySummary calculates income, expense, net balance and category breakdown for a month
type MonthlySummary struct {
	TotalIncome  decimal.Decimal            `json:"total_income"`
	TotalExpense decimal.Decimal            `json:"total_expense"`
	NetBalance   decimal.Decimal            `json:"net_balance"`
	ByCategory   map[string]decimal.Decimal `json:"by_category"`
}

func (uc *FinanceUseCase) GetMonthlySummary(ctx context.Context, userID uuid.UUID, year, month int) (*MonthlySummary, error) {
	txs, err := uc.txRepo.GetByMonth(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	summary := &MonthlySummary{
		TotalIncome:  decimal.Zero,
		TotalExpense: decimal.Zero,
		NetBalance:   decimal.Zero,
		ByCategory:   make(map[string]decimal.Decimal),
	}

	usdBRLRate := decimal.NewFromInt(1)
	if uc.marketData != nil {
		if rate, err := uc.marketData.GetExchangeRate(ctx, "USD", "BRL"); err == nil && !rate.IsZero() {
			usdBRLRate = rate
		}
	}

	for _, tx := range txs {
		amountBRL := tx.Amount
		if tx.Account != nil && strings.EqualFold(tx.Account.Currency, "USD") {
			amountBRL = tx.Amount.Mul(usdBRLRate).Round(4)
		}

		if tx.Type == domain.TxTypeIncome {
			summary.TotalIncome = summary.TotalIncome.Add(amountBRL)
		} else if tx.Type == domain.TxTypeExpense {
			summary.TotalExpense = summary.TotalExpense.Add(amountBRL)
			catName := "Sem Categoria"
			if tx.Category != nil {
				catName = tx.Category.Name
			}
			summary.ByCategory[catName] = summary.ByCategory[catName].Add(amountBRL)
		}
	}
	summary.NetBalance = summary.TotalIncome.Sub(summary.TotalExpense)
	return summary, nil
}

// SetBudget creates or updates a category budget
func (uc *FinanceUseCase) SetBudget(ctx context.Context, userID uuid.UUID, categoryID *uuid.UUID, monthYear string, amountLimit decimal.Decimal) (*domain.Budget, error) {
	budget := &domain.Budget{
		ID:          uuid.New(),
		UserID:      userID,
		CategoryID:  categoryID,
		MonthYear:   monthYear,
		AmountLimit: amountLimit,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := uc.budgetRepo.CreateOrUpdate(ctx, budget); err != nil {
		return nil, err
	}
	return budget, nil
}

// GetBudgetsWithProgress retrieves all budgets for a given month with current spent and alerts
type BudgetProgressDTO struct {
	Budget     *domain.Budget  `json:"budget"`
	Spent      decimal.Decimal `json:"spent"`
	Percentage float64         `json:"percentage"`
	IsAlert    bool            `json:"is_alert"`
	IsExceeded bool            `json:"is_exceeded"`
}

func (uc *FinanceUseCase) GetBudgetsWithProgress(ctx context.Context, userID uuid.UUID, year, month int) ([]BudgetProgressDTO, error) {
	monthYear := fmt.Sprintf("%04d-%02d", year, month)
	budgets, err := uc.budgetRepo.GetByMonth(ctx, userID, monthYear)
	if err != nil {
		return nil, err
	}

	txs, err := uc.txRepo.GetByMonth(ctx, userID, year, month)
	if err != nil {
		return nil, err
	}

	spentByCat := make(map[uuid.UUID]decimal.Decimal)
	for _, tx := range txs {
		if tx.Type == domain.TxTypeExpense && tx.CategoryID != nil {
			spentByCat[*tx.CategoryID] = spentByCat[*tx.CategoryID].Add(tx.Amount)
		}
	}

	result := make([]BudgetProgressDTO, len(budgets))
	for i, b := range budgets {
		var spent decimal.Decimal
		if b.CategoryID != nil {
			spent = spentByCat[*b.CategoryID]
		}
		pct, isAlert, isExceeded := b.CalculateProgress(spent)
		result[i] = BudgetProgressDTO{
			Budget:     &budgets[i],
			Spent:      spent,
			Percentage: pct,
			IsAlert:    isAlert,
			IsExceeded: isExceeded,
		}
	}

	return result, nil
}

// UpdateTransaction updates an existing transaction and adjusts account balance
func (uc *FinanceUseCase) UpdateTransaction(
	ctx context.Context,
	userID uuid.UUID,
	txID uuid.UUID,
	accountID *uuid.UUID,
	categoryID *uuid.UUID,
	txType domain.TransactionType,
	amount decimal.Decimal,
	date time.Time,
	desc, notes, tags string,
) (*domain.Transaction, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrInvalidAmount
	}

	if accountID == nil || *accountID == uuid.Nil {
		return nil, domain.ErrAccountRequired
	}

	tx, err := uc.txRepo.GetByID(ctx, txID)
	if err != nil || tx == nil || tx.UserID != userID {
		return nil, domain.ErrTransactionNotFound
	}

	// Verify target account exists and belongs to user
	newAcc, err := uc.accRepo.GetByID(ctx, *accountID)
	if err != nil || newAcc == nil || newAcc.UserID != userID {
		return nil, domain.ErrAccountNotFound
	}

	// Revert old transaction amount on old account
	if tx.AccountID != nil {
		oldAcc, err := uc.accRepo.GetByID(ctx, *tx.AccountID)
		if err == nil && oldAcc != nil && oldAcc.UserID == userID {
			if tx.Type == domain.TxTypeIncome {
				_ = oldAcc.ApplyTransaction(tx.Amount, domain.TxTypeExpense)
			} else if tx.Type == domain.TxTypeExpense {
				_ = oldAcc.ApplyTransaction(tx.Amount, domain.TxTypeIncome)
			}
			_ = uc.accRepo.Update(ctx, oldAcc)
		}
	}

	// If old and new account are the same, reload newAcc so we do not overwrite reverted balance with stale state
	if tx.AccountID != nil && *tx.AccountID == *accountID {
		newAcc, _ = uc.accRepo.GetByID(ctx, *accountID)
	}

	// Apply new transaction amount on new account
	if err := newAcc.ApplyTransaction(amount, txType); err != nil {
		return nil, err
	}
	if err := uc.accRepo.Update(ctx, newAcc); err != nil {
		return nil, err
	}

	tx.AccountID = accountID
	tx.CategoryID = categoryID
	tx.Type = txType
	tx.Amount = amount
	tx.Date = date
	tx.Description = desc
	tx.Notes = notes
	tx.Tags = tags
	tx.UpdatedAt = time.Now()

	if err := uc.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}

	return uc.txRepo.GetByID(ctx, txID)
}

// DeleteTransaction removes a transaction and reverts its balance effect
func (uc *FinanceUseCase) DeleteTransaction(ctx context.Context, userID, txID uuid.UUID) error {
	tx, err := uc.txRepo.GetByID(ctx, txID)
	if err != nil || tx == nil || tx.UserID != userID {
		return domain.ErrTransactionNotFound
	}

	if tx.AccountID != nil {
		acc, err := uc.accRepo.GetByID(ctx, *tx.AccountID)
		if err == nil && acc != nil && acc.UserID == userID {
			if tx.Type == domain.TxTypeIncome {
				_ = acc.ApplyTransaction(tx.Amount, domain.TxTypeExpense)
			} else if tx.Type == domain.TxTypeExpense {
				_ = acc.ApplyTransaction(tx.Amount, domain.TxTypeIncome)
			}
			_ = uc.accRepo.Update(ctx, acc)
		}
	}

	return uc.txRepo.Delete(ctx, txID)
}
