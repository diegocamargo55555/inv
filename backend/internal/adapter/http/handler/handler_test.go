package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	httpAdapter "github.com/invest/backend/internal/adapter/http"
	"github.com/invest/backend/internal/adapter/http/dto"
	"github.com/invest/backend/internal/adapter/http/handler"
	"github.com/invest/backend/internal/domain"
	"github.com/invest/backend/internal/usecase"
	"github.com/invest/backend/pkg/token"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *token.JWTMaker, *usecase.AuthUseCase) {
	gin.SetMode(gin.TestMode)
	tokenMaker := token.NewJWTMaker("supersecretkeyforhandlertesting32char")

	userRepo := newMockUserRepo()
	portfolioRepo := newMockPortfolioRepo()
	accountRepo := newMockAccountRepo()
	txRepo := newMockTxRepo()
	catRepo := newMockCategoryRepo()
	budgetRepo := newMockBudgetRepo()
	cardRepo := newMockCardRepo()
	assetRepo := newMockAssetRepo()
	investTxRepo := &mockInvestTxRepo{}
	earningRepo := &mockEarningRepo{}
	marketProvider := newMockMarketProvider()

	authUC := usecase.NewAuthUseCase(userRepo, portfolioRepo, tokenMaker, 15*time.Minute, 7*24*time.Hour)
	financeUC := usecase.NewFinanceUseCase(accountRepo, txRepo, budgetRepo)
	cardUC := usecase.NewCreditCardUseCase(cardRepo)
	investUC := usecase.NewInvestmentUseCase(portfolioRepo, assetRepo, investTxRepo, earningRepo, accountRepo, marketProvider)
	marketUC := usecase.NewMarketUseCase(assetRepo, marketProvider)

	authHandler := handler.NewAuthHandler(authUC)
	financeHandler := handler.NewFinanceHandler(financeUC, catRepo)
	cardHandler := handler.NewCreditCardHandler(cardUC)
	investHandler := handler.NewInvestmentHandler(investUC, marketUC, portfolioRepo, assetRepo, earningRepo)

	router := httpAdapter.SetupRouter(httpAdapter.RouterConfig{
		TokenMaker:        tokenMaker,
		AuthHandler:       authHandler,
		FinanceHandler:    financeHandler,
		CreditCardHandler: cardHandler,
		InvestmentHandler: investHandler,
	})

	return router, tokenMaker, authUC
}

func TestHTTPHandlers_AuthAndHealth(t *testing.T) {
	router, _, _ := setupTestRouter()

	t.Run("GET /api/v1/health returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
	})

	t.Run("POST /api/v1/auth/register creates user", func(t *testing.T) {
		body, _ := json.Marshal(dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "Password123!",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
	})

	t.Run("POST /api/v1/auth/login with valid credentials", func(t *testing.T) {
		body, _ := json.Marshal(dto.LoginRequest{
			Email:    "test@example.com",
			Password: "Password123!",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "access_token")
	})
}

func TestHTTPHandlers_FinanceAndCards(t *testing.T) {
	router, _, authUC := setupTestRouter()
	ctx := context.Background()

	res, err := authUC.Register(ctx, "Finance User", "finance@example.com", "Password123!")
	require.NoError(t, err)

	authHeader := "Bearer " + res.AccessToken

	var accountID string

	t.Run("POST /api/v1/accounts creates account", func(t *testing.T) {
		body, _ := json.Marshal(dto.CreateAccountRequest{
			Name:           "Nubank Conta",
			Type:           domain.AccountTypeChecking,
			InitialBalance: decimal.NewFromFloat(1500.00),
			Currency:       "BRL",
			Institution:    "Nu Pagamentos",
			Color:          "#820AD1",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var acc domain.Account
		_ = json.Unmarshal(w.Body.Bytes(), &acc)
		accountID = acc.ID.String()
		assert.Equal(t, "BRL", acc.Currency)
	})

	t.Run("POST /api/v1/accounts creates USD account", func(t *testing.T) {
		body, _ := json.Marshal(dto.CreateAccountRequest{
			Name:           "Nomad Global Account",
			Type:           domain.AccountTypeChecking,
			InitialBalance: decimal.NewFromFloat(500.00),
			Currency:       "USD",
			Institution:    "Nomad",
			Color:          "#00B4D8",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var usdAcc domain.Account
		_ = json.Unmarshal(w.Body.Bytes(), &usdAcc)
		assert.Equal(t, "USD", usdAcc.Currency)
		assert.Equal(t, "Nomad Global Account", usdAcc.Name)
		assert.True(t, usdAcc.Balance.Equal(decimal.NewFromFloat(500.00)))
	})

	var createdTxID string
	t.Run("POST /api/v1/transactions creates transaction and updates balance", func(t *testing.T) {
		accUUID := uuid.MustParse(accountID)
		body, _ := json.Marshal(dto.CreateTransactionRequest{
			AccountID:   &accUUID,
			Type:        domain.TxTypeIncome,
			Amount:      decimal.NewFromFloat(2000.00),
			Date:        time.Now(),
			Description: "Freelance",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var tx domain.Transaction
		_ = json.Unmarshal(w.Body.Bytes(), &tx)
		createdTxID = tx.ID.String()
	})

	t.Run("PUT /api/v1/transactions/:id updates transaction", func(t *testing.T) {
		accUUID := uuid.MustParse(accountID)
		body, _ := json.Marshal(dto.UpdateTransactionRequest{
			AccountID:   &accUUID,
			Type:        domain.TxTypeIncome,
			Amount:      decimal.NewFromFloat(2500.00),
			Date:        time.Now(),
			Description: "Freelance Editado",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+createdTxID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Freelance Editado")
	})

	t.Run("POST /api/v1/transactions returns 400 when AccountID is missing", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"type":        domain.TxTypeIncome,
			"amount":      100.0,
			"date":        time.Now(),
			"description": "Sem conta",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("PUT /api/v1/transactions/:id returns 400 when AccountID is missing", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{
			"type":        domain.TxTypeIncome,
			"amount":      100.0,
			"date":        time.Now(),
			"description": "Edit sem conta",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+createdTxID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/v1/transactions lists transactions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Freelance Editado")
	})

	t.Run("POST /api/v1/cards creates card", func(t *testing.T) {
		body, _ := json.Marshal(dto.CreateCreditCardRequest{
			Name:       "Mastercard Black",
			Limit:      decimal.NewFromFloat(10000.00),
			ClosingDay: 20,
			DueDay:     27,
			Brand:      "Mastercard",
			Color:      "#000000",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/cards", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestHTTPHandlers_InvestmentsOrdersCRUD(t *testing.T) {
	router, _, authUC := setupTestRouter()
	ctx := context.Background()

	res, err := authUC.Register(ctx, "Invest User", "investor_orders@example.com", "Password123!")
	require.NoError(t, err)

	authHeader := "Bearer " + res.AccessToken

	// Get user's default portfolio
	reqPort := httptest.NewRequest(http.MethodGet, "/api/v1/investments/portfolios", nil)
	reqPort.Header.Set("Authorization", authHeader)
	wPort := httptest.NewRecorder()
	router.ServeHTTP(wPort, reqPort)
	require.Equal(t, http.StatusOK, wPort.Code)

	var ports []domain.Portfolio
	_ = json.Unmarshal(wPort.Body.Bytes(), &ports)
	require.NotEmpty(t, ports)
	portID := ports[0].ID

	// Create an asset
	bodyAsset, _ := json.Marshal(map[string]string{
		"ticker":   "PETR4",
		"name":     "Petrobras",
		"type":     "stock",
		"currency": "BRL",
	})
	reqAsset := httptest.NewRequest(http.MethodPost, "/api/v1/investments/assets", bytes.NewBuffer(bodyAsset))
	reqAsset.Header.Set("Content-Type", "application/json")
	reqAsset.Header.Set("Authorization", authHeader)
	wAsset := httptest.NewRecorder()
	router.ServeHTTP(wAsset, reqAsset)
	require.Equal(t, http.StatusOK, wAsset.Code)

	var asset domain.Asset
	_ = json.Unmarshal(wAsset.Body.Bytes(), &asset)

	var orderID string

	t.Run("POST /api/v1/investments/orders/buy creates buy order", func(t *testing.T) {
		body, _ := json.Marshal(dto.ExecuteBuyOrderRequest{
			PortfolioID: portID,
			AssetID:     asset.ID,
			Quantity:    decimal.NewFromFloat(100),
			UnitPrice:   decimal.NewFromFloat(35.00),
			Fees:        decimal.NewFromFloat(5.00),
			Date:        time.Now(),
			Notes:       "Aporte inicial PETR4",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/investments/orders/buy", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var tx domain.InvestmentTransaction
		_ = json.Unmarshal(w.Body.Bytes(), &tx)
		orderID = tx.ID.String()
		assert.NotEmpty(t, orderID)
	})

	t.Run("GET /api/v1/investments/portfolios/:id/orders lists orders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/investments/portfolios/"+portID.String()+"/orders", nil)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Aporte inicial PETR4")
	})

	t.Run("PUT /api/v1/investments/orders/:id updates order", func(t *testing.T) {
		body, _ := json.Marshal(dto.UpdateInvestmentOrderRequest{
			Quantity:  decimal.NewFromFloat(150),
			UnitPrice: decimal.NewFromFloat(34.00),
			Fees:      decimal.NewFromFloat(4.50),
			Date:      time.Now(),
			Notes:     "Aporte inicial corrigido",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/investments/orders/"+orderID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Aporte inicial corrigido")
	})

	t.Run("DELETE /api/v1/investments/orders/:id deletes order and recalculates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/investments/orders/"+orderID, nil)
		req.Header.Set("Authorization", authHeader)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "recalculada com sucesso")
	})
}
