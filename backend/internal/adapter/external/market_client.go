package external

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/invest/backend/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type MarketClient struct {
	redisClient  *redis.Client
	httpClient   *http.Client
	brapiToken   string
	finnhubToken string
}

func NewMarketClient(redisClient *redis.Client, brapiToken string, finnhubToken string) *MarketClient {
	return &MarketClient{
		redisClient:  redisClient,
		httpClient:   &http.Client{Timeout: 8 * time.Second},
		brapiToken:   brapiToken,
		finnhubToken: finnhubToken,
	}
}

// isUSTicker checks if ticker represents a US stock/ETF (typically 1-5 alphabetic chars without numeric suffix)
func isUSTicker(ticker string) bool {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if len(ticker) == 0 || len(ticker) > 5 {
		return false
	}
	// If ends with a digit, it's almost certainly a Brazilian B3 ticker (e.g. PETR4, VALE3, MXRF11)
	if unicode.IsDigit(rune(ticker[len(ticker)-1])) {
		return false
	}
	for _, r := range ticker {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// GetQuote retrieves the asset quote from Brapi (B3), Finnhub (US NYSE/NASDAQ), or CoinGecko (Crypto), with 5-min Redis caching
func (c *MarketClient) GetQuote(ctx context.Context, ticker string) (decimal.Decimal, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	cacheKey := fmt.Sprintf("quote:%s", ticker)

	// 1. Check Redis cache first (cached for 5 minutes)
	if c.redisClient != nil {
		cachedPrice, err := c.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedPrice != "" {
			if d, err := decimal.NewFromString(cachedPrice); err == nil && d.GreaterThan(decimal.Zero) {
				return d, nil
			}
		}
	}

	// 2. Fetch directly from external API (Finnhub for US, Brapi for B3, CoinGecko for Crypto)
	var price decimal.Decimal
	var fetchErr error

	if strings.EqualFold(ticker, "BTC") || strings.EqualFold(ticker, "ETH") || strings.EqualFold(ticker, "SOL") {
		price, fetchErr = c.fetchCryptoQuote(ticker)
	} else if isUSTicker(ticker) && c.finnhubToken != "" {
		// Primary US provider: Finnhub
		price, fetchErr = c.fetchFinnhubQuote(ticker)
		if fetchErr != nil {
			log.Printf("[MarketClient] Failed to fetch US quote for %s from Finnhub: %v. Trying Brapi fallback.", ticker, fetchErr)
			var brapiErr error
			price, brapiErr = c.fetchBrapiQuote(ticker)
			if brapiErr == nil && !price.IsZero() {
				fetchErr = nil
			}
		}
	} else {
		// Primary B3 provider: Brapi
		price, fetchErr = c.fetchBrapiQuote(ticker)
		if fetchErr != nil && isUSTicker(ticker) && c.finnhubToken != "" {
			var finnErr error
			price, finnErr = c.fetchFinnhubQuote(ticker)
			if finnErr == nil && !price.IsZero() {
				fetchErr = nil
			}
		}
	}

	if fetchErr != nil {
		log.Printf("[MarketClient] Failed to fetch live quote for %s: %v. Using offline fallback if available.", ticker, fetchErr)
		// Fallback default prices for common assets during offline/dev mode or quota limits
		fallbackPrices := map[string]float64{
			// Brazilian B3 Stocks & FIIs
			"PETR4":  38.25,
			"VALE3":  61.50,
			"ITUB4":  35.80,
			"BBAS3":  27.40,
			"WEGE3":  52.10,
			"MXRF11": 10.15,
			"HGLG11": 162.00,
			// US NYSE / NASDAQ Stocks & ETFs
			"AAPL":  225.00,
			"MSFT":  420.00,
			"NVDA":  120.00,
			"TSLA":  215.00,
			"GOOGL": 165.00,
			"AMZN":  175.00,
			"META":  510.00,
			"VOO":   510.00,
			"SPY":   560.00,
			"QQQ":   480.00,
			// Crypto
			"BTC": 64500.00,
			"ETH": 2650.00,
			"SOL": 145.00,
		}
		if p, ok := fallbackPrices[ticker]; ok {
			price = decimal.NewFromFloat(p)
		} else {
			return decimal.Zero, fetchErr
		}
	}

	// Cache result in Redis for 5 minutes (as requested)
	if c.redisClient != nil && !price.IsZero() {
		_ = c.redisClient.Set(ctx, cacheKey, price.String(), 5*time.Minute).Err()
	}

	return price, nil
}

type brapiStockItem struct {
	Stock   string  `json:"stock"`
	Name    string  `json:"name"`
	Close   float64 `json:"close"`
	Type    string  `json:"type"`
	SubType string  `json:"subType"`
	Logo    string  `json:"logo"`
	Sector  string  `json:"sector"`
}

type brapiListResponse struct {
	Stocks []brapiStockItem `json:"stocks"`
}

type finnhubSearchItem struct {
	Description   string `json:"description"`
	DisplaySymbol string `json:"displaySymbol"`
	Symbol        string `json:"symbol"`
	Type          string `json:"type"`
}

type finnhubSearchResponse struct {
	Count  int                 `json:"count"`
	Result []finnhubSearchItem `json:"result"`
}

// SearchAssets searches for matching B3 tickers (via Brapi) and US tickers (via Finnhub)
func (c *MarketClient) SearchAssets(ctx context.Context, query string) ([]domain.Asset, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	var assets []domain.Asset
	seen := make(map[string]bool)

	// 1. Search B3 via Brapi
	brapiUrl := fmt.Sprintf("https://brapi.dev/api/quote/list?search=%s&limit=10", query)
	if c.brapiToken != "" {
		brapiUrl = fmt.Sprintf("%s&token=%s", brapiUrl, c.brapiToken)
	}

	reqBrapi, err := http.NewRequestWithContext(ctx, http.MethodGet, brapiUrl, nil)
	if err == nil {
		reqBrapi.Header.Set("User-Agent", "CapitalHub-Fintech/1.0")
		if respBrapi, err := c.httpClient.Do(reqBrapi); err == nil && respBrapi.StatusCode == http.StatusOK {
			defer respBrapi.Body.Close()
			var data brapiListResponse
			if err := json.NewDecoder(respBrapi.Body).Decode(&data); err == nil {
				for _, s := range data.Stocks {
					ticker := strings.ToUpper(s.Stock)
					if seen[ticker] {
						continue
					}
					seen[ticker] = true

					assetType := domain.AssetTypeStock
					if strings.EqualFold(s.Type, "fund") || strings.EqualFold(s.SubType, "fii") || strings.HasSuffix(s.Stock, "11") {
						assetType = domain.AssetTypeFII
					}

					assets = append(assets, domain.Asset{
						Ticker:       ticker,
						Name:         s.Name,
						Type:         assetType,
						Currency:     "BRL",
						CurrentPrice: decimal.NewFromFloat(s.Close),
					})
				}
			}
		}
	}

	// 2. Search US Stocks via Finnhub (if token configured)
	if c.finnhubToken != "" {
		finnhubUrl := fmt.Sprintf("https://finnhub.io/api/v1/search?q=%s&token=%s", query, c.finnhubToken)
		reqFinn, err := http.NewRequestWithContext(ctx, http.MethodGet, finnhubUrl, nil)
		if err == nil {
			reqFinn.Header.Set("User-Agent", "CapitalHub-Fintech/1.0")
			if respFinn, err := c.httpClient.Do(reqFinn); err == nil && respFinn.StatusCode == http.StatusOK {
				defer respFinn.Body.Close()
				var data finnhubSearchResponse
				if err := json.NewDecoder(respFinn.Body).Decode(&data); err == nil {
					for _, item := range data.Result {
						symbol := strings.ToUpper(item.Symbol)
						if strings.Contains(symbol, ".") || seen[symbol] {
							continue
						}
						seen[symbol] = true

						assets = append(assets, domain.Asset{
							Ticker:   symbol,
							Name:     item.Description,
							Type:     domain.AssetTypeStock,
							Currency: "USD",
						})
					}
				}
			}
		}
	}

	return assets, nil
}

func (c *MarketClient) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	if strings.EqualFold(fromCurrency, toCurrency) {
		return decimal.NewFromInt(1), nil
	}

	cacheKey := fmt.Sprintf("fx:%s_%s", strings.ToUpper(fromCurrency), strings.ToUpper(toCurrency))
	if c.redisClient != nil {
		if cached, err := c.redisClient.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			if d, err := decimal.NewFromString(cached); err == nil {
				return d, nil
			}
		}
	}

	// Exchange rate USD -> BRL (default 5.50)
	rate := decimal.NewFromFloat(5.50)
	if c.redisClient != nil {
		_ = c.redisClient.Set(ctx, cacheKey, rate.String(), 1*time.Hour).Err()
	}

	return rate, nil
}

type finnhubQuoteResponse struct {
	CurrentPrice  float64 `json:"c"`
	High          float64 `json:"h"`
	Low           float64 `json:"l"`
	Open          float64 `json:"o"`
	PreviousClose float64 `json:"pc"`
	Timestamp     int64   `json:"t"`
}

// fetchFinnhubQuote gets real-time US stock/ETF quotes from Finnhub (60 req/min free limit)
func (c *MarketClient) fetchFinnhubQuote(ticker string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s", ticker)
	if c.finnhubToken != "" {
		url = fmt.Sprintf("%s&token=%s", url, c.finnhubToken)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, err
	}
	req.Header.Set("User-Agent", "CapitalHub-Fintech/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("finnhub returned status %d for %s", resp.StatusCode, ticker)
	}

	var data finnhubQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return decimal.Zero, err
	}

	if data.CurrentPrice <= 0 {
		return decimal.Zero, fmt.Errorf("no valid price in finnhub for %s", ticker)
	}

	return decimal.NewFromFloat(data.CurrentPrice), nil
}

type brapiResponse struct {
	Results []struct {
		Symbol             string  `json:"symbol"`
		ShortName          string  `json:"shortName"`
		LongName           string  `json:"longName"`
		Currency           string  `json:"currency"`
		RegularMarketPrice float64 `json:"regularMarketPrice"`
	} `json:"results"`
	Error bool `json:"error"`
}

func (c *MarketClient) fetchBrapiQuote(ticker string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://brapi.dev/api/quote/%s", ticker)
	if c.brapiToken != "" {
		url = fmt.Sprintf("%s?token=%s", url, c.brapiToken)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, err
	}
	req.Header.Set("User-Agent", "CapitalHub-Fintech/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("brapi returned status %d for %s", resp.StatusCode, ticker)
	}

	var data brapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return decimal.Zero, err
	}

	if data.Error || len(data.Results) == 0 || data.Results[0].RegularMarketPrice <= 0 {
		return decimal.Zero, fmt.Errorf("no quote found in brapi for %s", ticker)
	}

	return decimal.NewFromFloat(data.Results[0].RegularMarketPrice), nil
}

func (c *MarketClient) fetchCryptoQuote(ticker string) (decimal.Decimal, error) {
	id := "bitcoin"
	if strings.EqualFold(ticker, "ETH") {
		id = "ethereum"
	} else if strings.EqualFold(ticker, "SOL") {
		id = "solana"
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, err
	}
	req.Header.Set("User-Agent", "CapitalHub-Fintech/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("coingecko returned status %d", resp.StatusCode)
	}

	var data map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return decimal.Zero, err
	}

	if val, ok := data[id]["usd"]; ok && val > 0 {
		return decimal.NewFromFloat(val), nil
	}

	return decimal.Zero, fmt.Errorf("crypto price not found for %s", ticker)
}
