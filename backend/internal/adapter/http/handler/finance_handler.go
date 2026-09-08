package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/invest/backend/internal/adapter/http/dto"
	"github.com/invest/backend/internal/adapter/http/middleware"
	"github.com/invest/backend/internal/domain"
	"github.com/invest/backend/internal/usecase"
	"github.com/invest/backend/pkg/token"
)

type FinanceHandler struct {
	financeUC *usecase.FinanceUseCase
}

func NewFinanceHandler(financeUC *usecase.FinanceUseCase) *FinanceHandler {
	return &FinanceHandler{
		financeUC: financeUC,
	}
}

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(middleware.AuthorizationPayloadKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "não autenticado"})
		return uuid.Nil, false
	}
	p := val.(*token.Payload)
	return p.UserID, true
}

// Account Endpoints
func (h *FinanceHandler) CreateAccount(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	acc, err := h.financeUC.CreateAccount(c.Request.Context(), userID, req.Name, req.Type, req.InitialBalance, req.Currency, req.Institution, req.Color)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, acc)
}

func (h *FinanceHandler) GetAccounts(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	accounts, err := h.financeUC.GetUserAccounts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// Transaction Endpoints
func (h *FinanceHandler) CreateTransaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AccountID == nil || *req.AccountID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrAccountRequired.Error()})
		return
	}

	tx, err := h.financeUC.CreateTransaction(
		c.Request.Context(),
		userID,
		req.AccountID,
		req.CategoryID,
		req.Type,
		req.Amount,
		req.Date,
		req.Description,
		req.Notes,
		req.Tags,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

func (h *FinanceHandler) GetTransactions(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	txs, err := h.financeUC.GetTransactions(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, txs)
}

func (h *FinanceHandler) GetMonthlySummary(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	now := time.Now()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))

	summary, err := h.financeUC.GetMonthlySummary(c.Request.Context(), userID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// Category Endpoints
func (h *FinanceHandler) GetCategories(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	categories, err := h.financeUC.GetCategories(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *FinanceHandler) CreateCategory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req struct {
		Name  string              `json:"name" binding:"required"`
		Type  domain.CategoryType `json:"type" binding:"required"`
		Icon  string              `json:"icon"`
		Color string              `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cat := &domain.Category{
		ID:        uuid.New(),
		UserID:    &userID,
		Name:      req.Name,
		Type:      req.Type,
		Icon:      req.Icon,
		Color:     req.Color,
		IsDefault: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.financeUC.CreateCategory(c.Request.Context(), cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cat)
}

// Budget Endpoints
func (h *FinanceHandler) SetBudget(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.SetBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.financeUC.SetBudget(c.Request.Context(), userID, req.CategoryID, req.MonthYear, req.AmountLimit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, budget)
}

func (h *FinanceHandler) GetBudgets(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	now := time.Now()
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))

	budgets, err := h.financeUC.GetBudgetsWithProgress(c.Request.Context(), userID, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, budgets)
}

func (h *FinanceHandler) UpdateTransaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da transação inválido"})
		return
	}

	var req dto.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.AccountID == nil || *req.AccountID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrAccountRequired.Error()})
		return
	}

	tx, err := h.financeUC.UpdateTransaction(
		c.Request.Context(),
		userID,
		txID,
		req.AccountID,
		req.CategoryID,
		req.Type,
		req.Amount,
		req.Date,
		req.Description,
		req.Notes,
		req.Tags,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *FinanceHandler) DeleteTransaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da transação inválido"})
		return
	}

	if err := h.financeUC.DeleteTransaction(c.Request.Context(), userID, txID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transação excluída com sucesso"})
}
