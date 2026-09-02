package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InvestmentTxType string

const (
	InvestTxBuy          InvestmentTxType = "buy"
	InvestTxSell         InvestmentTxType = "sell"
	InvestTxSplit        InvestmentTxType = "split"        // Desdobramento
	InvestTxGroup        InvestmentTxType = "group"        // Grupamento
	InvestTxAmortization InvestmentTxType = "amortization" // Amortização
)

type InvestmentTransaction struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	PortfolioID uuid.UUID        `json:"portfolio_id" gorm:"type:uuid;index;not null"`
	AssetID     uuid.UUID        `json:"asset_id" gorm:"type:uuid;index;not null"`
	AccountID   *uuid.UUID       `json:"account_id,omitempty" gorm:"type:uuid"` // Conta de onde saiu/entrou o dinheiro
	Type        InvestmentTxType `json:"type" gorm:"type:varchar(20);not null"`
	Quantity    decimal.Decimal  `json:"quantity" gorm:"type:numeric(18,8);not null"`
	UnitPrice   decimal.Decimal  `json:"unit_price" gorm:"type:numeric(18,4);not null"`
	Fees        decimal.Decimal  `json:"fees" gorm:"type:numeric(18,4);default:0"`
	TotalAmount decimal.Decimal  `json:"total_amount" gorm:"type:numeric(18,4);not null"`
	RealizedPnL decimal.Decimal  `json:"realized_pnl" gorm:"type:numeric(18,4);default:0"` // Lucro/prejuízo apurado na venda
	Date        time.Time        `json:"date" gorm:"not null"`
	Notes       string           `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time        `json:"created_at"`

	Asset *Asset `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
}
