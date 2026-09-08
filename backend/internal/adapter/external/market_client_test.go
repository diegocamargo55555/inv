package external_test

import (
	"context"
	"testing"

	"github.com/invest/backend/internal/adapter/external"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketClient_QuotesAndFX(t *testing.T) {
	client := external.NewMarketClient("", "")
	ctx := context.Background()

	t.Run("get exchange rate for same currency returns 1", func(t *testing.T) {
		rate, err := client.GetExchangeRate(ctx, "BRL", "BRL")
		require.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromInt(1)))
	})

	t.Run("invalid ticker returns error without provider", func(t *testing.T) {
		_, err := client.GetQuote(ctx, "NONEXISTENT_TICKER_9999")
		assert.Error(t, err)
	})
}
