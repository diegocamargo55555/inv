package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type CreditCardUseCase struct {
	cardRepo CreditCardRepository
	catRepo  CategoryRepository
}

func NewCreditCardUseCase(cardRepo CreditCardRepository, catRepo CategoryRepository) *CreditCardUseCase {
	return &CreditCardUseCase{
		cardRepo: cardRepo,
		catRepo:  catRepo,
	}
}

// CreateCard creates a new credit card
func (uc *CreditCardUseCase) CreateCard(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	limit decimal.Decimal,
	closingDay, dueDay int,
	brand, color string,
) (*domain.CreditCard, error) {
	if color == "" {
		color = "#8B5CF6"
	}
	if brand == "" {
		brand = "Mastercard"
	}

	card := &domain.CreditCard{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       name,
		Limit:      limit,
		ClosingDay: closingDay,
		DueDay:     dueDay,
		Brand:      brand,
		Color:      color,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := uc.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}

// GetUserCards retrieves all credit cards for a user
func (uc *CreditCardUseCase) GetUserCards(ctx context.Context, userID uuid.UUID) ([]domain.CreditCard, error) {
	return uc.cardRepo.GetByUserID(ctx, userID)
}

// CreateCardExpense registers a purchase on a credit card, automatically distributing installments across invoices
func (uc *CreditCardUseCase) CreateCardExpense(
	ctx context.Context,
	userID uuid.UUID,
	cardID uuid.UUID,
	categoryID *uuid.UUID,
	totalAmount decimal.Decimal,
	installmentsCount int,
	purchaseDate time.Time,
	description, notes, tags string,
) ([]domain.Transaction, error) {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil || card == nil {
		return nil, domain.ErrCreditCardNotFound
	}
	if card.UserID != userID {
		return nil, domain.ErrCreditCardNotFound
	}

	parcels, err := card.GenerateInstallments(totalAmount, installmentsCount, purchaseDate, description)
	if err != nil {
		return nil, err
	}

	parentID := uuid.New()
	txs := make([]domain.Transaction, len(parcels))

	for i, p := range parcels {
		var pid *uuid.UUID
		if installmentsCount > 1 {
			pid = &parentID
		}

		txs[i] = domain.Transaction{
			ID:                  uuid.New(),
			UserID:              userID,
			CreditCardID:        &cardID,
			CategoryID:          categoryID,
			Type:                domain.TxTypeExpense,
			Status:              domain.TxStatusCompleted,
			Amount:              p.Amount,
			Date:                p.DueDate,
			Description:         p.Description,
			Notes:               notes,
			Tags:                tags,
			InstallmentNumber:   p.InstallmentNumber,
			TotalInstallments:   p.TotalInstallments,
			ParentTransactionID: pid,
			InvoiceMonth:        p.InvoiceMonth,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
	}

	if err := uc.cardRepo.CreateTransactions(ctx, txs); err != nil {
		return nil, err
	}

	return txs, nil
}

// InvoiceSummary represents a credit card invoice for a specific month
type InvoiceSummary struct {
	CardID       uuid.UUID            `json:"card_id"`
	CardName     string               `json:"card_name"`
	InvoiceMonth string               `json:"invoice_month"`
	TotalAmount  decimal.Decimal      `json:"total_amount"`
	ClosingDate  time.Time            `json:"closing_date"`
	DueDate      time.Time            `json:"due_date"`
	Transactions []domain.Transaction `json:"transactions"`
}

// GetInvoice gets invoice details and items for a specific month
func (uc *CreditCardUseCase) GetInvoice(ctx context.Context, userID, cardID uuid.UUID, monthYear string) (*InvoiceSummary, error) {
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil || card == nil {
		return nil, domain.ErrCreditCardNotFound
	}
	if card.UserID != userID {
		return nil, domain.ErrCreditCardNotFound
	}

	txs, err := uc.cardRepo.GetInvoiceTransactions(ctx, cardID, monthYear)
	if err != nil {
		return nil, err
	}

	total := decimal.Zero
	for _, t := range txs {
		total = total.Add(t.Amount)
	}

	return &InvoiceSummary{
		CardID:       card.ID,
		CardName:     card.Name,
		InvoiceMonth: monthYear,
		TotalAmount:  total,
		Transactions: txs,
	}, nil
}
