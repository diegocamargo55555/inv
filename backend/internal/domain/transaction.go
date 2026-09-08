package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TxTypeIncome   TransactionType = "income"
	TxTypeExpense  TransactionType = "expense"
	TxTypeTransfer TransactionType = "transfer"
)

type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusCompleted TransactionStatus = "completed"
)

type Transaction struct {
	ID                  uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey"`
	UserID              uuid.UUID         `json:"user_id" gorm:"type:uuid;index;not null"`
	AccountID        *uuid.UUID        `json:"account_id,omitempty" gorm:"type:uuid;index"`
	DestinationAccID *uuid.UUID        `json:"destination_account_id,omitempty" gorm:"type:uuid"`
	CategoryID       *uuid.UUID        `json:"category_id,omitempty" gorm:"type:uuid;index"`
	Type             TransactionType   `json:"type" gorm:"type:varchar(20);not null"`
	Status           TransactionStatus `json:"status" gorm:"type:varchar(20);default:'completed'"`
	Amount           decimal.Decimal   `json:"amount" gorm:"type:numeric(18,4);not null"`
	Date             time.Time         `json:"date" gorm:"not null"`
	Description      string            `json:"description" gorm:"type:varchar(255);not null"`
	Notes            string            `json:"notes" gorm:"type:text"`
	Tags             string            `json:"tags" gorm:"type:varchar(255)"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`

	Account  *Account  `json:"account,omitempty" gorm:"foreignKey:AccountID"`
	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}
