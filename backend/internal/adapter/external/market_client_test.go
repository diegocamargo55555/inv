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
	client := external.NewMarketClient(nil, "", "")
	ctx := context.Background()

	t.Run("get quote with offline fallback for PETR4 (B3)", func(t *testing.T) {
		quote, err := client.GetQuote(ctx, "PETR4")
		require.NoError(t, err)
		assert.True(t, quote.GreaterThan(decimal.Zero))
	})

	t.Run("get quote with offline fallback for AAPL (US Stock)", func(t *testing.T) {
		quote, err := client.GetQuote(ctx, "AAPL")
		require.NoError(t, err)
		assert.True(t, quote.GreaterThan(decimal.Zero))
	})

	t.Run("get quote with offline fallback for VOO (US ETF)", func(t *testing.T) {
		quote, err := client.GetQuote(ctx, "VOO")
		require.NoError(t, err)
		assert.True(t, quote.GreaterThan(decimal.Zero))
	})

	t.Run("get quote with offline fallback for BTC (Crypto)", func(t *testing.T) {
		quote, err := client.GetQuote(ctx, "BTC")
		require.NoError(t, err)
		assert.True(t, quote.GreaterThan(decimal.Zero))
	})

	t.Run("get exchange rate for same currency returns 1", func(t *testing.T) {
		rate, err := client.GetExchangeRate(ctx, "BRL", "BRL")
		require.NoError(t, err)
		assert.True(t, rate.Equal(decimal.NewFromInt(1)))
	})

	t.Run("get USD to BRL exchange rate returns positive rate", func(t *testing.T) {
		rate, err := client.GetExchangeRate(ctx, "USD", "BRL")
		require.NoError(t, err)
		assert.True(t, rate.GreaterThan(decimal.Zero))
	})
}
