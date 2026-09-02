package domain_test

import (
	"testing"
	"time"

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

func TestCreditCard_InvoiceDatesAndInstallments(t *testing.T) {
	card := &domain.CreditCard{
		ID:         uuid.New(),
		Name:       "Mastercard Black",
		Limit:      decimal.NewFromFloat(10000.00),
		ClosingDay: 20,
		DueDay:     27,
	}

	t.Run("purchase before closing day enters current month invoice", func(t *testing.T) {
		purchaseDate := time.Date(2026, time.March, 15, 10, 0, 0, 0, time.UTC)
		closingDate, dueDate, monthYear := card.CalculateInvoiceDates(purchaseDate)

		assert.Equal(t, "2026-03", monthYear)
		assert.Equal(t, 20, closingDate.Day())
		assert.Equal(t, time.March, closingDate.Month())
		assert.Equal(t, 27, dueDate.Day())
		assert.Equal(t, time.March, dueDate.Month())
	})

	t.Run("purchase after closing day enters next month invoice (melhor dia de compra)", func(t *testing.T) {
		purchaseDate := time.Date(2026, time.March, 21, 15, 30, 0, 0, time.UTC)
		closingDate, dueDate, monthYear := card.CalculateInvoiceDates(purchaseDate)

		assert.Equal(t, "2026-04", monthYear)
		assert.Equal(t, 20, closingDate.Day())
		assert.Equal(t, time.April, closingDate.Month())
		assert.Equal(t, 27, dueDate.Day())
		assert.Equal(t, time.April, dueDate.Month())
	})

	t.Run("split installments with precision and remainder allocation", func(t *testing.T) {
		totalAmount := decimal.NewFromFloat(100.00)
		installmentsCount := 3
		purchaseDate := time.Date(2026, time.January, 10, 12, 0, 0, 0, time.UTC)

		parcels, err := card.GenerateInstallments(totalAmount, installmentsCount, purchaseDate, "Smartphone")
		require.NoError(t, err)
		require.Len(t, parcels, 3)

		// 100 / 3 = 33.33 each with 0.01 remainder on first installment: 33.34 + 33.33 + 33.33 = 100.00
		assert.True(t, parcels[0].Amount.Equal(decimal.NewFromFloat(33.34)))
		assert.Equal(t, 1, parcels[0].InstallmentNumber)
		assert.Equal(t, "2026-01", parcels[0].InvoiceMonth)

		assert.True(t, parcels[1].Amount.Equal(decimal.NewFromFloat(33.33)))
		assert.Equal(t, 2, parcels[1].InstallmentNumber)
		assert.Equal(t, "2026-02", parcels[1].InvoiceMonth)

		assert.True(t, parcels[2].Amount.Equal(decimal.NewFromFloat(33.33)))
		assert.Equal(t, 3, parcels[2].InstallmentNumber)
		assert.Equal(t, "2026-03", parcels[2].InvoiceMonth)

		// Verify total sum equals original
		sum := decimal.Zero
		for _, p := range parcels {
			sum = sum.Add(p.Amount)
		}
		assert.True(t, sum.Equal(totalAmount))
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
