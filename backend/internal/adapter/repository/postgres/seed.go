package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedInitialData(db *gorm.DB) error {
	defaultCategories := []domain.Category{
		// Expense categories
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "Alimentação & Mercado", Type: domain.CategoryTypeExpense, Icon: "utensils", Color: "#EF4444", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "Moradia & Contas", Type: domain.CategoryTypeExpense, Icon: "home", Color: "#F59E0B", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Name: "Transporte & Combustível", Type: domain.CategoryTypeExpense, Icon: "car", Color: "#3B82F6", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004"), Name: "Saúde & Farmácia", Type: domain.CategoryTypeExpense, Icon: "heart-pulse", Color: "#10B981", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"), Name: "Lazer & Restaurantes", Type: domain.CategoryTypeExpense, Icon: "film", Color: "#8B5CF6", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000006"), Name: "Educação & Cursos", Type: domain.CategoryTypeExpense, Icon: "graduation-cap", Color: "#06B6D4", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000007"), Name: "Compras & Vestuário", Type: domain.CategoryTypeExpense, Icon: "shopping-bag", Color: "#EC4899", IsDefault: true, CreatedAt: time.Now()},
		// Income categories
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000008"), Name: "Salário / Remuneração", Type: domain.CategoryTypeIncome, Icon: "wallet", Color: "#10B981", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000009"), Name: "Dividendos & Rendimentos", Type: domain.CategoryTypeIncome, Icon: "trending-up", Color: "#059669", IsDefault: true, CreatedAt: time.Now()},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000010"), Name: "Freelance & Outros", Type: domain.CategoryTypeIncome, Icon: "briefcase", Color: "#6366F1", IsDefault: true, CreatedAt: time.Now()},
	}

	for _, cat := range defaultCategories {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&cat)
	}

	defaultAssets := []domain.Asset{
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000001"), Ticker: "PETR4", Name: "Petrobras PN", Type: domain.AssetTypeStock, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(38.25), Sector: "Petróleo e Gás", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000002"), Ticker: "VALE3", Name: "Vale ON", Type: domain.AssetTypeStock, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(61.50), Sector: "Mineração", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000003"), Ticker: "ITUB4", Name: "Itaú Unibanco PN", Type: domain.AssetTypeStock, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(35.80), Sector: "Financeiro", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000004"), Ticker: "BBAS3", Name: "Banco do Brasil ON", Type: domain.AssetTypeStock, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(27.40), Sector: "Financeiro", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000005"), Ticker: "WEGE3", Name: "WEG ON", Type: domain.AssetTypeStock, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(52.10), Sector: "Bens Industriais", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000006"), Ticker: "MXRF11", Name: "Maxi Renda FII", Type: domain.AssetTypeFII, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(10.15), Sector: "Títulos e Valores Mobiliários", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000007"), Ticker: "HGLG11", Name: "CSHG Logística FII", Type: domain.AssetTypeFII, Currency: "BRL", CurrentPrice: decimal.NewFromFloat(162.00), Sector: "Logística", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000008"), Ticker: "BTC", Name: "Bitcoin", Type: domain.AssetTypeCrypto, Currency: "USD", CurrentPrice: decimal.NewFromFloat(64500.00), Sector: "Criptoativo", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000009"), Ticker: "ETH", Name: "Ethereum", Type: domain.AssetTypeCrypto, Currency: "USD", CurrentPrice: decimal.NewFromFloat(2650.00), Sector: "Smart Contracts", UpdatedAt: time.Now()},
		{ID: uuid.MustParse("10000000-0000-0000-0000-000000000010"), Ticker: "AAPL", Name: "Apple Inc.", Type: domain.AssetTypeStock, Currency: "USD", CurrentPrice: decimal.NewFromFloat(225.00), Sector: "Tecnologia", UpdatedAt: time.Now()},
	}

	for _, asset := range defaultAssets {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&asset)
	}

	return nil
}
