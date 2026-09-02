package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPosition_CalculateWeightedAveragePrice(t *testing.T) {
	pos := &domain.Position{
		ID:           uuid.New(),
		PortfolioID:  uuid.New(),
		AssetID:      uuid.New(),
		Quantity:     decimal.Zero,
		AveragePrice: decimal.Zero,
		TotalCost:    decimal.Zero,
	}

	t.Run("first purchase sets initial average price", func(t *testing.T) {
		// Buy 100 shares at R$ 30.00 each with R$ 5.00 fees -> Total cost: 3005.00, PM: 30.05
		err := pos.ApplyBuy(decimal.NewFromFloat(100), decimal.NewFromFloat(30.00), decimal.NewFromFloat(5.00))
		require.NoError(t, err)

		assert.True(t, pos.Quantity.Equal(decimal.NewFromFloat(100)))
		assert.True(t, pos.TotalCost.Equal(decimal.NewFromFloat(3005.00)))
		assert.True(t, pos.AveragePrice.Equal(decimal.NewFromFloat(30.05)))
	})

	t.Run("second purchase updates weighted average price", func(t *testing.T) {
		// Buy 100 more shares at R$ 40.00 each with R$ 5.00 fees
		// New total cost: 3005.00 + 4005.00 = 7010.00
		// Total quantity: 200
		// New PM: 7010 / 200 = 35.05
		err := pos.ApplyBuy(decimal.NewFromFloat(100), decimal.NewFromFloat(40.00), decimal.NewFromFloat(5.00))
		require.NoError(t, err)

		assert.True(t, pos.Quantity.Equal(decimal.NewFromFloat(200)))
		assert.True(t, pos.TotalCost.Equal(decimal.NewFromFloat(7010.00)))
		assert.True(t, pos.AveragePrice.Equal(decimal.NewFromFloat(35.05)))
	})

	t.Run("partial sell decreases quantity without changing average price", func(t *testing.T) {
		// Sell 50 shares at R$ 50.00 with R$ 2.50 fees
		// PM remains 35.05
		// Remaining quantity: 150
		// Realized profit: (50 * 50.00) - (50 * 35.05) - 2.50 = 2500 - 1752.50 - 2.50 = 745.00
		realizedPnL, err := pos.ApplySell(decimal.NewFromFloat(50), decimal.NewFromFloat(50.00), decimal.NewFromFloat(2.50))
		require.NoError(t, err)

		assert.True(t, pos.Quantity.Equal(decimal.NewFromFloat(150)))
		assert.True(t, pos.AveragePrice.Equal(decimal.NewFromFloat(35.05)))
		assert.True(t, realizedPnL.Equal(decimal.NewFromFloat(745.00)))
	})

	t.Run("selling more quantity than owned returns error", func(t *testing.T) {
		_, err := pos.ApplySell(decimal.NewFromFloat(200), decimal.NewFromFloat(50.00), decimal.Zero)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInsufficientAssetQuantity, err)
	})

	t.Run("total sell leaves zero quantity and zero total cost", func(t *testing.T) {
		_, err := pos.ApplySell(decimal.NewFromFloat(150), decimal.NewFromFloat(35.05), decimal.Zero)
		require.NoError(t, err)

		assert.True(t, pos.Quantity.IsZero())
		assert.True(t, pos.TotalCost.IsZero())
		assert.True(t, pos.AveragePrice.IsZero())
	})
}

func TestPosition_YieldOnCostAndValuation(t *testing.T) {
	pos := &domain.Position{
		Quantity:     decimal.NewFromFloat(100),
		AveragePrice: decimal.NewFromFloat(20.00),
		TotalCost:    decimal.NewFromFloat(2000.00),
	}

	t.Run("calculate yield on cost correctly", func(t *testing.T) {
		totalDividendsPerShare := decimal.NewFromFloat(2.50) // R$ 2.50 per share
		yoc := pos.CalculateYieldOnCost(totalDividendsPerShare)

		// (2.50 / 20.00) * 100 = 12.5%
		assert.InDelta(t, 12.50, yoc, 0.01)
	})

	t.Run("valuation in BRL with USD exchange rate", func(t *testing.T) {
		usdPos := &domain.Position{
			Quantity:     decimal.NewFromFloat(10), // 10 shares of Apple (AAPL)
			AveragePrice: decimal.NewFromFloat(150.00),
			TotalCost:    decimal.NewFromFloat(1500.00), // in USD
		}

		currentPriceUSD := decimal.NewFromFloat(180.00)
		exchangeRateUSDBRL := decimal.NewFromFloat(5.50)

		currentValueBRL, unrealizedPnLBRL := usdPos.CalculateValuationWithFX(currentPriceUSD, exchangeRateUSDBRL)

		// 10 * 180 = $1800 USD -> 1800 * 5.50 = R$ 9900.00
		assert.True(t, currentValueBRL.Equal(decimal.NewFromFloat(9900.00)))

		// Unrealized PnL in USD = $1800 - $1500 = $300 -> 300 * 5.50 = R$ 1650.00
		assert.True(t, unrealizedPnLBRL.Equal(decimal.NewFromFloat(1650.00)))
	})
}
