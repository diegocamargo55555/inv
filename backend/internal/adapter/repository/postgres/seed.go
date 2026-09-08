package postgres

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedInitialData(db *gorm.DB) error {
	now := time.Now()

	categories := []struct {
		name, icon, color string
		catType           domain.CategoryType
	}{
		{"Alimentação & Mercado", "utensils", "#EF4444", domain.CategoryTypeExpense},
		{"Moradia & Contas", "home", "#F59E0B", domain.CategoryTypeExpense},
		{"Transporte & Combustível", "car", "#3B82F6", domain.CategoryTypeExpense},
		{"Saúde & Farmácia", "heart-pulse", "#10B981", domain.CategoryTypeExpense},
		{"Lazer & Restaurantes", "film", "#8B5CF6", domain.CategoryTypeExpense},
		{"Educação & Cursos", "graduation-cap", "#06B6D4", domain.CategoryTypeExpense},
		{"Compras & Vestuário", "shopping-bag", "#EC4899", domain.CategoryTypeExpense},
		{"Salário / Remuneração", "wallet", "#10B981", domain.CategoryTypeIncome},
		{"Dividendos & Rendimentos", "trending-up", "#059669", domain.CategoryTypeIncome},
		{"Freelance & Outros", "briefcase", "#6366F1", domain.CategoryTypeIncome},
	}

	for i, c := range categories {
		cat := domain.Category{
			ID:        uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i+1)),
			Name:      c.name,
			Type:      c.catType,
			Icon:      c.icon,
			Color:     c.color,
			IsDefault: true,
			CreatedAt: now,
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&cat)
	}

	assets := []struct {
		ticker, name, currency, sector string
		price                          float64
		assetType                      domain.AssetType
	}{
		{"PETR4", "Petrobras PN", "BRL", "Petróleo e Gás", 38.25, domain.AssetTypeStock},
		{"VALE3", "Vale ON", "BRL", "Mineração", 61.50, domain.AssetTypeStock},
		{"ITUB4", "Itaú Unibanco PN", "BRL", "Financeiro", 35.80, domain.AssetTypeStock},
		{"BBAS3", "Banco do Brasil ON", "BRL", "Financeiro", 27.40, domain.AssetTypeStock},
		{"WEGE3", "WEG ON", "BRL", "Bens Industriais", 52.10, domain.AssetTypeStock},
		{"MXRF11", "Maxi Renda FII", "BRL", "Títulos e Valores Mobiliários", 10.15, domain.AssetTypeFII},
		{"HGLG11", "CSHG Logística FII", "BRL", "Logística", 162.00, domain.AssetTypeFII},
		{"BTC", "Bitcoin", "USD", "Criptoativo", 64500.00, domain.AssetTypeCrypto},
		{"ETH", "Ethereum", "USD", "Smart Contracts", 2650.00, domain.AssetTypeCrypto},
		{"AAPL", "Apple Inc.", "USD", "Tecnologia", 225.00, domain.AssetTypeStock},
	}

	for i, a := range assets {
		asset := domain.Asset{
			ID:           uuid.MustParse(fmt.Sprintf("10000000-0000-0000-0000-%012d", i+1)),
			Ticker:       a.ticker,
			Name:         a.name,
			Type:         a.assetType,
			Currency:     a.currency,
			CurrentPrice: decimal.NewFromFloat(a.price),
			Sector:       a.sector,
			UpdatedAt:    now,
		}
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&asset)
	}

	return nil
}
