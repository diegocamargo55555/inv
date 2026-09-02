package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AccountType string

const (
	AccountTypeChecking   AccountType = "checking"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeInvestment AccountType = "investment"
	AccountTypeCash       AccountType = "cash"
	AccountTypeWallet     AccountType = "wallet"
)

type Account struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID       `json:"user_id" gorm:"type:uuid;index;not null"`
	Name        string          `json:"name" gorm:"type:varchar(100);not null"`
	Type        AccountType     `json:"type" gorm:"type:varchar(30);not null"`
	Balance     decimal.Decimal `json:"balance" gorm:"type:numeric(18,4);not null;default:0"`
	Currency    string          `json:"currency" gorm:"type:varchar(3);default:'BRL'"`
	Color       string          `json:"color" gorm:"type:varchar(20);default:'#10B981'"`
	Institution string          `json:"institution" gorm:"type:varchar(100)"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ApplyTransaction updates the account balance based on transaction type
func (a *Account) ApplyTransaction(amount decimal.Decimal, txType TransactionType) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	switch txType {
	case TxTypeIncome:
		a.Balance = a.Balance.Add(amount)
	case TxTypeExpense:
		a.Balance = a.Balance.Sub(amount)
	default:
		// Transfers are handled specifically in use case
	}
	return nil
}
