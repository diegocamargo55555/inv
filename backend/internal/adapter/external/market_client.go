package external

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type cacheEntry struct {
	value     decimal.Decimal
	expiresAt time.Time
}

type MarketClient struct {
	cache        sync.Map
	httpClient   *http.Client
	brapiToken   string
	finnhubToken string
}

func NewMarketClient(brapiToken string, finnhubToken string) *MarketClient {
	return &MarketClient{
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

// GetQuote retrieves the asset quote with in-memory 5-minute caching
func (c *MarketClient) GetQuote(ctx context.Context, ticker string) (decimal.Decimal, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	cacheKey := fmt.Sprintf("quote:%s", ticker)

	// 1. Check in-memory cache first (cached for 5 minutes)
	if val, ok := c.cache.Load(cacheKey); ok {
		entry := val.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.value, nil
		}
		c.cache.Delete(cacheKey)
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
		return decimal.Zero, fetchErr
	}

	if !price.IsZero() {
		c.cache.Store(cacheKey, cacheEntry{value: price, expiresAt: time.Now().Add(5 * time.Minute)})
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
	from := strings.ToUpper(strings.TrimSpace(fromCurrency))
	to := strings.ToUpper(strings.TrimSpace(toCurrency))

	if from == "" || to == "" || strings.EqualFold(from, to) {
		return decimal.NewFromInt(1), nil
	}

	cacheKey := fmt.Sprintf("fx:%s_%s", from, to)
	lastKnownKey := fmt.Sprintf("fx:last_known:%s_%s", from, to)

	// 1. Check in-memory cache (5 minutes)
	if val, ok := c.cache.Load(cacheKey); ok {
		entry := val.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.value, nil
		}
		c.cache.Delete(cacheKey)
	}

	// 2. Fetch live rate from primary provider (AwesomeAPI)
	rate, err := c.fetchLiveRateAwesome(ctx, from, to)
	if err != nil || rate.IsZero() {
		// 3. Fallback to secondary provider (Frankfurter ECB rates)
		rate, err = c.fetchLiveRateFrankfurter(ctx, from, to)
	}

	if err == nil && !rate.IsZero() {
		c.cache.Store(cacheKey, cacheEntry{value: rate, expiresAt: time.Now().Add(5 * time.Minute)})
		c.cache.Store(lastKnownKey, cacheEntry{value: rate, expiresAt: time.Now().Add(100 * 365 * 24 * time.Hour)})
		return rate, nil
	}

	// 4. If all live providers fail, use last successfully recorded market rate from in-memory cache
	if val, ok := c.cache.Load(lastKnownKey); ok {
		entry := val.(cacheEntry)
		if entry.value.GreaterThan(decimal.Zero) {
			return entry.value, nil
		}
	}

	return decimal.Zero, err
}

func (c *MarketClient) fetchLiveRateAwesome(ctx context.Context, from, to string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://economia.awesomeapi.com.br/last/%s-%s", from, to)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return decimal.Zero, fmt.Errorf("awesomeapi status %d", resp.StatusCode)
	}

	var data map[string]struct {
		Bid string `json:"bid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return decimal.Zero, err
	}

	pairKey := fmt.Sprintf("%s%s", from, to)
	if item, ok := data[pairKey]; ok && item.Bid != "" {
		if d, err := decimal.NewFromString(item.Bid); err == nil && d.GreaterThan(decimal.Zero) {
			return d, nil
		}
	}
	return decimal.Zero, fmt.Errorf("cotação não encontrada no AwesomeAPI")
}

func (c *MarketClient) fetchLiveRateFrankfurter(ctx context.Context, from, to string) (decimal.Decimal, error) {
	url := fmt.Sprintf("https://api.frankfurter.dev/v1/latest?from=%s&to=%s", from, to)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return decimal.Zero, fmt.Errorf("frankfurter status %d", resp.StatusCode)
	}

	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return decimal.Zero, err
	}

	if val, ok := data.Rates[to]; ok && val > 0 {
		return decimal.NewFromFloat(val), nil
	}
	return decimal.Zero, fmt.Errorf("cotação não encontrada no Frankfurter")
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
