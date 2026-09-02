package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormFinanceRepo struct {
	db *gorm.DB
}

func NewFinanceRepo(db *gorm.DB) *GormFinanceRepo {
	return &GormFinanceRepo{db: db}
}

// AccountRepository implementation
func (r *GormFinanceRepo) Create(ctx context.Context, account *domain.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *GormFinanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	var acc domain.Account
	if err := r.db.WithContext(ctx).First(&acc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *GormFinanceRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Account, error) {
	var list []domain.Account
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *GormFinanceRepo) Update(ctx context.Context, account *domain.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *GormFinanceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Account{}, "id = ?", id).Error
}

// TransactionRepository implementation
type GormTxRepo struct {
	db *gorm.DB
}

func NewTxRepo(db *gorm.DB) *GormTxRepo {
	return &GormTxRepo{db: db}
}

func (r *GormTxRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *GormTxRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var tx domain.Transaction
	err := r.db.WithContext(ctx).Preload("Account").Preload("Category").Preload("CreditCard").First(&tx, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *GormTxRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	var list []domain.Transaction
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Preload("Account").Preload("Category").Preload("CreditCard").
		Order("date DESC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&list).Error
	return list, err
}

func (r *GormTxRepo) GetByMonth(ctx context.Context, userID uuid.UUID, year int, month int) ([]domain.Transaction, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	var list []domain.Transaction
	err := r.db.WithContext(ctx).Where("user_id = ? AND date >= ? AND date < ?", userID, startDate, endDate).
		Preload("Account").Preload("Category").Preload("CreditCard").
		Order("date DESC").Find(&list).Error
	return list, err
}

func (r *GormTxRepo) Update(ctx context.Context, tx *domain.Transaction) error {
	return r.db.WithContext(ctx).Save(tx).Error
}

func (r *GormTxRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Transaction{}, "id = ?", id).Error
}

// CategoryRepository implementation
type GormCategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *GormCategoryRepo {
	return &GormCategoryRepo{db: db}
}

func (r *GormCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *GormCategoryRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	var list []domain.Category
	err := r.db.WithContext(ctx).Where("user_id = ? OR is_default = ?", userID, true).Order("type ASC, name ASC").Find(&list).Error
	return list, err
}

func (r *GormCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	var cat domain.Category
	if err := r.db.WithContext(ctx).First(&cat, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

// BudgetRepository implementation
type GormBudgetRepo struct {
	db *gorm.DB
}

func NewBudgetRepo(db *gorm.DB) *GormBudgetRepo {
	return &GormBudgetRepo{db: db}
}

func (r *GormBudgetRepo) CreateOrUpdate(ctx context.Context, budget *domain.Budget) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(budget).Error
}

func (r *GormBudgetRepo) GetByMonth(ctx context.Context, userID uuid.UUID, monthYear string) ([]domain.Budget, error) {
	var list []domain.Budget
	err := r.db.WithContext(ctx).Where("user_id = ? AND month_year = ?", userID, monthYear).Preload("Category").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching budgets: %w", err)
	}
	return list, nil
}
