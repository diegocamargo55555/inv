package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormPortfolioRepo struct {
	db *gorm.DB
}

func NewPortfolioRepo(db *gorm.DB) *GormPortfolioRepo {
	return &GormPortfolioRepo{db: db}
}

func (r *GormPortfolioRepo) Create(ctx context.Context, p *domain.Portfolio) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *GormPortfolioRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Portfolio, error) {
	var p domain.Portfolio
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *GormPortfolioRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Portfolio, error) {
	var list []domain.Portfolio
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default DESC, name ASC").Find(&list).Error
	return list, err
}

func (r *GormPortfolioRepo) GetPosition(ctx context.Context, portfolioID, assetID uuid.UUID) (*domain.Position, error) {
	var pos domain.Position
	err := r.db.WithContext(ctx).Preload("Asset").First(&pos, "portfolio_id = ? AND asset_id = ?", portfolioID, assetID).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *GormPortfolioRepo) SavePosition(ctx context.Context, pos *domain.Position) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(pos).Error
}

func (r *GormPortfolioRepo) GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]domain.Position, error) {
	var list []domain.Position
	err := r.db.WithContext(ctx).Where("portfolio_id = ?", portfolioID).Preload("Asset").Order("total_cost DESC").Find(&list).Error
	return list, err
}

// Asset Repository
type GormAssetRepo struct {
	db *gorm.DB
}

func NewAssetRepo(db *gorm.DB) *GormAssetRepo {
	return &GormAssetRepo{db: db}
}

func (r *GormAssetRepo) CreateOrUpdate(ctx context.Context, asset *domain.Asset) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ticker"}},
		UpdateAll: true,
	}).Create(asset).Error
}

func (r *GormAssetRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Asset, error) {
	var asset domain.Asset
	if err := r.db.WithContext(ctx).First(&asset, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *GormAssetRepo) GetByTicker(ctx context.Context, ticker string) (*domain.Asset, error) {
	var asset domain.Asset
	if err := r.db.WithContext(ctx).First(&asset, "ticker = ?", ticker).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *GormAssetRepo) Search(ctx context.Context, query string) ([]domain.Asset, error) {
	var list []domain.Asset
	searchTerm := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("ticker ILIKE ? OR name ILIKE ?", searchTerm, searchTerm).
		Limit(20).
		Find(&list).Error
	return list, err
}

func (r *GormAssetRepo) UpdatePrice(ctx context.Context, assetID uuid.UUID, price decimal.Decimal) error {
	return r.db.WithContext(ctx).Model(&domain.Asset{}).Where("id = ?", assetID).Updates(map[string]interface{}{
		"current_price": price,
		"updated_at":    time.Now(),
	}).Error
}

// InvestmentTx Repository
type GormInvestTxRepo struct {
	db *gorm.DB
}

func NewInvestTxRepo(db *gorm.DB) *GormInvestTxRepo {
	return &GormInvestTxRepo{db: db}
}

func (r *GormInvestTxRepo) Create(ctx context.Context, tx *domain.InvestmentTransaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *GormInvestTxRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentTransaction, error) {
	var tx domain.InvestmentTransaction
	err := r.db.WithContext(ctx).Preload("Asset").First(&tx, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *GormInvestTxRepo) GetByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]domain.InvestmentTransaction, error) {
	var list []domain.InvestmentTransaction
	err := r.db.WithContext(ctx).Where("portfolio_id = ?", portfolioID).Preload("Asset").Order("date DESC, created_at DESC").Find(&list).Error
	return list, err
}

func (r *GormInvestTxRepo) GetByPortfolioAndAsset(ctx context.Context, portfolioID, assetID uuid.UUID) ([]domain.InvestmentTransaction, error) {
	var list []domain.InvestmentTransaction
	err := r.db.WithContext(ctx).Where("portfolio_id = ? AND asset_id = ?", portfolioID, assetID).Preload("Asset").Order("date ASC, created_at ASC").Find(&list).Error
	return list, err
}

func (r *GormInvestTxRepo) Update(ctx context.Context, tx *domain.InvestmentTransaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *GormInvestTxRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.InvestmentTransaction{}, "id = ?", id).Error
}

// Earning Repository
type GormEarningRepo struct {
	db *gorm.DB
}

func NewEarningRepo(db *gorm.DB) *GormEarningRepo {
	return &GormEarningRepo{db: db}
}

func (r *GormEarningRepo) Create(ctx context.Context, earning *domain.Earning) error {
	return r.db.WithContext(ctx).Create(earning).Error
}

func (r *GormEarningRepo) GetByPortfolioID(ctx context.Context, portfolioID uuid.UUID) ([]domain.Earning, error) {
	var list []domain.Earning
	err := r.db.WithContext(ctx).Where("portfolio_id = ?", portfolioID).Preload("Asset").Order("payment_date DESC").Find(&list).Error
	return list, err
}

func (r *GormEarningRepo) GetByDateRange(ctx context.Context, portfolioID uuid.UUID, start, end time.Time) ([]domain.Earning, error) {
	var list []domain.Earning
	err := r.db.WithContext(ctx).Where("portfolio_id = ? AND payment_date >= ? AND payment_date <= ?", portfolioID, start, end).
		Preload("Asset").Order("payment_date ASC").Find(&list).Error
	return list, err
}
