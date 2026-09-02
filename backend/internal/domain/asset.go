package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AssetType string

const (
	AssetTypeStock       AssetType = "stock"        // Ações B3 / US
	AssetTypeFII         AssetType = "fii"          // Fundos Imobiliários
	AssetTypeBDR         AssetType = "bdr"          // BDRs
	AssetTypeETF         AssetType = "etf"          // ETFs
	AssetTypeFixedIncome AssetType = "fixed_income" // Tesouro Direto, CDB, LCI, LCA
	AssetTypeCrypto      AssetType = "crypto"       // Bitcoin, Ethereum, etc.
	AssetTypeCurrency    AssetType = "currency"     // USD, EUR
)

type Asset struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	Ticker       string          `json:"ticker" gorm:"type:varchar(20);uniqueIndex;not null"`
	Name         string          `json:"name" gorm:"type:varchar(150);not null"`
	Type         AssetType       `json:"type" gorm:"type:varchar(30);not null"`
	Currency     string          `json:"currency" gorm:"type:varchar(3);default:'BRL'"` // BRL, USD, etc.
	CurrentPrice decimal.Decimal `json:"current_price" gorm:"type:numeric(18,4);default:0"`
	Sector       string          `json:"sector" gorm:"type:varchar(100)"`
	CNPJ         string          `json:"cnpj" gorm:"type:varchar(20)"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
