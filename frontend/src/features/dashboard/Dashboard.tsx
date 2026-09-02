import React, { useEffect, useState } from 'react'
import {
  Wallet,
  TrendingUp,
  ArrowDownRight,
  ArrowUpRight,
  PieChart as PieIcon,
  Activity,
  PlusCircle
} from 'lucide-react'
import { ResponsiveContainer, PieChart, Pie, Cell, Tooltip, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts'
import api from '../../services/api'
import { usePrivacyStore } from '../../stores/privacy'
import { formatCurrency, formatPercentage } from '../../utils/formatters'
import { Account, MonthlySummary, PortfolioSummary, Transaction } from '../../types'
import { Link } from 'react-router-dom'

const COLORS = ['#10B981', '#3B82F6', '#8B5CF6', '#F59E0B', '#EC4899', '#06B6D4']

export const Dashboard: React.FC = () => {
  const { hideValues } = usePrivacyStore()
  const [accounts, setAccounts] = useState<Account[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [summary, setSummary] = useState<MonthlySummary | null>(null)
  const [portfolio, setPortfolio] = useState<PortfolioSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [accRes, txRes, sumRes, portListRes] = await Promise.all([
          api.get<Account[]>('/accounts'),
          api.get<Transaction[]>('/transactions?limit=5'),
          api.get<MonthlySummary>('/transactions/summary'),
          api.get<any[]>('/investments/portfolios'),
        ])

        setAccounts(accRes.data)
        setTransactions(txRes.data)
        setSummary(sumRes.data)

        if (portListRes.data && portListRes.data.length > 0) {
          const defaultPort = portListRes.data[0]
          const portSummaryRes = await api.get<PortfolioSummary>(`/investments/portfolios/${defaultPort.id}/summary`)
          setPortfolio(portSummaryRes.data)
        }
      } catch (err) {
        console.error('Failed to load dashboard data:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  // Calculations
  const totalBankBalance = accounts.reduce((acc, a) => acc + parseFloat(a.balance || '0'), 0)
  const totalInvestments = portfolio ? parseFloat(portfolio.total_equity_brl || '0') : 0
  const netWorth = totalBankBalance + totalInvestments

  const displayVal = (amount: number | string, currency = 'BRL') =>
    hideValues ? '••••••' : formatCurrency(amount, currency)

  // Pie Chart Data
  const allocationData = portfolio?.allocation_by_type
    ? Object.entries(portfolio.allocation_by_type).map(([key, val]) => ({
        name: key === 'stock' ? 'Ações' : key === 'fii' ? 'FIIs' : key === 'crypto' ? 'Cripto' : key,
        value: parseFloat(val),
      }))
    : []

  // Cash Flow Chart Data
  const monthlyFlowData = [
    {
      name: 'Mês Atual',
      Receitas: parseFloat(summary?.total_income || '0'),
      Despesas: parseFloat(summary?.total_expense || '0'),
      Líquido: parseFloat(summary?.net_balance || '0'),
    },
  ]

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96 text-slate-400">
        <Activity className="w-6 h-6 animate-spin mr-2 text-brand-500" />
        <span>Carregando dados da sua conta...</span>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      {/* Header with Title & Quick Action */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Visão Geral Consolidada</h1>
          <p className="text-sm text-slate-400">Acompanhe seu patrimônio líquido, despesas e investimentos em tempo real.</p>
        </div>
        <div className="flex gap-3">
          <Link
            to="/finances"
            className="flex items-center gap-2 px-4 py-2 bg-brand-500 hover:bg-brand-600 active:scale-95 text-white text-xs font-semibold rounded-xl shadow-lg shadow-brand-500/20 transition-all"
          >
            <PlusCircle className="w-4 h-4" />
            <span>Nova Transação</span>
          </Link>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5">
        {/* Total Net Worth */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Patrimônio Líquido</span>
            <div className="w-8 h-8 rounded-xl bg-brand-500/10 text-brand-400 flex items-center justify-center">
              <TrendingUp className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-black text-white">{displayVal(netWorth)}</div>
            <span className="text-xs text-slate-500 mt-1 block">Contas bancárias + Investimentos</span>
          </div>
        </div>

        {/* Bank & Cash Balance */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Saldo em Contas</span>
            <div className="w-8 h-8 rounded-xl bg-blue-500/10 text-blue-400 flex items-center justify-center">
              <Wallet className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-black text-white">{displayVal(totalBankBalance)}</div>
            <span className="text-xs text-slate-500 mt-1 block">{accounts.length} conta(s) cadastradas</span>
          </div>
        </div>

        {/* Total Invested */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Carteira de Investimentos</span>
            <div className="w-8 h-8 rounded-xl bg-purple-500/10 text-purple-400 flex items-center justify-center">
              <PieIcon className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <div className="text-2xl font-black text-white">{displayVal(totalInvestments)}</div>
            <div className="flex items-center gap-1.5 mt-1">
              <span className={`text-xs font-semibold ${(portfolio?.total_profit_percentage || 0) >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                {hideValues ? '•••' : formatPercentage(portfolio?.total_profit_percentage || 0)}
              </span>
              <span className="text-xs text-slate-500">rentabilidade acumulada</span>
            </div>
          </div>
        </div>

        {/* Monthly Balance */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Resultado do Mês</span>
            <div className="w-8 h-8 rounded-xl bg-amber-500/10 text-amber-400 flex items-center justify-center">
              <Activity className="w-4 h-4" />
            </div>
          </div>
          <div className="mt-3">
            <div className={`text-2xl font-black ${parseFloat(summary?.net_balance || '0') >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
              {displayVal(summary?.net_balance || '0')}
            </div>
            <div className="flex items-center gap-3 mt-1 text-xs">
              <span className="text-emerald-400 flex items-center gap-0.5">
                <ArrowUpRight className="w-3 h-3" />
                {displayVal(summary?.total_income || '0')}
              </span>
              <span className="text-red-400 flex items-center gap-0.5">
                <ArrowDownRight className="w-3 h-3" />
                {displayVal(summary?.total_expense || '0')}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Visual Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Cash Flow Chart */}
        <div className="lg:col-span-2 bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-base font-bold text-white">Fluxo de Caixa Mensal</h2>
              <span className="text-xs text-slate-400">Comparativo entre receitas recebidas e despesas realizadas</span>
            </div>
          </div>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={monthlyFlowData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" opacity={0.5} />
                <XAxis dataKey="name" stroke="#94A3B8" fontSize={12} />
                <YAxis stroke="#94A3B8" fontSize={12} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '0.75rem', fontSize: '12px' }}
                  formatter={(val: any) => [displayVal(val), '']}
                />
                <Bar dataKey="Receitas" fill="#10B981" radius={[6, 6, 0, 0]} />
                <Bar dataKey="Despesas" fill="#EF4444" radius={[6, 6, 0, 0]} />
                <Bar dataKey="Líquido" fill="#3B82F6" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Portfolio Asset Allocation */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg flex flex-col justify-between">
          <div>
            <h2 className="text-base font-bold text-white mb-1">Alocação de Carteira</h2>
            <span className="text-xs text-slate-400">Distribuição patrimonial por classe de ativo</span>
          </div>
          <div className="h-48 my-auto">
            {allocationData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={allocationData} dataKey="value" nameKey="name" cx="50%" cy="50%" innerRadius={45} outerRadius={70} paddingAngle={4}>
                    {allocationData.map((_, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '0.75rem', fontSize: '12px' }}
                    formatter={(val: any) => [displayVal(val), '']}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div className="h-full flex items-center justify-center text-xs text-slate-500">
                Nenhum ativo em carteira ainda.
              </div>
            )}
          </div>
          <div className="grid grid-cols-2 gap-2 pt-2 border-t border-slate-800">
            {allocationData.map((item, idx) => (
              <div key={item.name} className="flex items-center gap-2 text-xs text-slate-300">
                <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: COLORS[idx % COLORS.length] }}></span>
                <span className="truncate">{item.name}:</span>
                <span className="font-semibold text-white ml-auto">{displayVal(item.value)}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Recent Transactions Section */}
      <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-base font-bold text-white">Últimos Lançamentos</h2>
            <span className="text-xs text-slate-400">Extrato recente de movimentações</span>
          </div>
          <Link to="/finances" className="text-xs font-semibold text-brand-500 hover:text-brand-400 transition-colors">
            Ver todas &rarr;
          </Link>
        </div>

        {transactions.length === 0 ? (
          <div className="py-8 text-center text-xs text-slate-500">
            Nenhuma transação registrada ainda. Clique em "Nova Transação" para começar.
          </div>
        ) : (
          <div className="divide-y divide-slate-800/60">
            {transactions.map((tx) => (
              <div key={tx.id} className="py-3.5 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div
                    className={`w-9 h-9 rounded-xl flex items-center justify-center text-xs font-bold ${
                      tx.type === 'income' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'
                    }`}
                  >
                    {tx.type === 'income' ? <ArrowUpRight className="w-4 h-4" /> : <ArrowDownRight className="w-4 h-4" />}
                  </div>
                  <div>
                    <span className="text-sm font-semibold text-slate-200 block">{tx.description}</span>
                    <span className="text-xs text-slate-500">
                      {new Date(tx.date).toLocaleDateString('pt-BR')} • {tx.category?.name || 'Geral'}
                    </span>
                  </div>
                </div>
                <div className={`text-sm font-bold ${tx.type === 'income' ? 'text-emerald-400' : 'text-slate-200'}`}>
                  {tx.type === 'expense' ? `-${displayVal(tx.amount)}` : `+${displayVal(tx.amount)}`}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
