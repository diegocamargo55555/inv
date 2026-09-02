import React, { useEffect, useState } from 'react'
import { Plus, CreditCard as CardIcon } from 'lucide-react'
import api from '../../services/api'
import { usePrivacyStore } from '../../stores/privacy'
import { formatCurrency, formatInvoiceMonth } from '../../utils/formatters'
import { Category, CreditCard, InvoiceSummary } from '../../types'

export const CreditCards: React.FC = () => {
  const { hideValues } = usePrivacyStore()
  const [cards, setCards] = useState<CreditCard[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [selectedCard, setSelectedCard] = useState<CreditCard | null>(null)
  const [invoice, setInvoice] = useState<InvoiceSummary | null>(null)
  const [monthYear, setMonthYear] = useState(() => {
    const now = new Date()
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  })

  // Modals
  const [isCardModalOpen, setIsCardModalOpen] = useState(false)
  const [isExpenseModalOpen, setIsExpenseModalOpen] = useState(false)

  // Card Form
  const [cardName, setCardName] = useState('')
  const [cardLimit, setCardLimit] = useState('')
  const [closingDay, setClosingDay] = useState('20')
  const [dueDay, setDueDay] = useState('27')
  const [brand, setBrand] = useState('Mastercard')
  const [color] = useState('#8B5CF6')

  // Expense Form
  const [expenseDesc, setExpenseDesc] = useState('')
  const [expenseAmount, setExpenseAmount] = useState('')
  const [installmentsCount, setInstallmentsCount] = useState('1')
  const [expenseCatId, setExpenseCatId] = useState('')
  const [expenseDate, setExpenseDate] = useState(new Date().toISOString().split('T')[0])

  const loadCards = async () => {
    try {
      const [cardsRes, catRes] = await Promise.all([
        api.get<CreditCard[]>('/cards'),
        api.get<Category[]>('/categories'),
      ])
      setCards(cardsRes.data)
      setCategories(catRes.data)
      if (cardsRes.data.length > 0 && !selectedCard) {
        setSelectedCard(cardsRes.data[0])
      }
    } catch (err) {
      console.error('Failed to load credit cards:', err)
    }
  }

  const loadInvoice = async (cardId: string, my: string) => {
    try {
      const res = await api.get<InvoiceSummary>(`/cards/${cardId}/invoice?month_year=${my}`)
      setInvoice(res.data)
    } catch (err) {
      console.error('Failed to load invoice:', err)
    }
  }

  useEffect(() => {
    loadCards()
  }, [])

  useEffect(() => {
    if (selectedCard) {
      loadInvoice(selectedCard.id, monthYear)
    }
  }, [selectedCard, monthYear])

  const handleCreateCard = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.post('/cards', {
        name: cardName,
        limit: parseFloat(cardLimit),
        closing_day: parseInt(closingDay, 10),
        due_day: parseInt(dueDay, 10),
        brand,
        color,
      })
      setIsCardModalOpen(false)
      setCardName('')
      setCardLimit('')
      loadCards()
    } catch (err) {
      alert('Erro ao criar cartão de crédito')
    }
  }

  const handleCreateExpense = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedCard) return

    try {
      await api.post('/cards/expense', {
        card_id: selectedCard.id,
        category_id: expenseCatId || null,
        total_amount: parseFloat(expenseAmount),
        installments_count: parseInt(installmentsCount, 10),
        purchase_date: new Date(expenseDate).toISOString(),
        description: expenseDesc,
      })
      setIsExpenseModalOpen(false)
      setExpenseDesc('')
      setExpenseAmount('')
      setInstallmentsCount('1')
      loadInvoice(selectedCard.id, monthYear)
    } catch (err) {
      alert('Erro ao lançar despesa no cartão')
    }
  }

  const displayVal = (val: string | number) => (hideValues ? '••••••' : formatCurrency(val))

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white tracking-tight">Cartões de Crédito & Faturas</h1>
          <p className="text-sm text-slate-400">Controle faturas mensais, melhor dia de compra e parcelamentos futuros.</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => setIsCardModalOpen(true)}
            className="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 flex items-center gap-1.5 transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Novo Cartão</span>
          </button>
          {selectedCard && (
            <button
              onClick={() => setIsExpenseModalOpen(true)}
              className="px-4 py-2 bg-brand-500 hover:bg-brand-600 active:scale-95 text-white text-xs font-semibold rounded-xl shadow-lg shadow-brand-500/20 flex items-center gap-1.5 transition-all"
            >
              <Plus className="w-4 h-4" />
              <span>Lançar Compra</span>
            </button>
          )}
        </div>
      </div>

      {/* Cards List Horizontal */}
      {cards.length === 0 ? (
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-8 text-center text-xs text-slate-400">
          Nenhum cartão cadastrado. Clique em "Novo Cartão" para começar.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {cards.map((c) => {
            const isSelected = selectedCard?.id === c.id
            return (
              <div
                key={c.id}
                onClick={() => setSelectedCard(c)}
                className={`cursor-pointer rounded-2xl p-5 border transition-all relative overflow-hidden ${
                  isSelected
                    ? 'bg-dark-900 border-brand-500 shadow-lg shadow-brand-500/10'
                    : 'bg-dark-900/60 border-slate-800 hover:border-slate-700'
                }`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl flex items-center justify-center text-white" style={{ backgroundColor: c.color }}>
                      <CardIcon className="w-5 h-5" />
                    </div>
                    <div>
                      <h3 className="text-sm font-bold text-white">{c.name}</h3>
                      <span className="text-xs text-slate-400">{c.brand}</span>
                    </div>
                  </div>
                  {isSelected && (
                    <span className="text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full bg-brand-500/20 text-brand-400 border border-brand-500/30">
                      Selecionado
                    </span>
                  )}
                </div>

                <div className="mt-4 pt-3 border-t border-slate-800 flex justify-between text-xs text-slate-400">
                  <div>
                    <span>Fecha dia: <strong className="text-slate-200">{c.closing_day}</strong></span>
                  </div>
                  <div>
                    <span>Vence dia: <strong className="text-slate-200">{c.due_day}</strong></span>
                  </div>
                  <div>
                    <span>Limite: <strong className="text-slate-200">{displayVal(c.limit)}</strong></span>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Invoice Management for Selected Card */}
      {selectedCard && (
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h2 className="text-lg font-bold text-white flex items-center gap-2">
                <span>Fatura de {selectedCard.name}</span>
                <span className="text-xs font-normal text-slate-400">({formatInvoiceMonth(monthYear)})</span>
              </h2>
              <span className="text-xs text-slate-400">Detalhamento dos gastos e parcelas programadas</span>
            </div>

            {/* Month Year Selector */}
            <div className="flex items-center gap-2">
              <span className="text-xs text-slate-400">Competência:</span>
              <input
                type="month"
                value={monthYear}
                onChange={(e) => setMonthYear(e.target.value)}
                className="px-3 py-1.5 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white focus:outline-none focus:border-brand-500"
              />
            </div>
          </div>

          {/* Invoice Summary Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="p-4 rounded-xl bg-dark-950 border border-slate-800">
              <span className="text-xs text-slate-400 block">Total da Fatura Atual</span>
              <span className="text-2xl font-bold text-red-400">{displayVal(invoice?.total_amount || '0')}</span>
            </div>
            <div className="p-4 rounded-xl bg-dark-950 border border-slate-800">
              <span className="text-xs text-slate-400 block">Limite Disponível Estimado</span>
              <span className="text-2xl font-bold text-emerald-400">
                {displayVal(Math.max(0, parseFloat(selectedCard.limit) - parseFloat(invoice?.total_amount || '0')))}
              </span>
            </div>
          </div>

          {/* Invoice Transactions Table */}
          <div>
            <h3 className="text-xs font-bold text-slate-300 uppercase tracking-wider mb-3">Lançamentos na Fatura</h3>
            {!invoice?.transactions || invoice.transactions.length === 0 ? (
              <div className="py-8 text-center text-xs text-slate-500 bg-dark-950 rounded-xl border border-slate-800/60">
                Nenhum lançamento nesta fatura.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
                      <th className="pb-3">Data</th>
                      <th className="pb-3">Descrição</th>
                      <th className="pb-3">Categoria</th>
                      <th className="pb-3">Parcela</th>
                      <th className="pb-3 text-right">Valor</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/50">
                    {invoice.transactions.map((tx) => (
                      <tr key={tx.id} className="hover:bg-slate-800/30 transition-colors">
                        <td className="py-3 text-slate-400">{new Date(tx.date).toLocaleDateString('pt-BR')}</td>
                        <td className="py-3 font-semibold text-slate-200">{tx.description}</td>
                        <td className="py-3 text-slate-400">{tx.category?.name || 'Geral'}</td>
                        <td className="py-3 text-slate-400">
                          {tx.total_installments > 1 ? (
                            <span className="px-2 py-0.5 bg-slate-800 rounded text-slate-300 font-mono">
                              {tx.installment_number}/{tx.total_installments}
                            </span>
                          ) : (
                            'À vista'
                          )}
                        </td>
                        <td className="py-3 text-right font-bold text-slate-200">{displayVal(tx.amount)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* New Card Modal */}
      {isCardModalOpen && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-lg font-bold text-white">Novo Cartão de Crédito</h3>
            <form onSubmit={handleCreateCard} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-slate-300">Nome / Identificação</label>
                <input
                  type="text"
                  required
                  placeholder="Ex: Nubank Ultravioleta, C6 Carbon"
                  value={cardName}
                  onChange={(e) => setCardName(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Limite Total (R$)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="Ex: 5000,00"
                    value={cardLimit}
                    onChange={(e) => setCardLimit(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Bandeira</label>
                  <select
                    value={brand}
                    onChange={(e) => setBrand(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    <option value="Mastercard">Mastercard</option>
                    <option value="Visa">Visa</option>
                    <option value="Elo">Elo</option>
                    <option value="Amex">American Express</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Dia de Fechamento</label>
                  <input
                    type="number"
                    min="1"
                    max="31"
                    required
                    value={closingDay}
                    onChange={(e) => setClosingDay(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Dia de Vencimento</label>
                  <input
                    type="number"
                    min="1"
                    max="31"
                    required
                    value={dueDay}
                    onChange={(e) => setDueDay(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCardModalOpen(false)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 text-white text-xs font-semibold rounded-xl"
                >
                  Salvar Cartão
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* New Card Expense Modal */}
      {isExpenseModalOpen && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 max-w-md w-full shadow-2xl space-y-4">
            <h3 className="text-lg font-bold text-white">Lançar Compra no Cartão</h3>
            <form onSubmit={handleCreateExpense} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-slate-300">Descrição</label>
                <input
                  type="text"
                  required
                  placeholder="Ex: Passagens Aéreas, Smartphone"
                  value={expenseDesc}
                  onChange={(e) => setExpenseDesc(e.target.value)}
                  className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Valor Total (R$)</label>
                  <input
                    type="number"
                    step="0.01"
                    required
                    placeholder="0,00"
                    value={expenseAmount}
                    onChange={(e) => setExpenseAmount(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Parcelas</label>
                  <select
                    value={installmentsCount}
                    onChange={(e) => setInstallmentsCount(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 18, 24].map((num) => (
                      <option key={num} value={num}>
                        {num === 1 ? '1x (À vista)' : `${num}x parcelado`}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs font-medium text-slate-300">Data da Compra</label>
                  <input
                    type="date"
                    required
                    value={expenseDate}
                    onChange={(e) => setExpenseDate(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-slate-300">Categoria</label>
                  <select
                    value={expenseCatId}
                    onChange={(e) => setExpenseCatId(e.target.value)}
                    className="w-full px-3 py-2 bg-dark-950 border border-slate-800 rounded-xl text-xs text-white"
                  >
                    <option value="">Selecione...</option>
                    {categories.filter((c) => c.type === 'expense').map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsExpenseModalOpen(false)}
                  className="flex-1 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold rounded-xl"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 bg-brand-500 hover:bg-brand-600 text-white text-xs font-semibold rounded-xl"
                >
                  Confirmar Compra
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
