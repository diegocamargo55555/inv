package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Budget struct {
	ID          uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID       `json:"user_id" gorm:"type:uuid;index;not null"`
	CategoryID  *uuid.UUID      `json:"category_id,omitempty" gorm:"type:uuid;index"`
	MonthYear   string          `json:"month_year" gorm:"type:varchar(7);not null"` // e.g. "2026-03"
	AmountLimit decimal.Decimal `json:"amount_limit" gorm:"type:numeric(18,4);not null"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Category *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
}

// CalculateProgress calculates the consumption percentage and alerts
func (b *Budget) CalculateProgress(spent decimal.Decimal) (percentage float64, isAlert bool, isExceeded bool) {
	if b.AmountLimit.IsZero() {
		return 0, false, false
	}

	pctDecimal := spent.Div(b.AmountLimit).Mul(decimal.NewFromInt(100))
	pctFloat, _ := pctDecimal.Float64()

	isAlert = pctFloat >= 80.0
	isExceeded = pctFloat > 100.0

	return pctFloat, isAlert, isExceeded
}
