package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Portfolio struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;index;not null"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null"`
	Description string    `json:"description" gorm:"type:varchar(255)"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Positions []Position `json:"positions,omitempty" gorm:"foreignKey:PortfolioID"`
}

type Position struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	PortfolioID  uuid.UUID       `json:"portfolio_id" gorm:"type:uuid;index;not null"`
	AssetID      uuid.UUID       `json:"asset_id" gorm:"type:uuid;index;not null"`
	Quantity     decimal.Decimal `json:"quantity" gorm:"type:numeric(18,8);not null;default:0"`
	AveragePrice decimal.Decimal `json:"average_price" gorm:"type:numeric(18,4);not null;default:0"`
	TotalCost    decimal.Decimal `json:"total_cost" gorm:"type:numeric(18,4);not null;default:0"`
	UpdatedAt    time.Time       `json:"updated_at"`

	Asset *Asset `json:"asset,omitempty" gorm:"foreignKey:AssetID"`
}

// ApplyBuy calculates the weighted average price and updates position quantity and total cost
func (p *Position) ApplyBuy(quantity, price, fees decimal.Decimal) error {
	if quantity.LessThanOrEqual(decimal.Zero) || price.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	buyCost := quantity.Mul(price).Add(fees)
	newTotalCost := p.TotalCost.Add(buyCost)
	newQuantity := p.Quantity.Add(quantity)

	p.Quantity = newQuantity
	p.TotalCost = newTotalCost
	p.AveragePrice = newTotalCost.Div(newQuantity).Round(4)
	return nil
}

// ApplySell updates position quantity upon selling and computes realized profit or loss
// Note: In Brazilian B3 / IRS standard, selling does NOT change average purchase price.
func (p *Position) ApplySell(quantity, price, fees decimal.Decimal) (decimal.Decimal, error) {
	if quantity.LessThanOrEqual(decimal.Zero) || price.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrInvalidAmount
	}
	if quantity.GreaterThan(p.Quantity) {
		return decimal.Zero, ErrInsufficientAssetQuantity
	}

	// Realized PnL = (quantity * sell_price) - (quantity * average_price) - fees
	grossRevenue := quantity.Mul(price)
	costBasis := quantity.Mul(p.AveragePrice)
	realizedPnL := grossRevenue.Sub(costBasis).Sub(fees)

	p.Quantity = p.Quantity.Sub(quantity)
	if p.Quantity.IsZero() {
		p.TotalCost = decimal.Zero
		p.AveragePrice = decimal.Zero
	} else {
		p.TotalCost = p.Quantity.Mul(p.AveragePrice).Round(4)
	}

	return realizedPnL, nil
}

// CalculateYieldOnCost calculates the dividend yield on cost percentage
func (p *Position) CalculateYieldOnCost(totalDividendsPerShare decimal.Decimal) float64 {
	if p.AveragePrice.IsZero() {
		return 0
	}
	yocDecimal := totalDividendsPerShare.Div(p.AveragePrice).Mul(decimal.NewFromInt(100))
	yocFloat, _ := yocDecimal.Float64()
	return yocFloat
}

// CalculateValuationWithFX calculates current valuation and unrealized gain/loss converted to BRL
func (p *Position) CalculateValuationWithFX(currentPrice, exchangeRate decimal.Decimal) (currentValueBRL, unrealizedPnLBRL decimal.Decimal) {
	if exchangeRate.IsZero() {
		exchangeRate = decimal.NewFromInt(1)
	}

	currentValueNative := p.Quantity.Mul(currentPrice)
	currentValueBRL = currentValueNative.Mul(exchangeRate).Round(4)

	unrealizedGainNative := currentValueNative.Sub(p.TotalCost)
	unrealizedPnLBRL = unrealizedGainNative.Mul(exchangeRate).Round(4)

	return currentValueBRL, unrealizedPnLBRL
}
