package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/invest/backend/internal/adapter/http/dto"
	"github.com/invest/backend/internal/domain"
	"github.com/invest/backend/internal/usecase"
)

type InvestmentHandler struct {
	investUC *usecase.InvestmentUseCase
	marketUC *usecase.MarketUseCase
}

func NewInvestmentHandler(
	investUC *usecase.InvestmentUseCase,
	marketUC *usecase.MarketUseCase,
) *InvestmentHandler {
	return &InvestmentHandler{
		investUC: investUC,
		marketUC: marketUC,
	}
}

// Portfolio Endpoints
func (h *InvestmentHandler) GetPortfolios(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	portfolios, err := h.investUC.GetUserPortfolios(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, portfolios)
}

func (h *InvestmentHandler) GetPortfolioSummary(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da carteira inválido"})
		return
	}

	summary, err := h.investUC.GetPortfolioSummary(c.Request.Context(), userID, portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// Order Execution Endpoints
func (h *InvestmentHandler) ExecuteBuy(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.ExecuteOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.investUC.ExecuteBuy(
		c.Request.Context(),
		userID,
		req.PortfolioID,
		req.AssetID,
		req.AccountID,
		req.Quantity,
		req.UnitPrice,
		req.Fees,
		req.Date,
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

func (h *InvestmentHandler) ExecuteSell(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.ExecuteOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.investUC.ExecuteSell(
		c.Request.Context(),
		userID,
		req.PortfolioID,
		req.AssetID,
		req.AccountID,
		req.Quantity,
		req.UnitPrice,
		req.Fees,
		req.Date,
		req.Notes,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tx)
}

// Asset Search & Quotes
func (h *InvestmentHandler) SearchAssets(c *gin.Context) {
	query := c.Query("q")
	assets, err := h.marketUC.SearchAssets(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assets)
}

func (h *InvestmentHandler) GetLiveTickerQuote(c *gin.Context) {
	ticker := c.Query("ticker")
	if ticker == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ticker é obrigatório"})
		return
	}

	quote, err := h.marketUC.GetLiveQuote(c.Request.Context(), ticker)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ticker":        ticker,
		"current_price": quote,
	})
}

func (h *InvestmentHandler) CreateOrGetAsset(c *gin.Context) {
	var req struct {
		Ticker   string           `json:"ticker" binding:"required"`
		Name     string           `json:"name"`
		Type     domain.AssetType `json:"type"`
		Currency string           `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset, err := h.marketUC.GetOrCreateAsset(c.Request.Context(), req.Ticker, req.Name, req.Type, req.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, asset)
}

func (h *InvestmentHandler) UpdateQuote(c *gin.Context) {
	assetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do ativo inválido"})
		return
	}

	quote, err := h.marketUC.UpdateAssetPrice(c.Request.Context(), assetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"current_price": quote})
}

// Earnings Endpoints
func (h *InvestmentHandler) GetEarnings(c *gin.Context) {
	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da carteira inválido"})
		return
	}

	earnings, err := h.investUC.GetEarnings(c.Request.Context(), portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, earnings)
}

func (h *InvestmentHandler) CreateEarning(c *gin.Context) {
	var req domain.Earning
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New()
	if req.NetTotalAmount.IsZero() {
		req.NetTotalAmount = req.TotalAmount
	}

	if err := h.investUC.CreateEarning(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *InvestmentHandler) GetOrders(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	portfolioID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da carteira inválido"})
		return
	}

	orders, err := h.investUC.GetOrders(c.Request.Context(), userID, portfolioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *InvestmentHandler) UpdateOrder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da ordem inválido"})
		return
	}

	var req dto.UpdateInvestmentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.investUC.UpdateOrder(
		c.Request.Context(),
		userID,
		orderID,
		req.Quantity,
		req.UnitPrice,
		req.Fees,
		req.Date,
		req.Notes,
		req.AccountID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *InvestmentHandler) DeleteOrder(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id da ordem inválido"})
		return
	}

	if err := h.investUC.DeleteOrder(c.Request.Context(), userID, orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ordem excluída e carteira recalculada com sucesso"})
}

func (h *InvestmentHandler) GetExchangeRate(c *gin.Context) {
	from := c.DefaultQuery("from", "USD")
	to := c.DefaultQuery("to", "BRL")

	rate, err := h.marketUC.GetExchangeRate(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from": from,
		"to":   to,
		"rate": rate.String(),
	})
}
