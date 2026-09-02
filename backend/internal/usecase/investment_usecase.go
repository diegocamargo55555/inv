package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type InvestmentUseCase struct {
	portfolioRepo PortfolioRepository
	assetRepo     AssetRepository
	txRepo        InvestmentTxRepository
	earningRepo   EarningRepository
	accRepo       AccountRepository
	marketData    MarketDataProvider
}

func NewInvestmentUseCase(
	portfolioRepo PortfolioRepository,
	assetRepo AssetRepository,
	txRepo InvestmentTxRepository,
	earningRepo EarningRepository,
	accRepo AccountRepository,
	marketData MarketDataProvider,
) *InvestmentUseCase {
	return &InvestmentUseCase{
		portfolioRepo: portfolioRepo,
		assetRepo:     assetRepo,
		txRepo:        txRepo,
		earningRepo:   earningRepo,
		accRepo:       accRepo,
		marketData:    marketData,
	}
}

// ExecuteBuy executes a buy transaction, updating position average price and cash account
func (uc *InvestmentUseCase) ExecuteBuy(
	ctx context.Context,
	userID uuid.UUID,
	portfolioID uuid.UUID,
	assetID uuid.UUID,
	accountID *uuid.UUID,
	quantity, unitPrice, fees decimal.Decimal,
	date time.Time,
	notes string,
) (*domain.InvestmentTransaction, error) {
	portfolio, err := uc.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil || portfolio == nil || portfolio.UserID != userID {
		return nil, domain.ErrPortfolioNotFound
	}

	asset, err := uc.assetRepo.GetByID(ctx, assetID)
	if err != nil || asset == nil {
		return nil, domain.ErrAssetNotFound
	}

	pos, _ := uc.portfolioRepo.GetPosition(ctx, portfolioID, assetID)
	if pos == nil {
		pos = &domain.Position{
			ID:           uuid.New(),
			PortfolioID:  portfolioID,
			AssetID:      assetID,
			Quantity:     decimal.Zero,
			AveragePrice: decimal.Zero,
			TotalCost:    decimal.Zero,
			UpdatedAt:    time.Now(),
		}
	}

	if err := pos.ApplyBuy(quantity, unitPrice, fees); err != nil {
		return nil, err
	}

	if err := uc.portfolioRepo.SavePosition(ctx, pos); err != nil {
		return nil, err
	}

	totalAmount := quantity.Mul(unitPrice).Add(fees)
	tx := &domain.InvestmentTransaction{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		AssetID:     assetID,
		AccountID:   accountID,
		Type:        domain.InvestTxBuy,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Fees:        fees,
		TotalAmount: totalAmount,
		Date:        date,
		Notes:       notes,
		CreatedAt:   time.Now(),
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// If account specified, deduct balance
	if accountID != nil {
		acc, err := uc.accRepo.GetByID(ctx, *accountID)
		if err == nil && acc != nil && acc.UserID == userID {
			_ = acc.ApplyTransaction(totalAmount, domain.TxTypeExpense)
			_ = uc.accRepo.Update(ctx, acc)
		}
	}

	return tx, nil
}

// ExecuteSell executes a sell transaction and computes realized gain/loss
func (uc *InvestmentUseCase) ExecuteSell(
	ctx context.Context,
	userID uuid.UUID,
	portfolioID uuid.UUID,
	assetID uuid.UUID,
	accountID *uuid.UUID,
	quantity, unitPrice, fees decimal.Decimal,
	date time.Time,
	notes string,
) (*domain.InvestmentTransaction, error) {
	portfolio, err := uc.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil || portfolio == nil || portfolio.UserID != userID {
		return nil, domain.ErrPortfolioNotFound
	}

	pos, err := uc.portfolioRepo.GetPosition(ctx, portfolioID, assetID)
	if err != nil || pos == nil {
		return nil, domain.ErrAssetNotFound
	}

	realizedPnL, err := pos.ApplySell(quantity, unitPrice, fees)
	if err != nil {
		return nil, err
	}

	if err := uc.portfolioRepo.SavePosition(ctx, pos); err != nil {
		return nil, err
	}

	grossTotal := quantity.Mul(unitPrice).Sub(fees)
	tx := &domain.InvestmentTransaction{
		ID:          uuid.New(),
		PortfolioID: portfolioID,
		AssetID:     assetID,
		AccountID:   accountID,
		Type:        domain.InvestTxSell,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Fees:        fees,
		TotalAmount: grossTotal,
		RealizedPnL: realizedPnL,
		Date:        date,
		Notes:       notes,
		CreatedAt:   time.Now(),
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	// If account specified, credit balance
	if accountID != nil {
		acc, err := uc.accRepo.GetByID(ctx, *accountID)
		if err == nil && acc != nil && acc.UserID == userID {
			_ = acc.ApplyTransaction(grossTotal, domain.TxTypeIncome)
			_ = uc.accRepo.Update(ctx, acc)
		}
	}

	return tx, nil
}

// PortfolioSummary consolidates all positions, calculates equity, allocation and profit/loss
type PositionSummary struct {
	Position         domain.Position `json:"position"`
	CurrentPrice     decimal.Decimal `json:"current_price"`
	CurrentValueBRL  decimal.Decimal `json:"current_value_brl"`
	TotalCostBRL     decimal.Decimal `json:"total_cost_brl"`
	UnrealizedPnLBRL decimal.Decimal `json:"unrealized_pnl_brl"`
	ProfitPercentage float64         `json:"profit_percentage"`
	AllocationPct    float64         `json:"allocation_percentage"`
}

type PortfolioConsolidatedSummary struct {
	PortfolioID      uuid.UUID                  `json:"portfolio_id"`
	PortfolioName    string                     `json:"portfolio_name"`
	TotalEquityBRL   decimal.Decimal            `json:"total_equity_brl"`
	TotalCostBRL     decimal.Decimal            `json:"total_cost_brl"`
	TotalPnLBRL      decimal.Decimal            `json:"total_pnl_brl"`
	TotalProfitPct   float64                    `json:"total_profit_percentage"`
	Positions        []PositionSummary          `json:"positions"`
	AllocationByType map[string]decimal.Decimal `json:"allocation_by_type"`
}

func (uc *InvestmentUseCase) GetPortfolioSummary(ctx context.Context, userID, portfolioID uuid.UUID) (*PortfolioConsolidatedSummary, error) {
	portfolio, err := uc.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil || portfolio == nil || portfolio.UserID != userID {
		return nil, domain.ErrPortfolioNotFound
	}

	positions, err := uc.portfolioRepo.GetPositions(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	usdBRLRate := decimal.NewFromFloat(5.50)
	if uc.marketData != nil {
		if rate, err := uc.marketData.GetExchangeRate(ctx, "USD", "BRL"); err == nil && !rate.IsZero() {
			usdBRLRate = rate
		}
	}

	summary := &PortfolioConsolidatedSummary{
		PortfolioID:      portfolio.ID,
		PortfolioName:    portfolio.Name,
		TotalEquityBRL:   decimal.Zero,
		TotalCostBRL:     decimal.Zero,
		TotalPnLBRL:      decimal.Zero,
		Positions:        make([]PositionSummary, 0),
		AllocationByType: make(map[string]decimal.Decimal),
	}

	for _, pos := range positions {
		if pos.Quantity.IsZero() {
			continue
		}

		currentPrice := pos.AveragePrice
		if pos.Asset != nil && !pos.Asset.CurrentPrice.IsZero() {
			currentPrice = pos.Asset.CurrentPrice
		}

		fxRate := decimal.NewFromInt(1)
		if pos.Asset != nil && pos.Asset.Currency == "USD" {
			fxRate = usdBRLRate
		}

		valBRL, pnlBRL := pos.CalculateValuationWithFX(currentPrice, fxRate)
		costBRL := pos.TotalCost.Mul(fxRate).Round(4)

		profitPct := 0.0
		if !costBRL.IsZero() {
			pDecimal := pnlBRL.Div(costBRL).Mul(decimal.NewFromInt(100))
			profitPct, _ = pDecimal.Float64()
		}

		summary.TotalEquityBRL = summary.TotalEquityBRL.Add(valBRL)
		summary.TotalCostBRL = summary.TotalCostBRL.Add(costBRL)
		summary.TotalPnLBRL = summary.TotalPnLBRL.Add(pnlBRL)

		assetType := "other"
		if pos.Asset != nil {
			assetType = string(pos.Asset.Type)
		}
		summary.AllocationByType[assetType] = summary.AllocationByType[assetType].Add(valBRL)

		summary.Positions = append(summary.Positions, PositionSummary{
			Position:         pos,
			CurrentPrice:     currentPrice,
			CurrentValueBRL:  valBRL,
			TotalCostBRL:     costBRL,
			UnrealizedPnLBRL: pnlBRL,
			ProfitPercentage: profitPct,
		})
	}

	if !summary.TotalCostBRL.IsZero() {
		totalPctDecimal := summary.TotalPnLBRL.Div(summary.TotalCostBRL).Mul(decimal.NewFromInt(100))
		summary.TotalProfitPct, _ = totalPctDecimal.Float64()
	}

	// Calculate individual allocation percentages
	if !summary.TotalEquityBRL.IsZero() {
		for i := range summary.Positions {
			allocDec := summary.Positions[i].CurrentValueBRL.Div(summary.TotalEquityBRL).Mul(decimal.NewFromInt(100))
			summary.Positions[i].AllocationPct, _ = allocDec.Float64()
		}
	}

	return summary, nil
}

// GetOrders retrieves all investment buy/sell transactions for a portfolio
func (uc *InvestmentUseCase) GetOrders(ctx context.Context, userID, portfolioID uuid.UUID) ([]domain.InvestmentTransaction, error) {
	portfolio, err := uc.portfolioRepo.GetByID(ctx, portfolioID)
	if err != nil || portfolio == nil || portfolio.UserID != userID {
		return nil, domain.ErrPortfolioNotFound
	}
	return uc.txRepo.GetByPortfolioID(ctx, portfolioID)
}

// UpdateOrder edits an existing buy or sell order and recalculates the asset position and PnL
func (uc *InvestmentUseCase) UpdateOrder(
	ctx context.Context,
	userID uuid.UUID,
	orderID uuid.UUID,
	quantity, unitPrice, fees decimal.Decimal,
	date time.Time,
	notes string,
	accountID *uuid.UUID,
) (*domain.InvestmentTransaction, error) {
	if quantity.LessThanOrEqual(decimal.Zero) || unitPrice.LessThanOrEqual(decimal.Zero) {
		return nil, domain.ErrInvalidAmount
	}
	if fees.LessThan(decimal.Zero) {
		fees = decimal.Zero
	}

	tx, err := uc.txRepo.GetByID(ctx, orderID)
	if err != nil || tx == nil {
		return nil, domain.ErrTransactionNotFound
	}

	portfolio, err := uc.portfolioRepo.GetByID(ctx, tx.PortfolioID)
	if err != nil || portfolio == nil || portfolio.UserID != userID {
		return nil, domain.ErrPortfolioNotFound
	}

	// Revert old account balance movement if applicable
	if tx.AccountID != nil {
		oldAcc, err := uc.accRepo.GetByID(ctx, *tx.AccountID)
		if err == nil && oldAcc != nil && oldAcc.UserID == userID {
			if tx.Type == domain.InvestTxBuy {
				_ = oldAcc.ApplyTransaction(tx.TotalAmount, domain.TxTypeIncome)
			} else if tx.Type == domain.InvestTxSell {
				_ = oldAcc.ApplyTransaction(tx.TotalAmount, domain.TxTypeExpense)
			}
			_ = uc.accRepo.Update(ctx, oldAcc)
		}
	}

	// Calculate new total amount
	var newTotal decimal.Decimal
	if tx.Type == domain.InvestTxBuy {
		newTotal = quantity.Mul(unitPrice).Add(fees)
	} else {
		newTotal = quantity.Mul(unitPrice).Sub(fees)
	}

	tx.Quantity = quantity
	tx.UnitPrice = unitPrice
	tx.Fees = fees
	tx.TotalAmount = newTotal
	tx.Date = date
	tx.Notes = notes
	tx.AccountID = accountID

	if err := uc.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}

	// Apply new account balance movement if applicable
	if accountID != nil {
		newAcc, err := uc.accRepo.GetByID(ctx, *accountID)
		if err == nil && newAcc != nil && newAcc.UserID == userID {
			if tx.Type == domain.InvestTxBuy {
				_ = newAcc.ApplyTransaction(newTotal, domain.TxTypeExpense)
			} else if tx.Type == domain.InvestTxSell {
				_ = newAcc.ApplyTransaction(newTotal, domain.TxTypeIncome)
			}
			_ = uc.accRepo.Update(ctx, newAcc)
		}
	}

	// Recalculate position for this asset
	if _, err := uc.RecalculatePosition(ctx, tx.PortfolioID, tx.AssetID); err != nil {
		return nil, err
	}

	// Return fresh tx
	return uc.txRepo.GetByID(ctx, orderID)
}

// DeleteOrder removes a buy/sell order and recalculates the asset position and PnL
func (uc *InvestmentUseCase) DeleteOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	tx, err := uc.txRepo.GetByID(ctx, orderID)
	if err != nil || tx == nil {
		return domain.ErrTransactionNotFound
	}

	portfolio, err := uc.portfolioRepo.GetByID(ctx, tx.PortfolioID)
	if err != nil || portfolio == nil || portfolio.UserID != userID {
		return domain.ErrPortfolioNotFound
	}

	// Revert account balance movement
	if tx.AccountID != nil {
		acc, err := uc.accRepo.GetByID(ctx, *tx.AccountID)
		if err == nil && acc != nil && acc.UserID == userID {
			if tx.Type == domain.InvestTxBuy {
				_ = acc.ApplyTransaction(tx.TotalAmount, domain.TxTypeIncome)
			} else if tx.Type == domain.InvestTxSell {
				_ = acc.ApplyTransaction(tx.TotalAmount, domain.TxTypeExpense)
			}
			_ = uc.accRepo.Update(ctx, acc)
		}
	}

	portfolioID := tx.PortfolioID
	assetID := tx.AssetID

	if err := uc.txRepo.Delete(ctx, orderID); err != nil {
		return err
	}

	_, err = uc.RecalculatePosition(ctx, portfolioID, assetID)
	return err
}

// RecalculatePosition replays all historical transactions for an asset in chronological order
func (uc *InvestmentUseCase) RecalculatePosition(ctx context.Context, portfolioID, assetID uuid.UUID) (*domain.Position, error) {
	txs, err := uc.txRepo.GetByPortfolioAndAsset(ctx, portfolioID, assetID)
	if err != nil {
		return nil, err
	}

	pos, _ := uc.portfolioRepo.GetPosition(ctx, portfolioID, assetID)
	if pos == nil {
		pos = &domain.Position{
			ID:          uuid.New(),
			PortfolioID: portfolioID,
			AssetID:     assetID,
			UpdatedAt:   time.Now(),
		}
	}

	pos.Quantity = decimal.Zero
	pos.AveragePrice = decimal.Zero
	pos.TotalCost = decimal.Zero
	pos.UpdatedAt = time.Now()

	for i := range txs {
		t := &txs[i]
		if t.Type == domain.InvestTxBuy {
			_ = pos.ApplyBuy(t.Quantity, t.UnitPrice, t.Fees)
		} else if t.Type == domain.InvestTxSell {
			realizedPnL, err := pos.ApplySell(t.Quantity, t.UnitPrice, t.Fees)
			if err == nil {
				t.RealizedPnL = realizedPnL
				_ = uc.txRepo.Update(ctx, t)
			}
		}
	}

	if err := uc.portfolioRepo.SavePosition(ctx, pos); err != nil {
		return nil, err
	}

	return pos, nil
}
