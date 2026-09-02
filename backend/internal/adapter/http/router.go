package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/invest/backend/internal/adapter/http/handler"
	"github.com/invest/backend/internal/adapter/http/middleware"
	"github.com/invest/backend/pkg/token"
)

type RouterConfig struct {
	TokenMaker        token.Maker
	AuthHandler       *handler.AuthHandler
	FinanceHandler    *handler.FinanceHandler
	CreditCardHandler *handler.CreditCardHandler
	InvestmentHandler *handler.InvestmentHandler
}

func SetupRouter(cfg RouterConfig) *gin.Engine {
	r := gin.Default()

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "1.0.0",
		})
	})

	v1 := r.Group("/api/v1")

	// Public Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", cfg.AuthHandler.Register)
		auth.POST("/login", cfg.AuthHandler.Login)
		auth.POST("/refresh", cfg.AuthHandler.Refresh)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.TokenMaker))
	{
		// User info
		protected.GET("/auth/me", cfg.AuthHandler.Me)

		// Finance accounts & categories
		protected.POST("/accounts", cfg.FinanceHandler.CreateAccount)
		protected.GET("/accounts", cfg.FinanceHandler.GetAccounts)
		protected.GET("/categories", cfg.FinanceHandler.GetCategories)
		protected.POST("/categories", cfg.FinanceHandler.CreateCategory)

		// Transactions & Monthly Summary
		protected.POST("/transactions", cfg.FinanceHandler.CreateTransaction)
		protected.GET("/transactions", cfg.FinanceHandler.GetTransactions)
		protected.PUT("/transactions/:id", cfg.FinanceHandler.UpdateTransaction)
		protected.DELETE("/transactions/:id", cfg.FinanceHandler.DeleteTransaction)
		protected.GET("/transactions/summary", cfg.FinanceHandler.GetMonthlySummary)

		// Credit Cards & Invoices
		protected.POST("/cards", cfg.CreditCardHandler.CreateCard)
		protected.GET("/cards", cfg.CreditCardHandler.GetCards)
		protected.POST("/cards/expense", cfg.CreditCardHandler.CreateExpense)
		protected.GET("/cards/:id/invoice", cfg.CreditCardHandler.GetInvoice)

		// Budgets
		protected.POST("/budgets", cfg.FinanceHandler.SetBudget)
		protected.GET("/budgets", cfg.FinanceHandler.GetBudgets)

		// Investments & Assets
		protected.GET("/investments/portfolios", cfg.InvestmentHandler.GetPortfolios)
		protected.GET("/investments/portfolios/:id/summary", cfg.InvestmentHandler.GetPortfolioSummary)
		protected.GET("/investments/portfolios/:id/orders", cfg.InvestmentHandler.GetOrders)
		protected.POST("/investments/orders/buy", cfg.InvestmentHandler.ExecuteBuy)
		protected.POST("/investments/orders/sell", cfg.InvestmentHandler.ExecuteSell)
		protected.PUT("/investments/orders/:id", cfg.InvestmentHandler.UpdateOrder)
		protected.DELETE("/investments/orders/:id", cfg.InvestmentHandler.DeleteOrder)
		protected.GET("/investments/assets/search", cfg.InvestmentHandler.SearchAssets)
		protected.GET("/investments/quote", cfg.InvestmentHandler.GetLiveTickerQuote)
		protected.POST("/investments/assets", cfg.InvestmentHandler.CreateOrGetAsset)
		protected.POST("/investments/assets/:id/quote", cfg.InvestmentHandler.UpdateQuote)
		protected.GET("/investments/portfolios/:id/earnings", cfg.InvestmentHandler.GetEarnings)
		protected.POST("/investments/earnings", cfg.InvestmentHandler.CreateEarning)
	}

	return r
}
