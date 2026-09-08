package handler_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/invest/backend/internal/domain"
	"github.com/shopspring/decimal"
)

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
}

func newMockPortfolioRepo() *mockPortfolioRepo {
	return &mockPortfolioRepo{
		portfolios: make(map[uuid.UUID]*domain.Portfolio),
		positions:  make(map[string]*domain.Position),
	}
}

func (m *mockPortfolioRepo) Create(ctx context.Context, p *domain.Portfolio) error {
	m.portfolios[p.ID] = p
	return nil
}
func (m *mockPortfolioRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Portfolio, error) {
	return m.portfolios[id], nil
}
func (m *mockPortfolioRepo) GetByUserID(ctx context.Context, uid uuid.UUID) ([]domain.Portfolio, error) {
	var list []domain.Portfolio
	for _, p := range m.portfolios {
		if p.UserID == uid {
			list = append(list, *p)
		}
	}
	return list, nil
}
func (m *mockPortfolioRepo) GetPosition(ctx context.Context, pid, aid uuid.UUID) (*domain.Position, error) {
	key := pid.String() + "_" + aid.String()
	return m.positions[key], nil
}
func (m *mockPortfolioRepo) SavePosition(ctx context.Context, pos *domain.Position) error {
	key := pos.PortfolioID.String() + "_" + pos.AssetID.String()
	m.positions[key] = pos
	return nil
}
func (m *mockPortfolioRepo) GetPositions(ctx context.Context, pid uuid.UUID) ([]domain.Position, error) {
	var list []domain.Position
	for _, p := range m.positions {
		if p.PortfolioID == pid {
			list = append(list, *p)
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

type mockEarningRepo struct {
	earnings []domain.Earning
}

func (m *mockEarningRepo) Create(ctx context.Context, e *domain.Earning) error {
	m.earnings = append(m.earnings, *e)
	return nil
}
func (m *mockEarningRepo) GetByPortfolioID(ctx context.Context, pid uuid.UUID) ([]domain.Earning, error) {
	return m.earnings, nil
}
func (m *mockEarningRepo) GetByDateRange(ctx context.Context, pid uuid.UUID, start, end time.Time) ([]domain.Earning, error) {
	return m.earnings, nil
}

type mockMarketProvider struct{}

func newMockMarketProvider() *mockMarketProvider {
	return &mockMarketProvider{}
}
func (m *mockMarketProvider) GetQuote(ctx context.Context, ticker string) (decimal.Decimal, error) {
	return decimal.NewFromFloat(40.00), nil
}
func (m *mockMarketProvider) GetExchangeRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	return decimal.NewFromFloat(5.50), nil
}
func (m *mockMarketProvider) SearchAssets(ctx context.Context, query string) ([]domain.Asset, error) {
	return nil, nil
}
