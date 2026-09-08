package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/invest/backend/internal/adapter/external"
	httpAdapter "github.com/invest/backend/internal/adapter/http"
	"github.com/invest/backend/internal/adapter/http/handler"
	"github.com/invest/backend/internal/adapter/repository/postgres"
	"github.com/invest/backend/internal/config"
	"github.com/invest/backend/internal/usecase"
	"github.com/invest/backend/pkg/token"
)

func main() {
	log.Println("Starting Investment & Personal Finance API...")

	// 1. Load Configurations
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize PostgreSQL DB with AutoMigrate and Seeds
	db, err := postgres.NewDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 3. Initialize Infrastructure Repositories
	userRepo := postgres.NewUserRepo(db)
	financeRepo := postgres.NewFinanceRepo(db)
	txRepo := postgres.NewTxRepo(db)
	catRepo := postgres.NewCategoryRepo(db)
	budgetRepo := postgres.NewBudgetRepo(db)
	portfolioRepo := postgres.NewPortfolioRepo(db)
	assetRepo := postgres.NewAssetRepo(db)
	investTxRepo := postgres.NewInvestTxRepo(db)
	earningRepo := postgres.NewEarningRepo(db)

	// 4. Initialize Security / Token Maker
	tokenMaker := token.NewJWTMaker(cfg.JWTSecret)

	// 5. Initialize External Market Client (Brapi for B3 + Finnhub for US Markets)
	marketClient := external.NewMarketClient(cfg.BrapiToken, cfg.FinnhubToken)

	// 6. Initialize Use Cases (Business Logic)
	authUC := usecase.NewAuthUseCase(userRepo, portfolioRepo, tokenMaker, cfg.JWTAccessExp, cfg.JWTRefreshExp)
	financeUC := usecase.NewFinanceUseCase(financeRepo, txRepo, budgetRepo, catRepo).WithMarketData(marketClient)
	investUC := usecase.NewInvestmentUseCase(portfolioRepo, assetRepo, investTxRepo, earningRepo, financeRepo, marketClient)
	marketUC := usecase.NewMarketUseCase(assetRepo, marketClient)

	// 7. Initialize Delivery Handlers
	authHandler := handler.NewAuthHandler(authUC)
	financeHandler := handler.NewFinanceHandler(financeUC)
	investHandler := handler.NewInvestmentHandler(investUC, marketUC)

	// 8. Setup Gin Engine & Routes
	router := httpAdapter.SetupRouter(httpAdapter.RouterConfig{
		TokenMaker:        tokenMaker,
		AuthHandler:       authHandler,
		FinanceHandler:    financeHandler,
		InvestmentHandler: investHandler,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// 10. Start Server in background with graceful shutdown
	go func() {
		log.Printf("Server listening on port %s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting.")
}
