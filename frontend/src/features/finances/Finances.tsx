import React, { useEffect, useState } from 'react'
import { Plus, Wallet, ArrowUpRight, ArrowDownRight, Tag, AlertTriangle, CheckCircle, Search, Pencil, Trash2 } from 'lucide-react'
import api from '../../services/api'
import { usePrivacyStore } from '../../stores/privacy'
import { formatCurrency } from '../../utils/formatters'
import { Account, BudgetProgress, Category, Transaction } from '../../types'

export const Finances: React.FC = () => {
  const { hideValues } = usePrivacyStore()
  const [accounts, setAccounts] = useState<Account[]>([])
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [budgets, setBudgets] = useState<BudgetProgress[]>([])

  // Filters
  const [filterType, setFilterType] = useState<'all' | 'income' | 'expense'>('all')
  const [search, setSearch] = useState('')

  // Modals state
  const [isTxModalOpen, setIsTxModalOpen] = useState(false)
  const [isEditTxModalOpen, setIsEditTxModalOpen] = useState(false)
  const [editingTx, setEditingTx] = useState<Transaction | null>(null)
  const [isAccModalOpen, setIsAccModalOpen] = useState(false)
  const [isBudgetModalOpen, setIsBudgetModalOpen] = useState(false)

  // New Tx Form
  const [txDesc, setTxDesc] = useState('')
  const [txAmount, setTxAmount] = useState('')
  const [txType, setTxType] = useState<'income' | 'expense'>('expense')
  const [txAccId, setTxAccId] = useState('')
  const [txCatId, setTxCatId] = useState('')
  const [txDate, setTxDate] = useState(new Date().toISOString().split('T')[0])

  // Edit Tx Form
  const [editTxDesc, setEditTxDesc] = useState('')
  const [editTxAmount, setEditTxAmount] = useState('')
  const [editTxType, setEditTxType] = useState<'income' | 'expense'>('expense')
  const [editTxAccId, setEditTxAccId] = useState('')
  const [editTxCatId, setEditTxCatId] = useState('')
  const [editTxDate, setEditTxDate] = useState('')

  // New Account Form
  const [accName, setAccName] = useState('')
  const [accType, setAccType] = useState('checking')
  const [accBalance, setAccBalance] = useState('')
  const [accInstitution, setAccInstitution] = useState('')
  const [accColor] = useState('#10B981')

  // New Budget Form
  const [budgetCatId, setBudgetCatId] = useState('')
  const [budgetLimit, setBudgetLimit] = useState('')

  const loadData = async () => {
    try {
      const [accRes, txRes, catRes, budRes] = await Promise.all([
        api.get<Account[]>('/accounts'),
        api.get<Transaction[]>('/transactions?limit=100'),
        api.get<Category[]>('/categories'),
        api.get<BudgetProgress[]>('/budgets'),
      ])
      setAccounts(accRes.data)
      setTransactions(txRes.data)
      setCategories(catRes.data)
      setBudgets(budRes.data)
      if (accRes.data.length > 0 && !txAccId) setTxAccId(accRes.data[0].id)
    } catch (err) {
      console.error('Failed to load finances data:', err)
    }
  }

  useEffect(() => {
    loadData()
  }, [])

  const handleCreateTx = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/transactions', {
        description: txDesc,
        amount: parseFloat(txAmount),
        type: txType,
        account_id: txAccId || null,
        category_id: txCatId || null,
        date: new Date(txDate).toISOString(),
      })
      setIsTxModalOpen(false)
      setTxDesc('')
      setTxAmount('')
      loadData()
    } catch (err) {
      alert('Erro ao criar transação')
    }
  }

  const handleOpenEditTx = (tx: Transaction) => {
    setEditingTx(tx)
    setEditTxDesc(tx.description)
    setEditTxAmount(tx.amount)
    setEditTxType(tx.type === 'income' ? 'income' : 'expense')
    setEditTxAccId(tx.account_id || '')
    setEditTxCatId(tx.category_id || '')
    setEditTxDate(tx.date ? tx.date.split('T')[0] : new Date().toISOString().split('T')[0])
    setIsEditTxModalOpen(true)
  }

  const handleUpdateTx = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingTx) return
    try {
      await api.put(`/transactions/${editingTx.id}`, {
        description: editTxDesc,
        amount: parseFloat(editTxAmount),
        type: editTxType,
        account_id: editTxAccId || null,
        category_id: editTxCatId || null,
        date: new Date(editTxDate).toISOString(),
      })
      setIsEditTxModalOpen(false)
      setEditingTx(null)
      loadData()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Erro ao atualizar transação')
    }
  }

  const handleDeleteTx = async (tx: Transaction) => {
    if (!window.confirm(`Deseja realmente excluir o lançamento "${tx.description}"? O saldo da conta será ajustado.`)) return
    try {
      await api.delete(`/transactions/${tx.id}`)
      loadData()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Erro ao excluir transação')
    }
  }

  const handleCreateAccount = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/accounts', {
        name: accName,
        type: accType,
        initial_balance: parseFloat(accBalance || '0'),
        institution: accInstitution,
        color: accColor,
      })
      setIsAccModalOpen(false)
      setAccName('')
      setAccBalance('')
      loadData()
    } catch (err) {
      alert('Erro ao criar conta')
    }
  }

  const handleSetBudget = async (e: React.FormEvent) => {
    e.preventDefault()
    const now = new Date()
    const monthYear = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    try {
      await api.post('/budgets', {
        category_id: budgetCatId || null,
        month_year: monthYear,
        amount_limit: parseFloat(budgetLimit),
      })
      setIsBudgetModalOpen(false)
      setBudgetLimit('')
      loadData()
    } catch (err) {
      alert('Erro ao definir orçamento')
    }
  }

  const displayVal = (val: string | number) => (hideValues ? '••••••' : formatCurrency(val))

  const filteredTransactions = transactions.filter((tx) => {
    if (filterType !== 'all' && tx.type !== filterType) return false
    if (search && !tx.description.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Finanças & Orçamentos</h1>
          <p className="text-sm text-slate-400">Gerencie contas bancárias, transações diárias e limites de orçamento.</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => setIsAccModalOpen(true)}
            className="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 flex items-center gap-1.5 transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Nova Conta</span>
          </button>
          <button
            onClick={() => setIsTxModalOpen(true)}
            className="px-4 py-2 bg-brand-500 hover:bg-brand-600 active:scale-95 text-white text-xs font-semibold rounded-xl shadow-lg shadow-brand-500/20 flex items-center gap-1.5 transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>Nova Transação</span>
          </button>
        </div>
      </div>

      {/* Bank Accounts Grid */}
      <div>
        <h2 className="text-sm font-bold text-slate-300 uppercase tracking-wider mb-3">Minhas Contas</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {accounts.map((acc) => (
            <div key={acc.id} className="bg-dark-900 border border-slate-800 rounded-2xl p-4 shadow-md flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div className="w-8 h-8 rounded-lg flex items-center justify-center text-white" style={{ backgroundColor: acc.color }}>
                    <Wallet className="w-4 h-4" />
                  </div>
                  <div>
                    <h3 className="text-sm font-semibold text-slate-200">{acc.name}</h3>
                    <span className="text-xs text-slate-500 capitalize">{acc.institution || acc.type}</span>
                  </div>
                </div>
              </div>
              <div className="mt-4 pt-3 border-t border-slate-800/80">
                <span className="text-xs text-slate-400 block">Saldo Atual</span>
                <span className="text-lg font-bold text-white">{displayVal(acc.balance)}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Budgets Section */}
      <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-base font-bold text-white">Planejamento & Metas de Orçamento</h2>
            <span className="text-xs text-slate-400">Acompanhe limites de gastos por categoria no mês</span>
          </div>
          <button
            onClick={() => setIsBudgetModalOpen(true)}
            className="text-xs font-semibold text-brand-500 hover:text-brand-400 flex items-center gap-1 transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Definir Teto</span>
          </button>
        </div>

        {budgets.length === 0 ? (
          <div className="py-6 text-center text-xs text-slate-500">
            Nenhum orçamento configurado para este mês. Clique em "Definir Teto" para controlar seus gastos.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {budgets.map((b, idx) => (
              <div key={idx} className="p-4 rounded-xl bg-dark-950 border border-slate-800 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-semibold text-slate-200">{b.budget.category?.name || 'Geral'}</span>
                  {b.is_exceeded ? (
                    <span className="flex items-center gap-1 text-xs text-red-400 font-bold bg-red-500/10 px-2 py-0.5 rounded-full border border-red-500/20">
                      <AlertTriangle className="w-3 h-3" /> Estourado
                    </span>
                  ) : b.is_alert ? (
                    <span className="flex items-center gap-1 text-xs text-amber-400 font-bold bg-amber-500/10 px-2 py-0.5 rounded-full border border-amber-500/20">
                      <AlertTriangle className="w-3 h-3" /> Atenção 80%
                    </span>
                  ) : (
                    <span className="flex items-center gap-1 text-xs text-emerald-400 font-bold bg-emerald-500/10 px-2 py-0.5 rounded-full border border-emerald-500/20">
                      <CheckCircle className="w-3 h-3" /> No Alvo
                    </span>
                  )}
                </div>

                <div className="w-full bg-slate-800 rounded-full h-2 overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all ${
                      b.is_exceeded ? 'bg-red-500' : b.is_alert ? 'bg-amber-500' : 'bg-brand-500'
                    }`}
                    style={{ width: `${Math.min(b.percentage, 100)}%` }}
                  ></div>
                </div>

                <div className="flex justify-between text-xs text-slate-400">
                  <span>Gasto: <strong className="text-slate-200">{displayVal(b.spent)}</strong></span>
                  <span>Limite: <strong className="text-slate-200">{displayVal(b.budget.amount_limit)}</strong></span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Transactions Section */}
      <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h2 className="text-base font-bold text-white">Extrato de Transações</h2>
            <span className="text-xs text-slate-400">Histórico de receitas e despesas</span>
          </div>

          <div className="flex items-center gap-3">
            {/* Search */}
            <div className="relative">
              <Search className="w-3.5 h-3.5 absolute left-3 top-2.5 text-slate-500" />
              <input
                type="text"
                placeholder="Buscar lançamentos..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-8 pr-3 py-1.5 bg-dark-950 border border-slate-800 rounded-xl text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-brand-500"
              />
            </div>

            {/* Filter Tabs */}
            <div className="flex bg-dark-950 p-1 rounded-xl border border-slate-800">
              <button
                onClick={() => setFilterType('all')}
                className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                  filterType === 'all' ? 'bg-slate-800 text-white' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Todas
              </button>
              <button
                onClick={() => setFilterType('income')}
                className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                  filterType === 'income' ? 'bg-emerald-500/20 text-emerald-400' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Receitas
              </button>
              <button
                onClick={() => setFilterType('expense')}
                className={`px-3 py-1 rounded-lg text-xs font-semibold transition-all ${
                  filterType === 'expense' ? 'bg-red-500/20 text-red-400' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                Despesas
              </button>
            </div>
          </div>
        </div>

        {/* Transactions Table */}
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
                <th className="pb-3">Data</th>
                <th className="pb-3">Descrição</th>
                <th className="pb-3">Categoria</th>
                <th className="pb-3">Conta</th>
                <th className="pb-3 text-right">Valor</th>
                <th className="pb-3 text-center">Ações</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/50">
              {filteredTransactions.map((tx) => (
                <tr key={tx.id} className="hover:bg-slate-800/30 transition-colors">
                  <td className="py-3.5 text-slate-400">{new Date(tx.date).toLocaleDateString('pt-BR')}</td>
                  <td className="py-3.5 font-semibold text-slate-200 flex items-center gap-2">
                    <div
                      className={`w-6 h-6 rounded-md flex items-center justify-center ${
                        tx.type === 'income' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'
                      }`}
                    >
                      {tx.type === 'income' ? <ArrowUpRight className="w-3.5 h-3.5" /> : <ArrowDownRight className="w-3.5 h-3.5" />}
                    </div>
                    <span>{tx.description}</span>
                  </td>
                  <td className="py-3.5">
                    <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-slate-800 text-slate-300">
                      <Tag className="w-3 h-3 text-slate-500" />
                      {tx.category?.name || 'Geral'}
                    </span>
                  </td>
                  <td className="py-3.5 text-slate-400">{tx.account?.name || tx.credit_card?.name || 'Conta Padrão'}</td>
                  <td className={`py-3.5 text-right font-bold ${tx.type === 'income' ? 'text-emerald-400' : 'text-slate-200'}`}>
                    {tx.type === 'expense' ? `-${displayVal(tx.amount)}` : `+${displayVal(tx.amount)}`}
                  </td>
                  <td className="py-3.5 text-center">
                    <div className="flex items-center justify-center gap-1.5">
                      <button
                        onClick={() => handleOpenEditTx(tx)}
                        title="Editar Lançamento"
                        className="p-1.5 rounded-lg bg-slate-800 hover:bg-brand-500/20 text-slate-300 hover:text-brand-400 border border-slate-700 transition-colors"
                      >
                        <Pencil className="w-3.5 h-3.5" />
                      </button>
                      <button
                        onClick={() => handleDeleteTx(tx)}
                        title="Excluir Lançamento"
                        className="p-1.5 rounded-lg bg-slate-800 hover:bg-red-500/20 text-slate-300 hover:text-red-400 border border-slate-700 transition-colors"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Transaction Modal */}
      {isTxModalOpen && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-lg font-bold text-white">Novo Lançamento</h3>
            <form onSubmit={handleCreateTx} className="space-y-4">
              <div className="grid grid-cols-2 gap-2 p-1 bg-dark-950 rounded-xl border border-slate-800">
                <button
                  type="button"
                  onClick={() => setTxType('expense')}
                  className={`py-1.5 rounded-lg text-xs font-semibold ${txType === 'expense' ? 'bg-red-500 text-white' : 'text-slate-400'}`}
                >
                  Despesa
                </button>
                <button
                  type="button"
                  onClick={() => setTxType('income')}
                  className={`py-1.5 rounded-lg text-xs font-semibold ${txType === 'income' ? 'bg-emerald-500 text-white' : 'text-slate-400'}`}
                >
                  Receita
                </button>
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">Descrição</label>
                <input
                  type="text"
                  required
                  placeholder="Ex: Supermercado, Aluguel, Salário"
                  value={txDesc}
                  onChange={(e) => setTxDesc(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Valor (R$)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0,00"
                    value={txAmount}
                    onChange={(e) => setTxAmount(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Data</label>
                  <input
                    type="date"
                    required
                    value={txDate}
                    onChange={(e) => setTxDate(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Conta de Débito/Crédito</label>
                  <select
                    value={txAccId}
                    onChange={(e) => setTxAccId(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    {accounts.map((acc) => (
                      <option key={acc.id} value={acc.id}>{acc.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Categoria</label>
                  <select
                    value={txCatId}
                    onChange={(e) => setTxCatId(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    <option value="">Selecione uma categoria...</option>
                    {categories.filter((c) => c.type === txType).map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsTxModalOpen(false)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 text-white text-xs font-semibold rounded-xl"
                >
                  Salvar Lançamento
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Account Modal */}
      {isAccModalOpen && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-lg font-bold text-white">Nova Conta Bancária / Carteira</h3>
            <form onSubmit={handleCreateAccount} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-slate-300">Nome da Conta</label>
                <input
                  type="text"
                  required
                  placeholder="Ex: Nubank, Itaú, Carteira Física"
                  value={accName}
                  onChange={(e) => setAccName(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Tipo</label>
                  <select
                    value={accType}
                    onChange={(e) => setAccType(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    <option value="checking">Conta Corrente</option>
                    <option value="savings">Poupança</option>
                    <option value="investment">Investimento</option>
                    <option value="cash">Dinheiro em Espécie</option>
                  </select>
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Saldo Inicial (R$)</label>
                  <input
                    type="number"
                    step="0.01"
                    placeholder="0,00"
                    value={accBalance}
                    onChange={(e) => setAccBalance(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">Instituição</label>
                <input
                  type="text"
                  placeholder="Ex: Nubank, Banco do Brasil"
                  value={accInstitution}
                  onChange={(e) => setAccInstitution(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsAccModalOpen(false)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 text-white text-xs font-semibold rounded-xl"
                >
                  Cadastrar Conta
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Budget Modal */}
      {isBudgetModalOpen && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-lg font-bold text-white">Definir Orçamento de Gastos</h3>
            <form onSubmit={handleSetBudget} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-slate-300">Categoria</label>
                <select
                  value={budgetCatId}
                  onChange={(e) => setBudgetCatId(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                >
                  <option value="">Geral (Todas as despesas)</option>
                  {categories.filter((c) => c.type === 'expense').map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">Teto Limite Mensal (R$)</label>
                <input
                  type="number"
                  step="0.01"
                  required
                  placeholder="Ex: 1500,00"
                  value={budgetLimit}
                  onChange={(e) => setBudgetLimit(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsBudgetModalOpen(false)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 text-white text-xs font-semibold rounded-xl"
                >
                  Salvar Teto
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Transaction Modal */}
      {isEditTxModalOpen && editingTx && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-lg font-bold text-white flex items-center gap-2">
              <Pencil className="w-4 h-4 text-brand-400" />
              <span>Editar Lançamento</span>
            </h3>
            <form onSubmit={handleUpdateTx} className="space-y-4">
              <div className="grid grid-cols-2 gap-2 p-1 bg-dark-950 rounded-xl border border-slate-800">
                <button
                  type="button"
                  onClick={() => setEditTxType('expense')}
                  className={`py-1.5 rounded-lg text-xs font-semibold ${editTxType === 'expense' ? 'bg-red-500 text-white' : 'text-slate-400'}`}
                >
                  Despesa / Compra
                </button>
                <button
                  type="button"
                  onClick={() => setEditTxType('income')}
                  className={`py-1.5 rounded-lg text-xs font-semibold ${editTxType === 'income' ? 'bg-emerald-500 text-white' : 'text-slate-400'}`}
                >
                  Receita / Venda
                </button>
              </div>

              <div>
                <label className="text-xs font-medium text-slate-300">Descrição</label>
                <input
                  type="text"
                  required
                  value={editTxDesc}
                  onChange={(e) => setEditTxDesc(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Valor (R$)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    value={editTxAmount}
                    onChange={(e) => setEditTxAmount(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Data</label>
                  <input
                    type="date"
                    required
                    value={editTxDate}
                    onChange={(e) => setEditTxDate(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Conta de Débito/Crédito</label>
                  <select
                    value={editTxAccId}
                    onChange={(e) => setEditTxAccId(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    <option value="">Sem conta vinculada</option>
                    {accounts.map((acc) => (
                      <option key={acc.id} value={acc.id}>{acc.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Categoria</label>
                  <select
                    value={editTxCatId}
                    onChange={(e) => setEditTxCatId(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    <option value="">Sem categoria</option>
                    {categories.filter((c) => c.type === editTxType).map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsEditTxModalOpen(false)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 text-white text-xs font-semibold rounded-xl"
                >
                  Salvar Alterações
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
