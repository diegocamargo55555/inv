package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EarningType string

const (
	EarningTypeDividend EarningType = "dividend" // Dividendos (Isentos)
	EarningTypeJCP      EarningType = "jcp"      // Juros sobre Capital Próprio (Tributado)
	EarningTypeYield    EarningType = "yield"    // Rendimentos FIIs (Isentos)
	EarningTypeInterest EarningType = "interest" // Cupons de Renda Fixa
)

type Earning struct {
	ID             uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	PortfolioID    uuid.UUID       `json:"portfolio_id" gorm:"type:uuid;index;not null"`
	AssetID        uuid.UUID       `json:"asset_id" gorm:"type:uuid;index;not null"`
	AccountID      *uuid.UUID      `json:"account_id,omitempty" gorm:"type:uuid"` // Conta de crédito do provento
	Type           EarningType     `json:"type" gorm:"type:varchar(20);not null"`
	DateCom        time.Time       `json:"date_com" gorm:"not null"`     // Data base de corte
	PaymentDate    time.Time       `json:"payment_date" gorm:"not null"` // Data do pagamento efetivo
	ValuePerShare  decimal.Decimal `json:"value_per_share" gorm:"type:numeric(18,6);not null"`
	TotalAmount    decimal.Decimal `json:"total_amount" gorm:"type:numeric(18,4);not null"`
	NetTotalAmount decimal.Decimal `json:"net_total_amount" gorm:"type:numeric(18,4);not null"` // Líquido após IR (se JCP)
	CreatedAt      time.Time       `json:"created_at"`

	Asset *Asset `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
}
