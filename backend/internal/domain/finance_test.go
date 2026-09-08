package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccount_ApplyTransaction(t *testing.T) {
	account := &domain.Account{
		ID:      uuid.New(),
		Name:    "Nubank",
		Type:    domain.AccountTypeChecking,
		Balance: decimal.NewFromFloat(1000.00),
	}

	t.Run("apply income increases balance", func(t *testing.T) {
		err := account.ApplyTransaction(decimal.NewFromFloat(500.50), domain.TxTypeIncome)
		require.NoError(t, err)
		assert.True(t, account.Balance.Equal(decimal.NewFromFloat(1500.50)))
	})

	t.Run("apply expense decreases balance", func(t *testing.T) {
		err := account.ApplyTransaction(decimal.NewFromFloat(200.25), domain.TxTypeExpense)
		require.NoError(t, err)
		assert.True(t, account.Balance.Equal(decimal.NewFromFloat(1300.25)))
	})

	t.Run("apply negative or zero amount returns error", func(t *testing.T) {
		err := account.ApplyTransaction(decimal.NewFromFloat(-50.00), domain.TxTypeExpense)
		assert.Error(t, err)

		err = account.ApplyTransaction(decimal.Zero, domain.TxTypeIncome)
		assert.Error(t, err)
	})
}

func TestBudget_ProgressAndAlerts(t *testing.T) {
	budget := &domain.Budget{
		ID:          uuid.New(),
		AmountLimit: decimal.NewFromFloat(1000.00),
		MonthYear:   "2026-03",
	}

	t.Run("budget under alert threshold", func(t *testing.T) {
		spent := decimal.NewFromFloat(500.00)
		pct, isAlert, isExceeded := budget.CalculateProgress(spent)

		assert.InDelta(t, 50.0, pct, 0.01)
		assert.False(t, isAlert)
		assert.False(t, isExceeded)
	})

	t.Run("budget reaches alert threshold (80%)", func(t *testing.T) {
		spent := decimal.NewFromFloat(850.00)
		pct, isAlert, isExceeded := budget.CalculateProgress(spent)

		assert.InDelta(t, 85.0, pct, 0.01)
		assert.True(t, isAlert)
		assert.False(t, isExceeded)
	})

	t.Run("budget exceeded (> 100%)", func(t *testing.T) {
		spent := decimal.NewFromFloat(1150.00)
		pct, isAlert, isExceeded := budget.CalculateProgress(spent)

		assert.InDelta(t, 115.0, pct, 0.01)
		assert.True(t, isAlert)
		assert.True(t, isExceeded)
	})
}
