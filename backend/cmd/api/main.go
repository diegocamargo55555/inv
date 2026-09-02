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
	"github.com/redis/go-redis/v9"
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

	// 3. Initialize Redis Client
	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			log.Printf("Warning: Redis connection ping failed: %v. Running in-memory/direct mode.", err)
		} else {
			log.Println("Redis connection established.")
		}
	}

	// 4. Initialize Infrastructure Repositories
	userRepo := postgres.NewUserRepo(db)
	financeRepo := postgres.NewFinanceRepo(db)
	txRepo := postgres.NewTxRepo(db)
	catRepo := postgres.NewCategoryRepo(db)
	budgetRepo := postgres.NewBudgetRepo(db)
	cardRepo := postgres.NewCreditCardRepo(db)
	portfolioRepo := postgres.NewPortfolioRepo(db)
	assetRepo := postgres.NewAssetRepo(db)
	investTxRepo := postgres.NewInvestTxRepo(db)
	earningRepo := postgres.NewEarningRepo(db)

	// 5. Initialize Security / Token Maker
	tokenMaker := token.NewJWTMaker(cfg.JWTSecret)

	// 6. Initialize External Market Client (Brapi for B3 + Finnhub for US Markets)
	marketClient := external.NewMarketClient(redisClient, cfg.BrapiToken, cfg.FinnhubToken)

	// 7. Initialize Use Cases (Business Logic)
	authUC := usecase.NewAuthUseCase(userRepo, portfolioRepo, tokenMaker, cfg.JWTAccessExp, cfg.JWTRefreshExp)
	financeUC := usecase.NewFinanceUseCase(financeRepo, txRepo, catRepo, budgetRepo)
	cardUC := usecase.NewCreditCardUseCase(cardRepo, catRepo)
	investUC := usecase.NewInvestmentUseCase(portfolioRepo, assetRepo, investTxRepo, earningRepo, financeRepo, marketClient)
	marketUC := usecase.NewMarketUseCase(assetRepo, marketClient)

	// 8. Initialize Delivery Handlers
	authHandler := handler.NewAuthHandler(authUC)
	financeHandler := handler.NewFinanceHandler(financeUC, catRepo)
	cardHandler := handler.NewCreditCardHandler(cardUC)
	investHandler := handler.NewInvestmentHandler(investUC, marketUC, portfolioRepo, assetRepo, earningRepo)

	// 9. Setup Gin Engine & Routes
	router := httpAdapter.SetupRouter(httpAdapter.RouterConfig{
		TokenMaker:        tokenMaker,
		AuthHandler:       authHandler,
		FinanceHandler:    financeHandler,
		CreditCardHandler: cardHandler,
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
