import React, { useEffect, useState, useRef } from 'react'
import { Plus, RefreshCw, Search, Check, Sparkles, AlertCircle, ArrowDownRight, ArrowUpRight, TrendingDown, Pencil, Trash2, History, Layers } from 'lucide-react'
import api from '../../services/api'
import { usePrivacyStore } from '../../stores/privacy'
import { formatCurrency, formatPercentage } from '../../utils/formatters'
import { Account, Asset, PortfolioSummary, PositionSummary, InvestmentTransaction } from '../../types'

export const Investments: React.FC = () => {
  const { hideValues } = usePrivacyStore()
  const [accounts, setAccounts] = useState<Account[]>([])
  const [selectedPortId, setSelectedPortId] = useState<string>('')
  const [summary, setSummary] = useState<PortfolioSummary | null>(null)
  const [orders, setOrders] = useState<InvestmentTransaction[]>([])
  const [availableAssets, setAvailableAssets] = useState<Asset[]>([])
  const [syncingQuote, setSyncingQuote] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<'positions' | 'orders'>('positions')

  // Order Modal (New Buy / Sell)
  const [isOrderModalOpen, setIsOrderModalOpen] = useState(false)
  const [orderType, setOrderType] = useState<'buy' | 'sell'>('buy')
  const [tickerInput, setTickerInput] = useState('')
  const [selectedAsset, setSelectedAsset] = useState<Asset | null>(null)
  const [selectedAccountId, setSelectedAccountId] = useState<string>('')
  const [orderQty, setOrderQty] = useState('')
  const [orderPrice, setOrderPrice] = useState('')
  const [orderFees, setOrderFees] = useState('0')
  const [orderDate, setOrderDate] = useState(new Date().toISOString().split('T')[0])
  const [orderNotes, setOrderNotes] = useState('')
  const [isDropdownOpen, setIsDropdownOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [isSearching, setIsSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<Asset[]>([])
  const [isFetchingLiveQuote, setIsFetchingLiveQuote] = useState(false)
  const [liveQuoteBadge, setLiveQuoteBadge] = useState<string | null>(null)
  const [formError, setFormError] = useState('')

  // Edit Order Modal
  const [isEditOrderModalOpen, setIsEditOrderModalOpen] = useState(false)
  const [editingOrder, setEditingOrder] = useState<InvestmentTransaction | null>(null)
  const [editQty, setEditQty] = useState('')
  const [editPrice, setEditPrice] = useState('')
  const [editFees, setEditFees] = useState('0')
  const [editDate, setEditDate] = useState('')
  const [editNotes, setEditNotes] = useState('')
  const [editAccountId, setEditAccountId] = useState('')
  const [editError, setEditError] = useState('')
  const [isEditSubmitting, setIsEditSubmitting] = useState(false)

  // Specific state for Sell Operations
  const [availableCustodyQty, setAvailableCustodyQty] = useState<number | null>(null)
  const [custodyAveragePrice, setCustodyAveragePrice] = useState<number | null>(null)

  const dropdownRef = useRef<HTMLDivElement>(null)

  const fetchLiveQuoteFromBrapi = async (ticker: string) => {
    const clean = ticker.trim().toUpperCase()
    if (!clean || clean.length < 3) return
    setIsFetchingLiveQuote(true)
    setLiveQuoteBadge(null)
    try {
      const res = await api.get<{ ticker: string; current_price: string }>(`/investments/quote?ticker=${clean}`)
      if (res.data?.current_price && parseFloat(res.data.current_price) > 0) {
        setOrderPrice(res.data.current_price)
        setLiveQuoteBadge(`Brapi B3: ${formatCurrency(res.data.current_price, 'BRL')}`)
      }
    } catch (err) {
      console.log('Quote fetch fallback', err)
    } finally {
      setIsFetchingLiveQuote(false)
    }
  }

  const loadData = async () => {
    try {
      const [portsRes, assetsRes, accsRes] = await Promise.all([
        api.get<any[]>('/investments/portfolios'),
        api.get<Asset[]>('/investments/assets/search'),
        api.get<Account[]>('/accounts'),
      ])
      setAvailableAssets(assetsRes.data)
      setSearchResults(assetsRes.data)
      setAccounts(accsRes.data)
      if (accsRes.data.length > 0 && !selectedAccountId) {
        setSelectedAccountId(accsRes.data[0].id)
      }

      const portId = selectedPortId || (portsRes.data.length > 0 ? portsRes.data[0].id : '')
      if (portId) {
        setSelectedPortId(portId)
        const [sumRes, ordersRes] = await Promise.all([
          api.get<PortfolioSummary>(`/investments/portfolios/${portId}/summary`),
          api.get<InvestmentTransaction[]>(`/investments/portfolios/${portId}/orders`),
        ])
        setSummary(sumRes.data)
        setOrders(ordersRes.data || [])
      }
    } catch (err) {
      console.error('Failed to load investments:', err)
    }
  }

  useEffect(() => {
    loadData()
  }, [selectedPortId])

  // Debounced live search across Brapi and database as user types
  useEffect(() => {
    const query = tickerInput.trim().toUpperCase()
    if (!query) {
      setSearchResults(availableAssets)
      return
    }

    const timer = setTimeout(async () => {
      setIsSearching(true)
      try {
        const res = await api.get<Asset[]>(`/investments/assets/search?q=${query}`)
        if (res.data && res.data.length > 0) {
          setSearchResults(res.data)
        } else {
          setSearchResults(
            availableAssets.filter(
              (a) => a.ticker.toUpperCase().includes(query) || a.name.toUpperCase().includes(query)
            )
          )
        }
      } catch {
        setSearchResults(
          availableAssets.filter(
            (a) => a.ticker.toUpperCase().includes(query) || a.name.toUpperCase().includes(query)
          )
        )
      } finally {
        setIsSearching(false)
      }
    }, 250)

    return () => clearTimeout(timer)
  }, [tickerInput, availableAssets])

  // Close dropdown on click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleOpenOrderModal = (type: 'buy' | 'sell') => {
    setOrderType(type)
    setTickerInput('')
    setSelectedAsset(null)
    setOrderQty('')
    setOrderPrice('')
    setOrderFees('0')
    setOrderNotes('')
    setFormError('')
    setLiveQuoteBadge(null)
    setAvailableCustodyQty(null)
    setCustodyAveragePrice(null)
    if (accounts.length > 0) {
      setSelectedAccountId((prev) => (prev && accounts.some((a) => a.id === prev) ? prev : accounts[0].id))
    } else {
      setSelectedAccountId('')
    }
    setIsOrderModalOpen(true)
  }

  const handleOpenSellModalForPosition = (pos: PositionSummary) => {
    setOrderType('sell')
    setSelectedAsset(pos.position.asset || null)
    setTickerInput(pos.position.asset?.ticker || '')
    const qtyNum = parseFloat(pos.position.quantity)
    setAvailableCustodyQty(qtyNum)
    setCustodyAveragePrice(parseFloat(pos.position.average_price))
    setOrderQty(String(qtyNum))
    setOrderPrice(String(pos.current_price || pos.position.average_price))
    setOrderFees('0')
    setOrderNotes('Venda de posição em carteira')
    setFormError('')
    setLiveQuoteBadge(`Brapi B3: ${formatCurrency(pos.current_price, pos.position.asset?.currency)}`)
    setIsOrderModalOpen(true)
    if (pos.position.asset?.ticker) {
      fetchLiveQuoteFromBrapi(pos.position.asset.ticker)
    }
  }

  const handleSelectAsset = (asset: Asset) => {
    setSelectedAsset(asset)
    setTickerInput(asset.ticker)
    setOrderPrice(asset.current_price || '0')
    setIsDropdownOpen(false)
    setFormError('')

    // Check if this asset exists in current positions
    const existingPos = summary?.positions.find((p) => p.position.asset_id === asset.id)
    if (existingPos) {
      setAvailableCustodyQty(parseFloat(existingPos.position.quantity))
      setCustodyAveragePrice(parseFloat(existingPos.position.average_price))
    } else {
      setAvailableCustodyQty(null)
      setCustodyAveragePrice(null)
    }

    fetchLiveQuoteFromBrapi(asset.ticker)
  }

  const handleExecuteOrder = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')

    if (!selectedPortId) {
      setFormError('Selecione uma carteira.')
      return
    }

    const cleanTicker = tickerInput.trim().toUpperCase()
    if (!cleanTicker) {
      setFormError('Digite o código ou ticker do ativo.')
      return
    }

    const qtyNum = parseFloat(orderQty)
    const priceNum = parseFloat(orderPrice)

    if (isNaN(qtyNum) || qtyNum <= 0) {
      setFormError('Informe uma quantidade válida maior que zero.')
      return
    }
    if (isNaN(priceNum) || priceNum <= 0) {
      setFormError('Informe um preço unitário válido maior que zero.')
      return
    }

    // Validate sell quantity against custody
    if (orderType === 'sell' && availableCustodyQty !== null && qtyNum > availableCustodyQty) {
      setFormError(`Quantidade máxima para venda é de ${availableCustodyQty} cotas/ações em custódia.`)
      return
    }

    // Validate account debit for buy (aporte)
    if (orderType === 'buy') {
      if (accounts.length === 0) {
        setFormError('Você precisa cadastrar uma conta bancária em Finanças antes de realizar um aporte para que o valor possa ser debitado.')
        return
      }
      if (!selectedAccountId) {
        setFormError('Selecione a conta bancária para debitar o valor do aporte.')
        return
      }
    }

    setIsSubmitting(true)

    try {
      let targetAsset = selectedAsset

      // If user typed a ticker that wasn't previously selected or registered
      if (!targetAsset || targetAsset.ticker.toUpperCase() !== cleanTicker) {
        const assetRes = await api.post<Asset>('/investments/assets', {
          ticker: cleanTicker,
        })
        targetAsset = assetRes.data
        setSelectedAsset(targetAsset)
      }

      const endpoint = orderType === 'buy' ? '/investments/orders/buy' : '/investments/orders/sell'
      await api.post(endpoint, {
        portfolio_id: selectedPortId,
        asset_id: targetAsset.id,
        account_id: selectedAccountId || undefined,
        quantity: qtyNum,
        unit_price: priceNum,
        fees: parseFloat(orderFees || '0'),
        date: new Date(orderDate).toISOString(),
        notes: orderNotes,
      })

      setIsOrderModalOpen(false)
      setTickerInput('')
      setSelectedAsset(null)
      setOrderQty('')
      setOrderPrice('')
      setOrderFees('0')
      setOrderNotes('')
      loadData()
    } catch (err: any) {
      setFormError(err.response?.data?.error || 'Erro ao executar ordem. Verifique os valores.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleOpenEditOrder = (order: InvestmentTransaction) => {
    setEditingOrder(order)
    setEditQty(order.quantity)
    setEditPrice(order.unit_price)
    setEditFees(order.fees || '0')
    setEditDate(order.date ? order.date.split('T')[0] : new Date().toISOString().split('T')[0])
    setEditNotes(order.notes || '')
    setEditAccountId(order.account_id || '')
    setEditError('')
    setIsEditOrderModalOpen(true)
  }

  const handleSaveEditOrder = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!editingOrder) return
    setEditError('')

    const qtyNum = parseFloat(editQty)
    const priceNum = parseFloat(editPrice)
    if (isNaN(qtyNum) || qtyNum <= 0) {
      setEditError('Informe uma quantidade válida maior que zero.')
      return
    }
    if (isNaN(priceNum) || priceNum <= 0) {
      setEditError('Informe um preço unitário válido maior que zero.')
      return
    }

    setIsEditSubmitting(true)
    try {
      await api.put(`/investments/orders/${editingOrder.id}`, {
        quantity: qtyNum,
        unit_price: priceNum,
        fees: parseFloat(editFees || '0'),
        date: new Date(editDate).toISOString(),
        notes: editNotes,
        account_id: editAccountId || null,
      })
      setIsEditOrderModalOpen(false)
      setEditingOrder(null)
      await loadData()
    } catch (err: any) {
      setEditError(err.response?.data?.error || 'Erro ao salvar alterações na ordem.')
    } finally {
      setIsEditSubmitting(false)
    }
  }

  const handleDeleteOrder = async (order: InvestmentTransaction) => {
    const isBuy = order.type === 'buy'
    const ticker = order.asset?.ticker || 'ativo'
    const confirmMsg = `Deseja realmente excluir esta operação de ${isBuy ? 'COMPRA' : 'VENDA'} de ${ticker}?\n\nO preço médio e a carteira serão recalculados automaticamente.`
    if (!window.confirm(confirmMsg)) return

    try {
      await api.delete(`/investments/orders/${order.id}`)
      await loadData()
    } catch (err: any) {
      alert(err.response?.data?.error || 'Erro ao excluir ordem.')
    }
  }

  const handleUpdateQuote = async (assetId: string) => {
    setSyncingQuote(assetId)
    try {
      await api.post(`/investments/assets/${assetId}/quote`)
      await loadData()
    } catch (err) {
      console.error('Failed to sync quote:', err)
    } finally {
      setSyncingQuote(null)
    }
  }

  const displayVal = (val: string | number, currency = 'BRL') =>
    hideValues ? '••••••' : formatCurrency(val, currency)

  // Calculations for preview in modal
  const qtyVal = parseFloat(orderQty) || 0
  const priceVal = parseFloat(orderPrice) || 0
  const feesVal = parseFloat(orderFees) || 0
  const totalOrderValue = qtyVal * priceVal + (orderType === 'buy' ? feesVal : -feesVal)
  const estimatedSellPnL =
    orderType === 'sell' && custodyAveragePrice !== null
      ? (priceVal - custodyAveragePrice) * qtyVal - feesVal
      : null
  const estimatedSellPnLPercent =
    orderType === 'sell' && custodyAveragePrice !== null && custodyAveragePrice > 0
      ? ((priceVal - custodyAveragePrice) / custodyAveragePrice) * 100
      : null

  const selectedOrderAccount = accounts.find((a) => a.id === selectedAccountId)
  const projectedAccountBalance = selectedOrderAccount
    ? orderType === 'buy'
      ? parseFloat(selectedOrderAccount.balance || '0') - totalOrderValue
      : parseFloat(selectedOrderAccount.balance || '0') + totalOrderValue
    : null

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Carteira de Investimentos</h1>
          <p className="text-xs sm:text-sm text-slate-400">
            Acompanhamento consolidado de Ações B3, FIIs, Renda Fixa, Criptoativos e Cotações ao vivo.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => handleOpenOrderModal('buy')}
            className="px-4 py-2.5 bg-gradient-to-r from-brand-500 to-emerald-600 hover:from-brand-600 hover:to-emerald-700 active:scale-95 text-white text-xs font-bold rounded-xl shadow-lg shadow-brand-500/20 flex items-center gap-2 transition-all"
          >
            <Plus className="w-4 h-4" />
            <span>Novo Aporte (Compra)</span>
          </button>
          <button
            onClick={() => handleOpenOrderModal('sell')}
            className="px-4 py-2.5 bg-dark-900 hover:bg-red-500/20 border border-red-500/30 text-red-300 hover:text-white active:scale-95 text-xs font-bold rounded-xl flex items-center gap-2 transition-all"
          >
            <ArrowDownRight className="w-4 h-4 text-red-400" />
            <span>Registrar Venda</span>
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        {/* Total Equity */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg relative overflow-hidden">
          <div className="absolute top-0 right-0 w-24 h-24 bg-brand-500/5 rounded-full blur-2xl pointer-events-none"></div>
          <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Patrimônio Total Atual</span>
          <div className="text-2xl font-black text-white mt-2">{displayVal(summary?.total_equity_brl || '0')}</div>
          <span className="text-xs text-slate-500 mt-1 block">Valor de mercado atualizado</span>
        </div>

        {/* Total Cost */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg">
          <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Total Aportado (Custo)</span>
          <div className="text-2xl font-black text-slate-200 mt-2">{displayVal(summary?.total_cost_brl || '0')}</div>
          <span className="text-xs text-slate-500 mt-1 block">Base de custo com preço médio</span>
        </div>

        {/* Unrealized PnL */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg">
          <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Lucro / Prejuízo (R$)</span>
          <div className={`text-2xl font-black mt-2 ${parseFloat(summary?.total_pnl_brl || '0') >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
            {displayVal(summary?.total_pnl_brl || '0')}
          </div>
          <span className="text-xs text-slate-500 mt-1 block">Variação patrimonial total</span>
        </div>

        {/* Total Profit % */}
        <div className="bg-dark-900 border border-slate-800 rounded-2xl p-5 shadow-lg">
          <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Rentabilidade Geral</span>
          <div className={`text-2xl font-black mt-2 ${(summary?.total_profit_percentage || 0) >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
            {hideValues ? '•••' : formatPercentage(summary?.total_profit_percentage || 0)}
          </div>
          <span className="text-xs text-slate-500 mt-1 block">Retorno sobre o capital investido</span>
        </div>
      </div>

      {/* Tabs & Main Tables Container */}
      <div className="bg-dark-900 border border-slate-800 rounded-2xl p-6 shadow-lg space-y-4">
        {/* Navigation Tabs */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-4">
          <div className="flex items-center gap-2">
            <button
              onClick={() => setActiveTab('positions')}
              className={`px-4 py-2 rounded-xl text-xs font-bold flex items-center gap-2 transition-all ${
                activeTab === 'positions'
                  ? 'bg-brand-500/15 text-brand-400 border border-brand-500/30 shadow-sm shadow-brand-500/10'
                  : 'text-slate-400 hover:text-white hover:bg-slate-800/40'
              }`}
            >
              <Layers className="w-4 h-4" />
              <span>Posições em Custódia ({summary?.positions?.length || 0})</span>
            </button>
            <button
              onClick={() => setActiveTab('orders')}
              className={`px-4 py-2 rounded-xl text-xs font-bold flex items-center gap-2 transition-all ${
                activeTab === 'orders'
                  ? 'bg-brand-500/15 text-brand-400 border border-brand-500/30 shadow-sm shadow-brand-500/10'
                  : 'text-slate-400 hover:text-white hover:bg-slate-800/40'
              }`}
            >
              <History className="w-4 h-4" />
              <span>Histórico de Compras e Vendas ({orders.length})</span>
            </button>
          </div>
          <span className="text-xs text-slate-500">
            {activeTab === 'positions'
              ? 'Posições calculadas com preço médio ponderado'
              : 'Você pode editar ou excluir compras e vendas registradas'}
          </span>
        </div>

        {/* TAB 1: Positions Table */}
        {activeTab === 'positions' && (
          <div>
            {!summary?.positions || summary.positions.length === 0 ? (
              <div className="py-14 text-center text-xs text-slate-500 space-y-2">
                <p>Nenhum ativo registrado nesta carteira.</p>
                <p className="text-slate-600">Clique em <strong>"Novo Aporte (Compra)"</strong> acima para registrar compras de Ações B3 (ex: PETR4, VALE3), FIIs (ex: MXRF11), ou Cripto (ex: BTC).</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
                      <th className="pb-3">Ticker / Ativo</th>
                      <th className="pb-3">Classe</th>
                      <th className="pb-3 text-right">Qtd</th>
                      <th className="pb-3 text-right">Preço Médio</th>
                      <th className="pb-3 text-right">Cotação Atual</th>
                      <th className="pb-3 text-right">Total Investido</th>
                      <th className="pb-3 text-right">Valor Atual</th>
                      <th className="pb-3 text-right">Lucro/Prejuízo</th>
                      <th className="pb-3 text-right">Alocação</th>
                      <th className="pb-3 text-center">Ações</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/50">
                    {summary.positions.map((pos) => {
                      const isProfit = pos.profit_percentage >= 0
                      return (
                        <tr key={pos.position.id} className="hover:bg-slate-800/30 transition-colors">
                          <td className="py-3.5">
                            <div className="font-bold text-white text-sm">{pos.position.asset?.ticker}</div>
                            <span className="text-[11px] text-slate-500 block truncate max-w-[150px]">
                              {pos.position.asset?.name}
                            </span>
                          </td>
                          <td className="py-3.5">
                            <span className="px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 font-mono text-[11px] uppercase">
                              {pos.position.asset?.type}
                            </span>
                          </td>
                          <td className="py-3.5 text-right font-mono text-slate-200">
                            {parseFloat(pos.position.quantity).toLocaleString('pt-BR', { maximumFractionDigits: 4 })}
                          </td>
                          <td className="py-3.5 text-right font-mono text-slate-300">
                            {displayVal(pos.position.average_price, pos.position.asset?.currency)}
                          </td>
                          <td className="py-3.5 text-right font-mono font-semibold text-white">
                            {displayVal(pos.current_price, pos.position.asset?.currency)}
                          </td>
                          <td className="py-3.5 text-right font-mono text-slate-400">{displayVal(pos.total_cost_brl)}</td>
                          <td className="py-3.5 text-right font-mono font-bold text-slate-100">{displayVal(pos.current_value_brl)}</td>
                          <td className="py-3.5 text-right">
                            <div className={`font-bold font-mono ${isProfit ? 'text-emerald-400' : 'text-red-400'}`}>
                              {isProfit ? '+' : ''}{displayVal(pos.unrealized_pnl_brl)}
                            </div>
                            <span className={`text-[10px] block ${isProfit ? 'text-emerald-400' : 'text-red-400'}`}>
                              {hideValues ? '•••' : formatPercentage(pos.profit_percentage)}
                            </span>
                          </td>
                          <td className="py-3.5 text-right font-mono text-slate-300">
                            {pos.allocation_percentage.toFixed(1)}%
                          </td>
                          <td className="py-3.5 text-center">
                            <div className="flex items-center justify-center gap-1.5">
                              <button
                                onClick={() => handleOpenSellModalForPosition(pos)}
                                title="Vender este ativo da carteira"
                                className="px-2.5 py-1 rounded-lg bg-red-500/15 hover:bg-red-500 text-red-300 hover:text-white border border-red-500/30 text-[11px] font-semibold flex items-center gap-1 transition-all"
                              >
                                <ArrowDownRight className="w-3 h-3" />
                                <span>Vender</span>
                              </button>
                              <button
                                onClick={() => handleUpdateQuote(pos.position.asset_id)}
                                title="Atualizar Cotação em Tempo Real via Brapi"
                                className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
                              >
                                <RefreshCw className={`w-3.5 h-3.5 ${syncingQuote === pos.position.asset_id ? 'animate-spin text-brand-500' : ''}`} />
                              </button>
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* TAB 2: Orders History (Buys & Sells) with Full Edit & Delete */}
        {activeTab === 'orders' && (
          <div>
            {orders.length === 0 ? (
              <div className="py-14 text-center text-xs text-slate-500 space-y-2">
                <p>Nenhuma ordem de compra ou venda registrada nesta carteira.</p>
                <p className="text-slate-600">Suas operações de aporte e desinvestimento aparecerão aqui com opção de edição e exclusão.</p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-slate-800 text-slate-400 font-semibold uppercase tracking-wider">
                      <th className="pb-3">Data</th>
                      <th className="pb-3">Ativo</th>
                      <th className="pb-3">Operação</th>
                      <th className="pb-3 text-right">Quantidade</th>
                      <th className="pb-3 text-right">Preço Unitário</th>
                      <th className="pb-3 text-right">Taxas</th>
                      <th className="pb-3 text-right">Total Operação</th>
                      <th className="pb-3 text-right">Lucro/Prejuízo</th>
                      <th className="pb-3">Conta</th>
                      <th className="pb-3">Notas</th>
                      <th className="pb-3 text-center">Ações</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/50">
                    {orders.map((ord) => {
                      const isBuy = ord.type === 'buy'
                      const pnlNum = parseFloat(ord.realized_pnl || '0')
                      return (
                        <tr key={ord.id} className="hover:bg-slate-800/30 transition-colors">
                          <td className="py-3.5 text-slate-400 font-mono">
                            {new Date(ord.date).toLocaleDateString('pt-BR')}
                          </td>
                          <td className="py-3.5">
                            <span className="font-bold text-white font-mono text-sm block">
                              {ord.asset?.ticker || 'Ativo'}
                            </span>
                            <span className="text-[11px] text-slate-500 truncate block max-w-[140px]">
                              {ord.asset?.name}
                            </span>
                          </td>
                          <td className="py-3.5">
                            {isBuy ? (
                              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-emerald-500/15 border border-emerald-500/30 text-emerald-400 text-[11px] font-bold">
                                <ArrowUpRight className="w-3 h-3" />
                                Compra (Aporte)
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-red-500/15 border border-red-500/30 text-red-400 text-[11px] font-bold">
                                <ArrowDownRight className="w-3 h-3" />
                                Venda
                              </span>
                            )}
                          </td>
                          <td className="py-3.5 text-right font-mono font-semibold text-slate-200">
                            {parseFloat(ord.quantity).toLocaleString('pt-BR', { maximumFractionDigits: 4 })}
                          </td>
                          <td className="py-3.5 text-right font-mono text-slate-300">
                            {displayVal(ord.unit_price, ord.asset?.currency)}
                          </td>
                          <td className="py-3.5 text-right font-mono text-slate-400">
                            {displayVal(ord.fees || '0', ord.asset?.currency)}
                          </td>
                          <td className="py-3.5 text-right font-mono font-bold text-white">
                            {displayVal(ord.total_amount, ord.asset?.currency)}
                          </td>
                          <td className="py-3.5 text-right font-mono">
                            {!isBuy ? (
                              <span className={`font-bold ${pnlNum >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                                {pnlNum >= 0 ? '+' : ''}{displayVal(ord.realized_pnl, ord.asset?.currency)}
                              </span>
                            ) : (
                              <span className="text-slate-600">-</span>
                            )}
                          </td>
                          <td className="py-3.5 text-xs text-slate-300">
                            {(() => {
                              const acc = accounts.find((a) => a.id === ord.account_id)
                              return acc ? (
                                <span className="inline-flex items-center px-2 py-0.5 rounded bg-slate-800/80 border border-slate-700/60 text-slate-300 text-[11px] font-medium">
                                  {acc.name}
                                </span>
                              ) : (
                                <span className="text-slate-600">-</span>
                              )
                            })()}
                          </td>
                          <td className="py-3.5 text-slate-400 text-[11px] max-w-[140px] truncate">
                            {ord.notes || <span className="text-slate-600 italic">Sem notas</span>}
                          </td>
                          <td className="py-3.5 text-center">
                            <div className="flex items-center justify-center gap-1.5">
                              <button
                                onClick={() => handleOpenEditOrder(ord)}
                                title="Editar esta operação de compra/venda"
                                className="p-1.5 rounded-lg bg-slate-800 hover:bg-brand-500/20 text-slate-300 hover:text-brand-400 border border-slate-700 transition-colors"
                              >
                                <Pencil className="w-3.5 h-3.5" />
                              </button>
                              <button
                                onClick={() => handleDeleteOrder(ord)}
                                title="Excluir esta operação e recalcular carteira"
                                className="p-1.5 rounded-lg bg-slate-800 hover:bg-red-500/20 text-slate-300 hover:text-red-400 border border-slate-700 transition-colors"
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                              </button>
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Order Modal (Buy / Sell) */}
      {isOrderModalOpen && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-md z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-3xl p-6 sm:p-7 max-w-md w-full shadow-2xl space-y-5 relative">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                {orderType === 'buy' ? (
                  <>
                    <Sparkles className="w-4 h-4 text-emerald-400" />
                    <span>Registrar Compra (Aporte)</span>
                  </>
                ) : (
                  <>
                    <TrendingDown className="w-4 h-4 text-red-400" />
                    <span>Registrar Venda de Ativo</span>
                  </>
                )}
              </h3>
            </div>

            {/* Type selector (Buy vs Sell tabs) */}
            <div className="grid grid-cols-2 gap-2 p-1 bg-dark-950 rounded-xl border border-slate-800">
              <button
                type="button"
                onClick={() => {
                  setOrderType('buy')
                  setFormError('')
                }}
                className={`py-2 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 ${
                  orderType === 'buy' ? 'bg-emerald-500 text-white shadow-md shadow-emerald-500/20' : 'text-slate-400 hover:text-white'
                }`}
              >
                <ArrowUpRight className="w-3.5 h-3.5" />
                <span>Compra (Aporte)</span>
              </button>
              <button
                type="button"
                onClick={() => {
                  setOrderType('sell')
                  setFormError('')
                }}
                className={`py-2 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 ${
                  orderType === 'sell' ? 'bg-red-500 text-white shadow-md shadow-red-500/20' : 'text-slate-400 hover:text-white'
                }`}
              >
                <ArrowDownRight className="w-3.5 h-3.5" />
                <span>Venda</span>
              </button>
            </div>

            {/* Error Alert */}
            {formError && (
              <div className="flex items-start gap-2.5 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs">
                <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                <span>{formError}</span>
              </div>
            )}

            <form onSubmit={handleExecuteOrder} className="space-y-4">
              {/* Typeable Asset (Ticker) Input with Autocomplete Dropdown */}
              <div className="space-y-1.5 relative" ref={dropdownRef}>
                <div className="flex items-center justify-between">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Ativo (Ticker)
                  </label>
                  {isSearching && (
                    <span className="text-[10px] text-brand-400 font-semibold flex items-center gap-1">
                      <RefreshCw className="w-3 h-3 animate-spin" />
                      Buscando na B3 (Brapi)...
                    </span>
                  )}
                </div>
                <div className="relative">
                  <Search className="w-4 h-4 absolute left-3.5 top-3.5 text-slate-500" />
                  <input
                    type="text"
                    required
                    value={tickerInput}
                    onChange={(e) => {
                      const val = e.target.value.toUpperCase()
                      setTickerInput(val)
                      setIsDropdownOpen(true)

                      // Check exact match
                      const matched = searchResults.find((a) => a.ticker.toUpperCase() === val)
                      if (matched) {
                        setSelectedAsset(matched)
                        setOrderPrice(matched.current_price || '0')
                        setLiveQuoteBadge(`Brapi: ${formatCurrency(matched.current_price || '0', matched.currency)}`)
                        const pos = summary?.positions.find((p) => p.position.asset_id === matched.id)
                        if (pos) {
                          setAvailableCustodyQty(parseFloat(pos.position.quantity))
                          setCustodyAveragePrice(parseFloat(pos.position.average_price))
                        }
                      } else {
                        setSelectedAsset(null)
                      }
                    }}
                    onFocus={() => setIsDropdownOpen(true)}
                    placeholder="Digite o código (ex: PETR4, VALE3, MXRF11, BTC...)"
                    className="w-full pl-10 pr-4 py-3 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white placeholder-slate-600 uppercase focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>

                {/* Suggestions Dropdown */}
                {isDropdownOpen && (
                  <div className="absolute top-full left-0 right-0 mt-1.5 max-h-56 overflow-y-auto bg-dark-950 border border-slate-800 rounded-xl shadow-2xl z-50 divide-y divide-slate-900">
                    {searchResults.length > 0 ? (
                      searchResults.map((asset) => (
                        <button
                          key={asset.id || asset.ticker}
                          type="button"
                          onClick={() => handleSelectAsset(asset)}
                          className="w-full px-3.5 py-2.5 flex items-center justify-between text-left hover:bg-slate-900/80 transition-colors group"
                        >
                          <div>
                            <div className="flex items-center gap-2">
                              <span className="font-mono font-bold text-white text-xs group-hover:text-brand-400">
                                {asset.ticker}
                              </span>
                              <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 uppercase">
                                {asset.type}
                              </span>
                            </div>
                            <span className="text-[11px] text-slate-400 truncate block max-w-[220px]">
                              {asset.name}
                            </span>
                          </div>
                          <div className="text-right">
                            <span className="font-mono text-xs font-semibold text-slate-200 block">
                              {formatCurrency(asset.current_price || '0', asset.currency)}
                            </span>
                            {selectedAsset?.id === asset.id && (
                              <Check className="w-3.5 h-3.5 text-brand-400 ml-auto mt-0.5" />
                            )}
                          </div>
                        </button>
                      ))
                    ) : (
                      <div className="p-3 text-center text-xs text-slate-500">
                        {isSearching ? 'Buscando ativo na B3...' : 'Nenhum ativo encontrado.'}
                      </div>
                    )}

                    {/* Quick option to register a new custom typed ticker */}
                    {tickerInput.trim().length > 0 &&
                      !searchResults.some((a) => a.ticker.toUpperCase() === tickerInput.trim().toUpperCase()) && (
                        <div className="p-3 bg-brand-500/5 border-t border-brand-500/20 text-xs text-slate-300">
                          <span className="text-brand-400 font-semibold block mb-0.5">
                            + Usar novo ticker: {tickerInput.trim().toUpperCase()}
                          </span>
                          <span className="text-[11px] text-slate-400 block">
                            O sistema buscará a cotação de mercado automaticamente ao salvar.
                          </span>
                        </div>
                      )}
                  </div>
                )}
              </div>

              {/* Custody Info Badge for Sell Mode */}
              {orderType === 'sell' && availableCustodyQty !== null && (
                <div className="p-3 bg-dark-950 border border-slate-800 rounded-xl flex items-center justify-between text-xs">
                  <div>
                    <span className="text-slate-400 block text-[11px]">Em Custódia:</span>
                    <span className="font-mono font-bold text-white">{availableCustodyQty} cotas</span>
                    {custodyAveragePrice && (
                      <span className="text-[11px] text-slate-400 block"> (PM: {formatCurrency(custodyAveragePrice, 'BRL')})</span>
                    )}
                  </div>
                  <div className="flex gap-1.5">
                    <button
                      type="button"
                      onClick={() => setOrderQty(String(availableCustodyQty * 0.5))}
                      className="px-2 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded text-[10px] font-semibold"
                    >
                      50%
                    </button>
                    <button
                      type="button"
                      onClick={() => setOrderQty(String(availableCustodyQty))}
                      className="px-2 py-1 bg-brand-500/20 hover:bg-brand-500 text-brand-300 hover:text-white rounded text-[10px] font-bold transition-colors"
                    >
                      100% (Tudo)
                    </button>
                  </div>
                </div>
              )}

              {/* Quantity & Unit Price */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Quantidade
                  </label>
                  <input
                    type="number"
                    step="any"
                    required
                    placeholder="100"
                    value={orderQty}
                    onChange={(e) => setOrderQty(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                      Preço Unitário (R$)
                    </label>
                    {liveQuoteBadge && (
                      <span className="text-[10px] text-emerald-400 font-semibold flex items-center gap-1">
                        <Sparkles className="w-3 h-3 text-emerald-400" />
                        {liveQuoteBadge}
                      </span>
                    )}
                  </div>
                  <div className="relative">
                    <input
                      type="number"
                      step="0.0001"
                      required
                      placeholder="0,00"
                      value={orderPrice}
                      onChange={(e) => setOrderPrice(e.target.value)}
                      className="w-full pl-3.5 pr-8 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                    />
                    <button
                      type="button"
                      title="Buscar cotação ao vivo no Brapi"
                      onClick={() => fetchLiveQuoteFromBrapi(tickerInput)}
                      className="absolute right-2.5 top-2.5 text-slate-500 hover:text-brand-400 transition-colors"
                    >
                      <RefreshCw className={`w-4 h-4 ${isFetchingLiveQuote ? 'animate-spin text-brand-500' : ''}`} />
                    </button>
                  </div>
                </div>
              </div>

              {/* Estimated Results Card */}
              {qtyVal > 0 && priceVal > 0 && (
                <div className="p-3 rounded-xl bg-dark-950/80 border border-slate-800/80 text-xs space-y-1.5">
                  <div className="flex justify-between text-slate-400">
                    <span>{orderType === 'buy' ? 'Valor Total a Pagar:' : 'Valor Total a Receber:'}</span>
                    <strong className="text-white font-mono">{formatCurrency(totalOrderValue, 'BRL')}</strong>
                  </div>
                  {orderType === 'sell' && estimatedSellPnL !== null && (
                    <div className="flex justify-between border-t border-slate-800/60 pt-1.5">
                      <span className="text-slate-400">Lucro/Prejuízo Estimado:</span>
                      <strong className={`font-mono ${estimatedSellPnL >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                        {estimatedSellPnL >= 0 ? '+' : ''}{formatCurrency(estimatedSellPnL, 'BRL')} ({formatPercentage(estimatedSellPnLPercent || 0)})
                      </strong>
                    </div>
                  )}
                </div>
              )}

              {/* Fees & Date */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Taxas B3 / Corretagem
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={orderFees}
                    onChange={(e) => setOrderFees(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Data da Operação
                  </label>
                  <input
                    type="date"
                    required
                    value={orderDate}
                    onChange={(e) => setOrderDate(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm text-white focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>
              </div>

              {/* Account Debit/Credit */}
              {accounts.length === 0 ? (
                <div className="p-3.5 bg-amber-500/10 border border-amber-500/30 rounded-xl text-xs text-amber-300 flex items-start gap-2.5">
                  <AlertCircle className="w-4 h-4 shrink-0 mt-0.5 text-amber-400" />
                  <div>
                    <span className="font-semibold block text-amber-200">Nenhuma conta bancária cadastrada</span>
                    <span>
                      Cadastre uma conta em <strong>Finanças</strong> para que o valor do aporte seja debitado do saldo.
                    </span>
                  </div>
                </div>
              ) : (
                <div className="space-y-1.5">
                  <div className="flex justify-between items-center">
                    <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                      {orderType === 'buy' ? 'Conta Bancária para Débito *' : 'Creditar na Conta Bancária (Opcional)'}
                    </label>
                    {selectedOrderAccount && (
                      <span className="text-[11px] text-slate-400">
                        Saldo atual: <span className="font-mono text-slate-200 font-semibold">{formatCurrency(selectedOrderAccount.balance, selectedOrderAccount.currency)}</span>
                      </span>
                    )}
                  </div>
                  <select
                    value={selectedAccountId}
                    onChange={(e) => setSelectedAccountId(e.target.value)}
                    required={orderType === 'buy'}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-xs text-slate-200 focus:outline-none focus:border-brand-500"
                  >
                    {orderType === 'sell' && (
                      <option value="">Não movimentar saldo em conta</option>
                    )}
                    {accounts.map((acc) => (
                      <option key={acc.id} value={acc.id}>
                        {acc.name} ({acc.institution}) — Saldo: {formatCurrency(acc.balance, acc.currency)}
                      </option>
                    ))}
                  </select>
                  {orderType === 'buy' && selectedOrderAccount && totalOrderValue > 0 && projectedAccountBalance !== null && (
                    <div className="text-[11px] bg-slate-900/70 p-2.5 rounded-xl border border-slate-800 flex justify-between items-center">
                      <span className="text-slate-400">Saldo projetado após o aporte:</span>
                      <span className={`font-mono font-bold ${projectedAccountBalance >= 0 ? 'text-emerald-400' : 'text-amber-400'}`}>
                        {formatCurrency(projectedAccountBalance, selectedOrderAccount.currency)}
                      </span>
                    </div>
                  )}
                </div>
              )}

              {/* Strategy Notes */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                  Notas / Estratégia
                </label>
                <input
                  type="text"
                  placeholder={orderType === 'buy' ? 'Ex: Aporte mensal em ações' : 'Ex: Realização de lucro / rebalanceamento'}
                  value={orderNotes}
                  onChange={(e) => setOrderNotes(e.target.value)}
                  className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm text-white placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                />
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setIsOrderModalOpen(false)}
                  className="flex-1 py-3 bg-slate-800 hover:bg-slate-700 active:scale-95 text-slate-300 text-xs font-bold rounded-xl transition-all"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={isSubmitting || (orderType === 'buy' && accounts.length === 0)}
                  className={`flex-1 py-3 text-white text-xs font-bold rounded-xl active:scale-95 transition-all shadow-lg disabled:opacity-50 ${
                    orderType === 'buy'
                      ? 'bg-gradient-to-r from-brand-500 to-emerald-600 hover:from-brand-600 hover:to-emerald-700 shadow-brand-500/25'
                      : 'bg-gradient-to-r from-red-500 to-rose-600 hover:from-red-600 hover:to-rose-700 shadow-red-500/25'
                  }`}
                >
                  {isSubmitting ? 'Registrando...' : orderType === 'buy' ? 'Confirmar Compra' : 'Confirmar Venda'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Order Modal */}
      {isEditOrderModalOpen && editingOrder && (
        <div className="fixed inset-0 bg-black/75 backdrop-blur-md z-50 flex items-center justify-center p-4">
          <div className="bg-dark-900 border border-slate-800 rounded-3xl p-6 sm:p-7 max-w-md w-full shadow-2xl space-y-5 relative">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <Pencil className="w-4 h-4 text-brand-400" />
                <span>
                  Editar Operação: {editingOrder.type === 'buy' ? 'Compra' : 'Venda'} de {editingOrder.asset?.ticker || 'Ativo'}
                </span>
              </h3>
            </div>

            {/* Badge indicating order type */}
            <div className="p-3 bg-dark-950 border border-slate-800 rounded-xl flex items-center justify-between">
              <div>
                <span className="text-xs text-slate-400 block">Ativo:</span>
                <span className="font-mono font-bold text-white text-sm">{editingOrder.asset?.ticker}</span>
                <span className="text-xs text-slate-500 block truncate max-w-[200px]">{editingOrder.asset?.name}</span>
              </div>
              <div>
                {editingOrder.type === 'buy' ? (
                  <span className="px-2.5 py-1 rounded-lg bg-emerald-500/15 border border-emerald-500/30 text-emerald-400 text-xs font-bold flex items-center gap-1">
                    <ArrowUpRight className="w-3.5 h-3.5" /> Compra
                  </span>
                ) : (
                  <span className="px-2.5 py-1 rounded-lg bg-red-500/15 border border-red-500/30 text-red-400 text-xs font-bold flex items-center gap-1">
                    <ArrowDownRight className="w-3.5 h-3.5" /> Venda
                  </span>
                )}
              </div>
            </div>

            {/* Error Alert */}
            {editError && (
              <div className="flex items-start gap-2.5 p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs">
                <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                <span>{editError}</span>
              </div>
            )}

            <form onSubmit={handleSaveEditOrder} className="space-y-4">
              {/* Quantity & Unit Price */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Quantidade
                  </label>
                  <input
                    type="number"
                    step="any"
                    required
                    value={editQty}
                    onChange={(e) => setEditQty(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white focus:outline-none focus:border-brand-500"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Preço Unitário ({editingOrder.asset?.currency || 'BRL'})
                  </label>
                  <input
                    type="number"
                    step="0.0001"
                    required
                    value={editPrice}
                    onChange={(e) => setEditPrice(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white focus:outline-none focus:border-brand-500"
                  />
                </div>
              </div>

              {/* Total calculation preview */}
              {parseFloat(editQty) > 0 && parseFloat(editPrice) > 0 && (
                <div className="p-3 rounded-xl bg-dark-950/80 border border-slate-800/80 text-xs flex justify-between">
                  <span className="text-slate-400">Novo Total da Operação:</span>
                  <strong className="text-white font-mono">
                    {formatCurrency(
                      parseFloat(editQty) * parseFloat(editPrice) +
                        (editingOrder.type === 'buy' ? parseFloat(editFees || '0') : -parseFloat(editFees || '0')),
                      editingOrder.asset?.currency
                    )}
                  </strong>
                </div>
              )}

              {/* Fees & Date */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Taxas / Corretagem
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={editFees}
                    onChange={(e) => setEditFees(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm font-mono text-white focus:outline-none focus:border-brand-500"
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Data da Operação
                  </label>
                  <input
                    type="date"
                    required
                    value={editDate}
                    onChange={(e) => setEditDate(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm text-white focus:outline-none focus:border-brand-500"
                  />
                </div>
              </div>

              {/* Optional Account Link */}
              {accounts.length > 0 && (
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Conta Bancária Vinculada (Opcional)
                  </label>
                  <select
                    value={editAccountId}
                    onChange={(e) => setEditAccountId(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-xs text-slate-200 focus:outline-none focus:border-brand-500"
                  >
                    <option value="">Não movimentar saldo em conta</option>
                    {accounts.map((acc) => (
                      <option key={acc.id} value={acc.id}>
                        {acc.name} ({acc.institution}) - Saldo: {formatCurrency(acc.balance, acc.currency)}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* Strategy Notes */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                  Notas / Estratégia
                </label>
                <input
                  type="text"
                  placeholder="Observações sobre a operação..."
                  value={editNotes}
                  onChange={(e) => setEditNotes(e.target.value)}
                  className="w-full px-3.5 py-2.5 bg-dark-950 border border-slate-800 rounded-xl text-sm text-white focus:outline-none focus:border-brand-500"
                />
              </div>

              <div className="p-3 bg-brand-500/5 border border-brand-500/20 rounded-xl text-[11px] text-brand-300">
                ℹ️ Ao salvar, o preço médio e a rentabilidade da carteira serão recalculados automaticamente em ordem cronológica.
              </div>

              {/* Action Buttons */}
              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setIsEditOrderModalOpen(false)}
                  className="flex-1 py-3 bg-slate-800 hover:bg-slate-700 active:scale-95 text-slate-300 text-xs font-bold rounded-xl transition-all"
                >
                  Cancelar
                </button>
                <button
                  type="submit"
                  disabled={isEditSubmitting}
                  className="flex-1 py-3 bg-gradient-to-r from-brand-500 to-emerald-600 hover:from-brand-600 hover:to-emerald-700 text-white text-xs font-bold rounded-xl active:scale-95 transition-all shadow-lg shadow-brand-500/25 disabled:opacity-50"
                >
                  {isEditSubmitting ? 'Salvando...' : 'Salvar Alterações'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
