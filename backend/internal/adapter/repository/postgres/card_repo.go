package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"gorm.io/gorm"
)

type GormCreditCardRepo struct {
	db *gorm.DB
}

func NewCreditCardRepo(db *gorm.DB) *GormCreditCardRepo {
	return &GormCreditCardRepo{db: db}
}

func (r *GormCreditCardRepo) Create(ctx context.Context, card *domain.CreditCard) error {
	return r.db.WithContext(ctx).Create(card).Error
}

func (r *GormCreditCardRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CreditCard, error) {
	var card domain.CreditCard
	if err := r.db.WithContext(ctx).First(&card, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *GormCreditCardRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CreditCard, error) {
	var list []domain.CreditCard
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *GormCreditCardRepo) CreateTransactions(ctx context.Context, txs []domain.Transaction) error {
	return r.db.WithContext(ctx).Create(&txs).Error
}

func (r *GormCreditCardRepo) GetInvoiceTransactions(ctx context.Context, cardID uuid.UUID, invoiceMonth string) ([]domain.Transaction, error) {
	var list []domain.Transaction
	err := r.db.WithContext(ctx).
		Where("credit_card_id = ? AND invoice_month = ?", cardID, invoiceMonth).
		Preload("Category").
		Order("date ASC, installment_number ASC").
		Find(&list).Error
	return list, err
}

func (r *GormCreditCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.CreditCard{}, "id = ?", id).Error
}
