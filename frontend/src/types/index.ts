export type AccountType = 'checking' | 'savings' | 'investment' | 'cash' | 'wallet'
export type TransactionType = 'income' | 'expense' | 'transfer'
export type AssetType = 'stock' | 'fii' | 'bdr' | 'etf' | 'fixed_income' | 'crypto' | 'currency'
export type EarningType = 'dividend' | 'jcp' | 'yield' | 'interest'

export interface User {
  id: string
  name: string
  email: string
  base_currency: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

export interface Account {
  id: string
  name: string
  type: AccountType
  balance: string
  currency: string
  color: string
  institution: string
}

export interface Category {
  id: string
  name: string
  type: 'expense' | 'income'
  icon: string
  color: string
  is_default: boolean
}

export interface Transaction {
  id: string
  user_id: string
  account_id?: string
  category_id?: string
  type: TransactionType
  status: string
  amount: string
  date: string
  description: string
  notes?: string
  tags?: string
  account?: Account
  category?: Category
}

export interface BudgetProgress {
  budget: {
    id: string
    category_id?: string
    month_year: string
    amount_limit: string
    category?: Category
  }
  spent: string
  percentage: number
  is_alert: boolean
  is_exceeded: boolean
}

export interface MonthlySummary {
  total_income: string
  total_expense: string
  net_balance: string
  by_category: Record<string, string>
}

export interface Asset {
  id: string
  ticker: string
  name: string
  type: AssetType
  currency: string
  current_price: string
  sector?: string
}

export interface Position {
  id: string
  portfolio_id: string
  asset_id: string
  quantity: string
  average_price: string
  total_cost: string
  asset?: Asset
}

export interface PositionSummary {
  position: Position
  current_price: string
  current_value_brl: string
  total_cost_brl: string
  unrealized_pnl_brl: string
  profit_percentage: number
  allocation_percentage: number
}

export interface PortfolioSummary {
  portfolio_id: string
  portfolio_name: string
  total_equity_brl: string
  total_cost_brl: string
  total_pnl_brl: string
  total_profit_percentage: number
  positions: PositionSummary[]
  allocation_by_type: Record<string, string>
}

export interface Earning {
  id: string
  portfolio_id: string
  asset_id: string
  type: EarningType
  date_com: string
  payment_date: string
  value_per_share: string
  total_amount: string
  net_total_amount: string
  asset?: Asset
}

export type InvestmentTxType = 'buy' | 'sell' | 'split' | 'group' | 'amortization'

export interface InvestmentTransaction {
  id: string
  portfolio_id: string
  asset_id: string
  account_id?: string
  type: InvestmentTxType
  quantity: string
  unit_price: string
  fees: string
  total_amount: string
  realized_pnl: string
  date: string
  notes?: string
  created_at: string
  asset?: Asset
  account?: Account
}
