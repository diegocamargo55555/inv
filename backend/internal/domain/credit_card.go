package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreditCard struct {
	ID         uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID       `json:"user_id" gorm:"type:uuid;index;not null"`
	AccountID  *uuid.UUID      `json:"account_id,omitempty" gorm:"type:uuid"` // Conta vinculada para débito da fatura
	Name       string          `json:"name" gorm:"type:varchar(100);not null"`
	Limit      decimal.Decimal `json:"limit" gorm:"type:numeric(18,4);not null"`
	ClosingDay int             `json:"closing_day" gorm:"not null"` // Dia do fechamento da fatura (ex: 20)
	DueDay     int             `json:"due_day" gorm:"not null"`     // Dia do vencimento (ex: 27)
	Brand      string          `json:"brand" gorm:"type:varchar(50);default:'Mastercard'"`
	Color      string          `json:"color" gorm:"type:varchar(20);default:'#8B5CF6'"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Installment struct {
	InstallmentNumber int             `json:"installment_number"`
	TotalInstallments int             `json:"total_installments"`
	Amount            decimal.Decimal `json:"amount"`
	InvoiceMonth      string          `json:"invoice_month"` // "YYYY-MM"
	DueDate           time.Time       `json:"due_date"`
	ClosingDate       time.Time       `json:"closing_date"`
	Description       string          `json:"description"`
}

// CalculateInvoiceDates determines which invoice a transaction falls into based on closing day
func (c *CreditCard) CalculateInvoiceDates(purchaseDate time.Time) (closingDate, dueDate time.Time, monthYear string) {
	year := purchaseDate.Year()
	month := purchaseDate.Month()
	day := purchaseDate.Day()

	targetMonth := month
	targetYear := year

	// If purchase is after closing day, it belongs to the next month's invoice
	if day > c.ClosingDay {
		targetMonth = month + 1
		if targetMonth > 12 {
			targetMonth = time.January
			targetYear = year + 1
		}
	}

	closingDate = time.Date(targetYear, targetMonth, c.ClosingDay, 23, 59, 59, 0, time.UTC)

	dueMonth := targetMonth
	dueYear := targetYear
	if c.DueDay < c.ClosingDay {
		dueMonth = targetMonth + 1
		if dueMonth > 12 {
			dueMonth = time.January
			dueYear = targetYear + 1
		}
	}
	dueDate = time.Date(dueYear, dueMonth, c.DueDay, 23, 59, 59, 0, time.UTC)

	monthYear = fmt.Sprintf("%04d-%02d", targetYear, targetMonth)
	return closingDate, dueDate, monthYear
}

// GenerateInstallments splits an amount across future monthly invoices with exact cents remainder distribution
func (c *CreditCard) GenerateInstallments(totalAmount decimal.Decimal, installments int, purchaseDate time.Time, desc string) ([]Installment, error) {
	if installments <= 0 {
		return nil, ErrInvalidInstallmentCount
	}
	if totalAmount.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidAmount
	}

	// Calculate base parcel rounded down to 2 decimal places
	baseParcel := totalAmount.Div(decimal.NewFromInt(int64(installments))).RoundFloor(2)
	// Calculate remainder (centavos) to allocate to the first installment
	totalBase := baseParcel.Mul(decimal.NewFromInt(int64(installments)))
	remainder := totalAmount.Sub(totalBase)

	firstClosing, _, _ := c.CalculateInvoiceDates(purchaseDate)

	result := make([]Installment, installments)
	for i := 0; i < installments; i++ {
		// Advance months for each installment
		parcelDate := firstClosing.AddDate(0, i, 0)
		closingDate, dueDate, monthYear := c.CalculateInvoiceDates(parcelDate)

		amount := baseParcel
		if i == 0 {
			amount = amount.Add(remainder)
		}

		result[i] = Installment{
			InstallmentNumber: i + 1,
			TotalInstallments: installments,
			Amount:            amount,
			InvoiceMonth:      monthYear,
			DueDate:           dueDate,
			ClosingDate:       closingDate,
			Description:       fmt.Sprintf("%s (%d/%d)", desc, i+1, installments),
		}
	}

	return result, nil
}
