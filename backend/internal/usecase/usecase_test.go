package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/invest/backend/internal/usecase"
	"github.com/invest/backend/pkg/token"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- In-Memory Mocks for Comprehensive UseCases Tests ---

type mockUserRepo struct {
	users         map[uuid.UUID]*domain.User
	usersByEmail  map[string]*domain.User
	refreshTokens map[string]*domain.RefreshToken
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:         make(map[uuid.UUID]*domain.User),
		usersByEmail:  make(map[string]*domain.User),
		refreshTokens: make(map[string]*domain.RefreshToken),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	m.users[u.ID] = u
	m.usersByEmail[u.Email] = u
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return m.users[id], nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.usersByEmail[email], nil
}
func (m *mockUserRepo) CreateRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	m.refreshTokens[t.Token] = t
	return nil
}
func (m *mockUserRepo) GetRefreshToken(ctx context.Context, t string) (*domain.RefreshToken, error) {
	return m.refreshTokens[t], nil
}
func (m *mockUserRepo) DeleteRefreshToken(ctx context.Context, t string) error {
	delete(m.refreshTokens, t)
	return nil
}

type mockPortfolioRepo struct {
	portfolios map[uuid.UUID]*domain.Portfolio
	positions  map[string]*domain.Position
	assetRepo  usecase.AssetRepository
}

func newMockPortfolioRepo(assetRepo usecase.AssetRepository) *mockPortfolioRepo {
	return &mockPortfolioRepo{
		portfolios: make(map[uuid.UUID]*domain.Portfolio),
		positions:  make(map[string]*domain.Position),
		assetRepo:  assetRepo,
	}
}

func (m *mockPortfolioRepo) Create(ctx context.Context, p *domain.Portfolio) error {
	m.portfolios[p.ID] = p
	return nil
}
func (m *mockPortfolioRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Portfolio, error) {
	return m.portfolios[id], nil
}
func (m *mockPortfolioRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Portfolio, error) {
	var list []domain.Portfolio
	for _, p := range m.portfolios {
		if p.UserID == userID {
			list = append(list, *p)
		}
	}
	return list, nil
}
func (m *mockPortfolioRepo) GetPosition(ctx context.Context, portfolioID, assetID uuid.UUID) (*domain.Position, error) {
	key := portfolioID.String() + "_" + assetID.String()
	pos := m.positions[key]
	if pos != nil && pos.Asset == nil && m.assetRepo != nil {
		pos.Asset, _ = m.assetRepo.GetByID(ctx, assetID)
	}
	return pos, nil
}
func (m *mockPortfolioRepo) SavePosition(ctx context.Context, pos *domain.Position) error {
	key := pos.PortfolioID.String() + "_" + pos.AssetID.String()
	if pos.Asset == nil && m.assetRepo != nil {
		pos.Asset, _ = m.assetRepo.GetByID(ctx, pos.AssetID)
	}
	m.positions[key] = pos
	return nil
}
func (m *mockPortfolioRepo) GetPositions(ctx context.Context, portfolioID uuid.UUID) ([]domain.Position, error) {
	var list []domain.Position
	for _, pos := range m.positions {
		if pos.PortfolioID == portfolioID {
			if pos.Asset == nil && m.assetRepo != nil {
				pos.Asset, _ = m.assetRepo.GetByID(ctx, pos.AssetID)
			}
			list = append(list, *pos)
		}
	}
	return list, nil
}

type mockAccountRepo struct {
	accounts map[uuid.UUID]*domain.Account
}

func newMockAccountRepo() *mockAccountRepo {
	return &mockAccountRepo{accounts: make(map[uuid.UUID]*domain.Account)}
}
func (m *mockAccountRepo) Create(ctx context.Context, a *domain.Account) error {
	m.accounts[a.ID] = a
	return nil
}
func (m *mockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	return m.accounts[id], nil
}
func (m *mockAccountRepo) GetByUserID(ctx context.Context, uid uuid.UUID) ([]domain.Account, error) {
	var list []domain.Account
	for _, a := range m.accounts {
		if a.UserID == uid {
			list = append(list, *a)
		}
	}
	return list, nil
}
func (m *mockAccountRepo) Update(ctx context.Context, a *domain.Account) error {
	m.accounts[a.ID] = a
	return nil
}
func (m *mockAccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.accounts, id)
	return nil
}

type mockTxRepo struct {
	txs map[uuid.UUID]*domain.Transaction
}

func newMockTxRepo() *mockTxRepo {
	return &mockTxRepo{txs: make(map[uuid.UUID]*domain.Transaction)}
}
func (m *mockTxRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	m.txs[tx.ID] = tx
	return nil
}
func (m *mockTxRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	return m.txs[id], nil
}
func (m *mockTxRepo) GetByUserID(ctx context.Context, uid uuid.UUID, limit, offset int) ([]domain.Transaction, error) {
	var list []domain.Transaction
	for _, t := range m.txs {
		if t.UserID == uid {
			list = append(list, *t)
		}
	}
	return list, nil
}
func (m *mockTxRepo) GetByMonth(ctx context.Context, uid uuid.UUID, year, month int) ([]domain.Transaction, error) {
	var list []domain.Transaction
	for _, t := range m.txs {
		if t.UserID == uid && t.Date.Year() == year && int(t.Date.Month()) == month {
			list = append(list, *t)
		}
	}
	return list, nil
}
func (m *mockTxRepo) Update(ctx context.Context, tx *domain.Transaction) error {
	m.txs[tx.ID] = tx
	return nil
}
func (m *mockTxRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.txs, id)
	return nil
}

type mockCardRepo struct {
	cards       map[uuid.UUID]*domain.CreditCard
	cardTxs     map[uuid.UUID][]domain.Transaction
	invoiceMap  map[string][]domain.Transaction
}

func newMockCardRepo() *mockCardRepo {
	return &mockCardRepo{
		cards:      make(map[uuid.UUID]*domain.CreditCard),
		cardTxs:    make(map[uuid.UUID][]domain.Transaction),
		invoiceMap: make(map[string][]domain.Transaction),
	}
}

func (m *mockCardRepo) Create(ctx context.Context, card *domain.CreditCard) error {
	m.cards[card.ID] = card
	return nil
}
func (m *mockCardRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.CreditCard, error) {
	return m.cards[id], nil
}
func (m *mockCardRepo) GetByUserID(ctx context.Context, uid uuid.UUID) ([]domain.CreditCard, error) {
	var list []domain.CreditCard
	for _, c := range m.cards {
		if c.UserID == uid {
			list = append(list, *c)
		}
	}
	return list, nil
}
func (m *mockCardRepo) CreateTransactions(ctx context.Context, txs []domain.Transaction) error {
	for _, tx := range txs {
		if tx.CreditCardID != nil {
			m.cardTxs[*tx.CreditCardID] = append(m.cardTxs[*tx.CreditCardID], tx)
			key := tx.CreditCardID.String() + "_" + tx.InvoiceMonth
			m.invoiceMap[key] = append(m.invoiceMap[key], tx)
		}
	}
	return nil
}
func (m *mockCardRepo) GetInvoiceTransactions(ctx context.Context, cardID uuid.UUID, monthYear string) ([]domain.Transaction, error) {
	key := cardID.String() + "_" + monthYear
	return m.invoiceMap[key], nil
}
func (m *mockCardRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.cards, id)
	return nil
}

type mockBudgetRepo struct {
	budgets map[string]*domain.Budget
}

func newMockBudgetRepo() *mockBudgetRepo {
	return &mockBudgetRepo{budgets: make(map[string]*domain.Budget)}
}
func (m *mockBudgetRepo) CreateOrUpdate(ctx context.Context, b *domain.Budget) error {
	key := b.UserID.String() + "_" + b.MonthYear + "_" + b.ID.String()
	m.budgets[key] = b
	return nil
}
func (m *mockBudgetRepo) GetByMonth(ctx context.Context, uid uuid.UUID, monthYear string) ([]domain.Budget, error) {
	var list []domain.Budget
	for _, b := range m.budgets {
		if b.UserID == uid && b.MonthYear == monthYear {
			list = append(list, *b)
		}
	}
	return list, nil
}

type mockCategoryRepo struct {
	categories map[uuid.UUID]*domain.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[uuid.UUID]*domain.Category)}
}
func (m *mockCategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	m.categories[c.ID] = c
	return nil
}
func (m *mockCategoryRepo) GetByUserID(ctx context.Context, uid uuid.UUID) ([]domain.Category, error) {
	var list []domain.Category
	for _, c := range m.categories {
		list = append(list, *c)
	}
	return list, nil
}
func (m *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return m.categories[id], nil
}

type mockAssetRepo struct {
	assets map[uuid.UUID]*domain.Asset
}

func newMockAssetRepo() *mockAssetRepo {
	return &mockAssetRepo{assets: make(map[uuid.UUID]*domain.Asset)}
}
func (m *mockAssetRepo) CreateOrUpdate(ctx context.Context, a *domain.Asset) error {
	m.assets[a.ID] = a
	return nil
}
func (m *mockAssetRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Asset, error) {
	return m.assets[id], nil
}
func (m *mockAssetRepo) GetByTicker(ctx context.Context, ticker string) (*domain.Asset, error) {
	for _, a := range m.assets {
		if a.Ticker == ticker {
			return a, nil
		}
	}
	return nil, nil
}
func (m *mockAssetRepo) Search(ctx context.Context, query string) ([]domain.Asset, error) {
	var list []domain.Asset
	for _, a := range m.assets {
		list = append(list, *a)
	}
	return list, nil
}
func (m *mockAssetRepo) UpdatePrice(ctx context.Context, id uuid.UUID, price decimal.Decimal) error {
	if a, ok := m.assets[id]; ok {
		a.CurrentPrice = price
	}
	return nil
}

type mockInvestTxRepo struct {
	txs []domain.InvestmentTransaction
}

func (m *mockInvestTxRepo) Create(ctx context.Context, tx *domain.InvestmentTransaction) error {
	m.txs = append(m.txs, *tx)
	return nil
}
func (m *mockInvestTxRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.InvestmentTransaction, error) {
	for i := range m.txs {
		if m.txs[i].ID == id {
			return &m.txs[i], nil
		}
	}
	return nil, nil
}
func (m *mockInvestTxRepo) GetByPortfolioID(ctx context.Context, pid uuid.UUID) ([]domain.InvestmentTransaction, error) {
	var list []domain.InvestmentTransaction
	for _, t := range m.txs {
		if t.PortfolioID == pid {
			list = append(list, t)
		}
	}
	return list, nil
}
func (m *mockInvestTxRepo) GetByPortfolioAndAsset(ctx context.Context, pid, aid uuid.UUID) ([]domain.InvestmentTransaction, error) {
	var list []domain.InvestmentTransaction
	for _, t := range m.txs {
		if t.PortfolioID == pid && t.AssetID == aid {
			list = append(list, t)
		}
	}
	return list, nil
}
func (m *mockInvestTxRepo) Update(ctx context.Context, tx *domain.InvestmentTransaction) error {
	for i := range m.txs {
		if m.txs[i].ID == tx.ID {
			m.txs[i] = *tx
			return nil
		}
	}
	return nil
}
func (m *mockInvestTxRepo) Delete(ctx context.Context, id uuid.UUID) error {
	var remaining []domain.InvestmentTransaction
	for _, t := range m.txs {
		if t.ID != id {
			remaining = append(remaining, t)
		}
	}
	m.txs = remaining
	return nil
}

type mockMarketProvider struct {
	quotes        map[string]decimal.Decimal
	exchangeRates map[string]decimal.Decimal
}

func newMockMarketProvider() *mockMarketProvider {
	return &mockMarketProvider{
		quotes:        make(map[string]decimal.Decimal),
		exchangeRates: make(map[string]decimal.Decimal),
	}
}

func (m *mockMarketProvider) GetQuote(ctx context.Context, ticker string) (decimal.Decimal, error) {
	if q, ok := m.quotes[ticker]; ok {
		return q, nil
	}
	return decimal.NewFromFloat(50.00), nil
}

func (m *mockMarketProvider) GetExchangeRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	key := from + "_" + to
	if r, ok := m.exchangeRates[key]; ok {
		return r, nil
	}
	return decimal.NewFromFloat(5.50), nil
}

// --- Tests ---

func TestAuthUseCase_Full(t *testing.T) {
	userRepo := newMockUserRepo()
	portfolioRepo := newMockPortfolioRepo(nil)
	tokenMaker := token.NewJWTMaker("supersecretkeyforusecasetesting32char")

	uc := usecase.NewAuthUseCase(userRepo, portfolioRepo, tokenMaker, 15*time.Minute, 7*24*time.Hour)
	ctx := context.Background()

	var refreshToken string

	t.Run("register new user successfully", func(t *testing.T) {
		res, err := uc.Register(ctx, "Lucas Silva", "lucas@example.com", "SenhaForte123!")
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.Equal(t, "lucas@example.com", res.User.Email)
		refreshToken = res.RefreshToken

		// Verify default portfolio created
		portfolios, err := portfolioRepo.GetByUserID(ctx, res.User.ID)
		require.NoError(t, err)
		assert.Len(t, portfolios, 1)
		assert.True(t, portfolios[0].IsDefault)
	})

	t.Run("refresh token rotation succeeds", func(t *testing.T) {
		res, err := uc.Refresh(ctx, refreshToken)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
		assert.NotEmpty(t, res.RefreshToken)
		assert.NotEqual(t, refreshToken, res.RefreshToken) // Token rotated!
	})

	t.Run("refreshing revoked token fails", func(t *testing.T) {
		_, err := uc.Refresh(ctx, refreshToken) // Old token was revoked
		assert.Error(t, err)
	})

	t.Run("duplicate registration fails", func(t *testing.T) {
		_, err := uc.Register(ctx, "Lucas Silva", "lucas@example.com", "OutraSenha123!")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserAlreadyExists, err)
	})

	t.Run("login with correct password succeeds", func(t *testing.T) {
		res, err := uc.Login(ctx, "lucas@example.com", "SenhaForte123!")
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.AccessToken)
	})

	t.Run("login with wrong password fails", func(t *testing.T) {
		_, err := uc.Login(ctx, "lucas@example.com", "SenhaErrada")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})
}

func TestFinanceUseCase_Full(t *testing.T) {
	accRepo := newMockAccountRepo()
	txRepo := newMockTxRepo()
	catRepo := newMockCategoryRepo()
	budgetRepo := newMockBudgetRepo()

	uc := usecase.NewFinanceUseCase(accRepo, txRepo, budgetRepo)
	ctx := context.Background()
	userID := uuid.New()

	categoryFood := &domain.Category{ID: uuid.New(), Name: "Alimentação", Type: domain.CategoryTypeExpense}
	_ = catRepo.Create(ctx, categoryFood)

	t.Run("create account and apply income and expense", func(t *testing.T) {
		acc, err := uc.CreateAccount(ctx, userID, "Itaú", domain.AccountTypeChecking, decimal.NewFromFloat(1000.00), "Itaú Unibanco", "#EC7000")
		require.NoError(t, err)
		require.NotNil(t, acc)
		assert.True(t, acc.Balance.Equal(decimal.NewFromFloat(1000.00)))

		accounts, err := uc.GetUserAccounts(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, accounts, 1)

		// Create Income transaction (R$ 3.000,00)
		date := time.Date(2026, time.March, 5, 10, 0, 0, 0, time.UTC)
		txIncome, err := uc.CreateTransaction(ctx, userID, &acc.ID, nil, domain.TxTypeIncome, decimal.NewFromFloat(3000.00), date, "Salário", "", "")
		require.NoError(t, err)
		require.NotNil(t, txIncome)

		// Create Expense transaction (R$ 500,00)
		txExpense, err := uc.CreateTransaction(ctx, userID, &acc.ID, &categoryFood.ID, domain.TxTypeExpense, decimal.NewFromFloat(500.00), date, "Supermercado", "", "")
		require.NoError(t, err)
		require.NotNil(t, txExpense)

		// Check monthly summary
		summary, err := uc.GetMonthlySummary(ctx, userID, 2026, 3)
		require.NoError(t, err)
		assert.True(t, summary.TotalIncome.Equal(decimal.NewFromFloat(3000.00)))
		assert.True(t, summary.TotalExpense.Equal(decimal.NewFromFloat(500.00)))
		assert.True(t, summary.NetBalance.Equal(decimal.NewFromFloat(2500.00)))
	})

	t.Run("budgets with progress and alerts", func(t *testing.T) {
		budget, err := uc.SetBudget(ctx, userID, &categoryFood.ID, "2026-03", decimal.NewFromFloat(600.00))
		require.NoError(t, err)
		require.NotNil(t, budget)

		progressList, err := uc.GetBudgetsWithProgress(ctx, userID, 2026, 3)
		require.NoError(t, err)
		require.Len(t, progressList, 1)

		// 500 spent of 600 limit = 83.33% -> isAlert true
		assert.True(t, progressList[0].IsAlert)
		assert.False(t, progressList[0].IsExceeded)
		assert.InDelta(t, 83.33, progressList[0].Percentage, 0.1)
	})
}

func TestCreditCardUseCase_Full(t *testing.T) {
	cardRepo := newMockCardRepo()
	uc := usecase.NewCreditCardUseCase(cardRepo)
	ctx := context.Background()
	userID := uuid.New()

	var card *domain.CreditCard

	t.Run("create credit card", func(t *testing.T) {
		var err error
		card, err = uc.CreateCard(ctx, userID, "Nubank Ultravioleta", decimal.NewFromFloat(8000.00), 20, 27, "Mastercard", "#820AD1")
		require.NoError(t, err)
		require.NotNil(t, card)

		cards, err := uc.GetUserCards(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, cards, 1)
	})

	t.Run("create installment expense distributed across invoices", func(t *testing.T) {
		purchaseDate := time.Date(2026, time.February, 10, 14, 0, 0, 0, time.UTC)
		txs, err := uc.CreateCardExpense(ctx, userID, card.ID, nil, decimal.NewFromFloat(300.00), 3, purchaseDate, "Passagem Aérea", "", "")
		require.NoError(t, err)
		require.Len(t, txs, 3)

		// Invoices: Feb (2026-02), Mar (2026-03), Apr (2026-04)
		assert.Equal(t, "2026-02", txs[0].InvoiceMonth)
		assert.Equal(t, "2026-03", txs[1].InvoiceMonth)
		assert.Equal(t, "2026-04", txs[2].InvoiceMonth)

		// Check February Invoice
		invFeb, err := uc.GetInvoice(ctx, userID, card.ID, "2026-02")
		require.NoError(t, err)
		assert.True(t, invFeb.TotalAmount.Equal(decimal.NewFromFloat(100.00)))
		assert.Len(t, invFeb.Transactions, 1)
	})
}

func TestInvestmentUseCase_FullWithFXAndSummary(t *testing.T) {
	assetRepo := newMockAssetRepo()
	portfolioRepo := newMockPortfolioRepo(assetRepo)
	investTxRepo := &mockInvestTxRepo{}
	accRepo := newMockAccountRepo()
	marketProvider := newMockMarketProvider()

	marketProvider.exchangeRates["USD_BRL"] = decimal.NewFromFloat(5.60)
	marketProvider.quotes["PETR4"] = decimal.NewFromFloat(42.00)
	marketProvider.quotes["AAPL"] = decimal.NewFromFloat(200.00)

	uc := usecase.NewInvestmentUseCase(portfolioRepo, assetRepo, investTxRepo, nil, accRepo, marketProvider)
	ctx := context.Background()

	userID := uuid.New()
	portfolio := &domain.Portfolio{ID: uuid.New(), UserID: userID, Name: "Carteira Global"}
	_ = portfolioRepo.Create(ctx, portfolio)

	assetB3 := &domain.Asset{
		ID:           uuid.New(),
		Ticker:       "PETR4",
		Name:         "Petrobras PN",
		Type:         domain.AssetTypeStock,
		Currency:     "BRL",
		CurrentPrice: decimal.NewFromFloat(42.00),
	}
	_ = assetRepo.CreateOrUpdate(ctx, assetB3)

	assetUS := &domain.Asset{
		ID:           uuid.New(),
		Ticker:       "AAPL",
		Name:         "Apple Inc.",
		Type:         domain.AssetTypeStock,
		Currency:     "USD",
		CurrentPrice: decimal.NewFromFloat(200.00),
	}
	_ = assetRepo.CreateOrUpdate(ctx, assetUS)

	t.Run("buy local and international assets", func(t *testing.T) {
		// Buy 100 PETR4 at R$ 35.00
		_, err := uc.ExecuteBuy(ctx, userID, portfolio.ID, assetB3.ID, nil, decimal.NewFromFloat(100), decimal.NewFromFloat(35.00), decimal.Zero, time.Now(), "Aporte BR")
		require.NoError(t, err)

		// Buy 10 AAPL at $150.00 USD
		_, err = uc.ExecuteBuy(ctx, userID, portfolio.ID, assetUS.ID, nil, decimal.NewFromFloat(10), decimal.NewFromFloat(150.00), decimal.Zero, time.Now(), "Aporte US")
		require.NoError(t, err)
	})

	t.Run("consolidated portfolio summary with FX conversion", func(t *testing.T) {
		summary, err := uc.GetPortfolioSummary(ctx, userID, portfolio.ID)
		require.NoError(t, err)
		require.NotNil(t, summary)

		// PETR4: 100 * R$ 42 = R$ 4200 (Cost: R$ 3500, Gain: R$ 700)
		// AAPL: 10 * $200 = $2000 USD * 5.60 FX = R$ 11200 (Cost: $1500 * 5.60 = R$ 8400, Gain: R$ 2800)
		// Total Equity: 4200 + 11200 = R$ 15400
		// Total Cost: 3500 + 8400 = R$ 11900
		// Total Gain: 700 + 2800 = R$ 3500

		assert.True(t, summary.TotalEquityBRL.Equal(decimal.NewFromFloat(15400.00)))
		assert.True(t, summary.TotalCostBRL.Equal(decimal.NewFromFloat(11900.00)))
		assert.True(t, summary.TotalPnLBRL.Equal(decimal.NewFromFloat(3500.00)))
		assert.Len(t, summary.Positions, 2)
	})
}

func TestInvestmentUseCase_EditAndRecalculate(t *testing.T) {
	assetRepo := newMockAssetRepo()
	portfolioRepo := newMockPortfolioRepo(assetRepo)
	investTxRepo := &mockInvestTxRepo{}
	accRepo := newMockAccountRepo()
	marketProvider := newMockMarketProvider()

	uc := usecase.NewInvestmentUseCase(portfolioRepo, assetRepo, investTxRepo, nil, accRepo, marketProvider)
	ctx := context.Background()
	userID := uuid.New()

	acc := &domain.Account{
		ID:      uuid.New(),
		UserID:  userID,
		Name:    "XP Investimentos",
		Balance: decimal.NewFromFloat(10000.00),
	}
	_ = accRepo.Create(ctx, acc)

	portfolio := &domain.Portfolio{ID: uuid.New(), UserID: userID, Name: "Ações Brasil"}
	_ = portfolioRepo.Create(ctx, portfolio)

	asset := &domain.Asset{
		ID:           uuid.New(),
		Ticker:       "WEGE3",
		Name:         "WEG S.A.",
		Type:         domain.AssetTypeStock,
		Currency:     "BRL",
		CurrentPrice: decimal.NewFromFloat(50.00),
	}
	_ = assetRepo.CreateOrUpdate(ctx, asset)

	// Step 1: Buy 100 @ 30.00 with account debit
	t1 := time.Date(2026, time.January, 10, 10, 0, 0, 0, time.UTC)
	buyTx1, err := uc.ExecuteBuy(ctx, userID, portfolio.ID, asset.ID, &acc.ID, decimal.NewFromFloat(100), decimal.NewFromFloat(30.00), decimal.Zero, t1, "1º Aporte")
	require.NoError(t, err)
	require.NotNil(t, buyTx1)

	// Verify account balance: 10000 - 3000 = 7000
	updatedAcc, _ := accRepo.GetByID(ctx, acc.ID)
	assert.True(t, updatedAcc.Balance.Equal(decimal.NewFromFloat(7000.00)))

	// Step 2: Buy 100 @ 40.00
	t2 := time.Date(2026, time.January, 20, 10, 0, 0, 0, time.UTC)
	buyTx2, err := uc.ExecuteBuy(ctx, userID, portfolio.ID, asset.ID, nil, decimal.NewFromFloat(100), decimal.NewFromFloat(40.00), decimal.Zero, t2, "2º Aporte")
	require.NoError(t, err)
	require.NotNil(t, buyTx2)

	// Step 3: Sell 50 @ 50.00
	t3 := time.Date(2026, time.January, 30, 10, 0, 0, 0, time.UTC)
	sellTx, err := uc.ExecuteSell(ctx, userID, portfolio.ID, asset.ID, nil, decimal.NewFromFloat(50), decimal.NewFromFloat(50.00), decimal.Zero, t3, "Venda Parcial")
	require.NoError(t, err)
	require.NotNil(t, sellTx)

	// Check position before edit: Qty 150, PM 35.00, TotalCost 5250.00
	posBefore, err := portfolioRepo.GetPosition(ctx, portfolio.ID, asset.ID)
	require.NoError(t, err)
	assert.True(t, posBefore.Quantity.Equal(decimal.NewFromFloat(150)))
	assert.True(t, posBefore.AveragePrice.Equal(decimal.NewFromFloat(35.00)))

	t.Run("edit buy order updates quantity and automatically recalculates average price and positions", func(t *testing.T) {
		// Edit buyTx1: Change quantity from 100 to 200 shares @ 30.00
		updatedBuyTx, err := uc.UpdateOrder(ctx, userID, buyTx1.ID, decimal.NewFromFloat(200), decimal.NewFromFloat(30.00), decimal.Zero, t1, "1º Aporte Editado", &acc.ID)
		require.NoError(t, err)
		require.NotNil(t, updatedBuyTx)
		assert.True(t, updatedBuyTx.Quantity.Equal(decimal.NewFromFloat(200)))
		assert.True(t, updatedBuyTx.TotalAmount.Equal(decimal.NewFromFloat(6000.00)))

		// Account balance should have adjusted: 10000 - 6000 = 4000
		accAfterEdit, _ := accRepo.GetByID(ctx, acc.ID)
		assert.True(t, accAfterEdit.Balance.Equal(decimal.NewFromFloat(4000.00)))

		// Check recalculated position:
		// Buy 200 @ 30 (6000) + Buy 100 @ 40 (4000) = 300 shares @ cost 10000 -> PM 33.3333
		// Sell 50 shares leaves 250 shares @ PM 33.3333
		posAfter, err := portfolioRepo.GetPosition(ctx, portfolio.ID, asset.ID)
		require.NoError(t, err)
		assert.True(t, posAfter.Quantity.Equal(decimal.NewFromFloat(250)))
		assert.True(t, posAfter.AveragePrice.Equal(decimal.NewFromFloat(33.3333)))
	})

	t.Run("edit sell order recalculates remaining position and realized PnL", func(t *testing.T) {
		// Edit sellTx: Change sell quantity from 50 to 100 shares @ 50.00
		updatedSellTx, err := uc.UpdateOrder(ctx, userID, sellTx.ID, decimal.NewFromFloat(100), decimal.NewFromFloat(50.00), decimal.Zero, t3, "Venda Editada", nil)
		require.NoError(t, err)
		require.NotNil(t, updatedSellTx)
		assert.True(t, updatedSellTx.Quantity.Equal(decimal.NewFromFloat(100)))

		// Check position after edit: 300 - 100 = 200 shares
		posAfterSellEdit, err := portfolioRepo.GetPosition(ctx, portfolio.ID, asset.ID)
		require.NoError(t, err)
		assert.True(t, posAfterSellEdit.Quantity.Equal(decimal.NewFromFloat(200)))
	})

	t.Run("delete order recalculates entire portfolio position", func(t *testing.T) {
		// Delete buyTx2 (100 @ 40)
		err := uc.DeleteOrder(ctx, userID, buyTx2.ID)
		require.NoError(t, err)

		// Remaining: Buy 200 @ 30 (6000) -> Sell 100 @ 50 -> Remaining 100 @ PM 30.00
		posAfterDelete, err := portfolioRepo.GetPosition(ctx, portfolio.ID, asset.ID)
		require.NoError(t, err)
		assert.True(t, posAfterDelete.Quantity.Equal(decimal.NewFromFloat(100)))
		assert.True(t, posAfterDelete.AveragePrice.Equal(decimal.NewFromFloat(30.00)))
		assert.True(t, posAfterDelete.TotalCost.Equal(decimal.NewFromFloat(3000.00)))
	})
}

func TestFinanceUseCase_UpdateAndDeleteTransaction(t *testing.T) {
	accRepo := newMockAccountRepo()
	txRepo := newMockTxRepo()
	budgetRepo := newMockBudgetRepo()

	uc := usecase.NewFinanceUseCase(accRepo, txRepo, budgetRepo)
	ctx := context.Background()
	userID := uuid.New()

	acc, err := uc.CreateAccount(ctx, userID, "Nubank", domain.AccountTypeChecking, decimal.NewFromFloat(2000.00), "Nu Pagamentos", "#820AD1")
	require.NoError(t, err)

	date := time.Date(2026, time.March, 10, 12, 0, 0, 0, time.UTC)
	tx, err := uc.CreateTransaction(ctx, userID, &acc.ID, nil, domain.TxTypeExpense, decimal.NewFromFloat(500.00), date, "Mercado", "", "")
	require.NoError(t, err)

	// Balance after expense: 2000 - 500 = 1500
	accAfter, _ := accRepo.GetByID(ctx, acc.ID)
	assert.True(t, accAfter.Balance.Equal(decimal.NewFromFloat(1500.00)))

	t.Run("update transaction amount adjusts account balance", func(t *testing.T) {
		// Update expense from 500 to 700
		updatedTx, err := uc.UpdateTransaction(ctx, userID, tx.ID, &acc.ID, nil, domain.TxTypeExpense, decimal.NewFromFloat(700.00), date, "Mercado Grande", "", "")
		require.NoError(t, err)
		assert.True(t, updatedTx.Amount.Equal(decimal.NewFromFloat(700.00)))

		// Balance after update: 2000 - 700 = 1300
		accUpdated, _ := accRepo.GetByID(ctx, acc.ID)
		assert.True(t, accUpdated.Balance.Equal(decimal.NewFromFloat(1300.00)))
	})

	t.Run("delete transaction reverts account balance", func(t *testing.T) {
		err := uc.DeleteTransaction(ctx, userID, tx.ID)
		require.NoError(t, err)

		// Balance after delete: returns to initial 2000
		accReverted, _ := accRepo.GetByID(ctx, acc.ID)
		assert.True(t, accReverted.Balance.Equal(decimal.NewFromFloat(2000.00)))
	})

	t.Run("create transaction requires account ID", func(t *testing.T) {
		_, err := uc.CreateTransaction(ctx, userID, nil, nil, domain.TxTypeIncome, decimal.NewFromFloat(100.00), date, "Sem Conta", "", "")
		assert.ErrorIs(t, err, domain.ErrAccountRequired)
	})

	t.Run("update transaction requires account ID", func(t *testing.T) {
		_, err := uc.UpdateTransaction(ctx, userID, tx.ID, nil, nil, domain.TxTypeExpense, decimal.NewFromFloat(100.00), date, "Edit Sem Conta", "", "")
		assert.ErrorIs(t, err, domain.ErrAccountRequired)
	})
}
