package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/invest/backend/internal/adapter/http/dto"
	"github.com/invest/backend/internal/usecase"
)

type CreditCardHandler struct {
	cardUC *usecase.CreditCardUseCase
}

func NewCreditCardHandler(cardUC *usecase.CreditCardUseCase) *CreditCardHandler {
	return &CreditCardHandler{cardUC: cardUC}
}

func (h *CreditCardHandler) CreateCard(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateCreditCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.cardUC.CreateCard(c.Request.Context(), userID, req.Name, req.Limit, req.ClosingDay, req.DueDay, req.Brand, req.Color)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, card)
}

func (h *CreditCardHandler) GetCards(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	cards, err := h.cardUC.GetUserCards(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cards)
}

func (h *CreditCardHandler) CreateExpense(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateCardExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txs, err := h.cardUC.CreateCardExpense(
		c.Request.Context(),
		userID,
		req.CardID,
		req.CategoryID,
		req.TotalAmount,
		req.InstallmentsCount,
		req.PurchaseDate,
		req.Description,
		req.Notes,
		req.Tags,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, txs)
}

func (h *CreditCardHandler) GetInvoice(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do cartão inválido"})
		return
	}

	now := time.Now()
	defaultMonthYear := fmt.Sprintf("%04d-%02d", now.Year(), now.Month())
	monthYear := c.DefaultQuery("month_year", defaultMonthYear)
	if len(c.Query("year")) > 0 && len(c.Query("month")) > 0 {
		y, _ := strconv.Atoi(c.Query("year"))
		m, _ := strconv.Atoi(c.Query("month"))
		monthYear = fmt.Sprintf("%04d-%02d", y, m)
	}

	invoice, err := h.cardUC.GetInvoice(c.Request.Context(), userID, cardID, monthYear)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}
