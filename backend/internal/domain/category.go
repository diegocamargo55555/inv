package domain

import (
	"time"

	"github.com/google/uuid"
)

type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
)

type Category struct {
	ID        uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID   `json:"user_id,omitempty" gorm:"type:uuid;index"` // Nullable for system default categories
	Name      string       `json:"name" gorm:"type:varchar(100);not null"`
	Type      CategoryType `json:"type" gorm:"type:varchar(20);not null"`
	Icon      string       `json:"icon" gorm:"type:varchar(50);default:'tag'"`
	Color     string       `json:"color" gorm:"type:varchar(20);default:'#6B7280'"`
	IsDefault bool         `json:"is_default" gorm:"default:false"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}
