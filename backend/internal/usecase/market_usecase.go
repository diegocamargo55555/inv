package usecase

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type MarketUseCase struct {
	assetRepo  AssetRepository
	marketData MarketDataProvider
}

func NewMarketUseCase(assetRepo AssetRepository, marketData MarketDataProvider) *MarketUseCase {
	return &MarketUseCase{
		assetRepo:  assetRepo,
		marketData: marketData,
	}
}

// SearchAssets searches for assets in repository and queries Brapi for live B3 tickers
func (uc *MarketUseCase) SearchAssets(ctx context.Context, query string) ([]domain.Asset, error) {
	query = strings.TrimSpace(query)

	// 1. Search in local database
	dbAssets, err := uc.assetRepo.Search(ctx, query)
	if err != nil {
		dbAssets = []domain.Asset{}
	}

	// 2. If query is provided, query Brapi list endpoint
	if query != "" && uc.marketData != nil {
		if searcher, ok := uc.marketData.(interface {
			SearchAssets(ctx context.Context, query string) ([]domain.Asset, error)
		}); ok {
			brapiAssets, err := searcher.SearchAssets(ctx, query)
			if err == nil && len(brapiAssets) > 0 {
				seen := make(map[string]bool)
				var merged []domain.Asset

				for _, a := range dbAssets {
					seen[a.Ticker] = true
					merged = append(merged, a)
				}

				for _, bAsset := range brapiAssets {
					if !seen[bAsset.Ticker] {
						seen[bAsset.Ticker] = true
						bAsset.ID = uuid.New()
						_ = uc.assetRepo.CreateOrUpdate(ctx, &bAsset)
						merged = append(merged, bAsset)
					}
				}

				return merged, nil
			}
		}
	}

	return dbAssets, nil
}

// GetOrCreateAsset searches by ticker or registers a new asset with automatic quote lookup
func (uc *MarketUseCase) GetOrCreateAsset(ctx context.Context, ticker, name string, assetType domain.AssetType, currency string) (*domain.Asset, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return nil, domain.ErrInvalidAssetTicker
	}

	existing, err := uc.assetRepo.GetByTicker(ctx, ticker)
	if err == nil && existing != nil {
		return existing, nil
	}

	if assetType == "" {
		if strings.HasSuffix(ticker, "11") {
			assetType = domain.AssetTypeFII
		} else if strings.Contains("BTC ETH SOL USDT BNB", ticker) {
			assetType = domain.AssetTypeCrypto
		} else {
			assetType = domain.AssetTypeStock
		}
	}

	if currency == "" {
		if assetType == domain.AssetTypeCrypto {
			currency = "USD"
		} else {
			currency = "BRL"
		}
	}

	if name == "" {
		name = ticker
	}

	var quote decimal.Decimal
	if uc.marketData != nil {
		q, err := uc.marketData.GetQuote(ctx, ticker)
		if err == nil {
			quote = q
		}
	}

	asset := &domain.Asset{
		ID:           uuid.New(),
		Ticker:       ticker,
		Name:         name,
		Type:         assetType,
		Currency:     currency,
		CurrentPrice: quote,
	}

	if err := uc.assetRepo.CreateOrUpdate(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// GetLiveQuote fetches real-time quote directly from provider (Brapi)
func (uc *MarketUseCase) GetLiveQuote(ctx context.Context, ticker string) (decimal.Decimal, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))
	if ticker == "" {
		return decimal.Zero, domain.ErrInvalidAssetTicker
	}
	if uc.marketData == nil {
		return decimal.Zero, domain.ErrAssetNotFound
	}
	return uc.marketData.GetQuote(ctx, ticker)
}

// UpdateAssetPrice fetches the latest market quote and updates the asset record
func (uc *MarketUseCase) UpdateAssetPrice(ctx context.Context, assetID uuid.UUID) (decimal.Decimal, error) {
	asset, err := uc.assetRepo.GetByID(ctx, assetID)
	if err != nil || asset == nil {
		return decimal.Zero, domain.ErrAssetNotFound
	}

	if uc.marketData == nil {
		return asset.CurrentPrice, nil
	}

	quote, err := uc.marketData.GetQuote(ctx, asset.Ticker)
	if err != nil {
		log.Printf("Failed to fetch quote for ticker %s: %v", asset.Ticker, err)
		return asset.CurrentPrice, err
	}

	if err := uc.assetRepo.UpdatePrice(ctx, assetID, quote); err != nil {
		return decimal.Zero, err
	}

	return quote, nil
}
