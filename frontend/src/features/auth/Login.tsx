import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  TrendingUp,
  Lock,
  Mail,
  ArrowRight,
  AlertCircle,
  Eye,
  EyeOff,
  ShieldCheck,
  Zap,
  PieChart,
  Sparkles
} from 'lucide-react'
import api from '../../services/api'
import { useAuthStore } from '../../stores/auth'
import { AuthResponse } from '../../types'

export const Login: React.FC = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const res = await api.post<AuthResponse>('/auth/login', { email, password })
      const { user, access_token, refresh_token } = res.data
      setAuth(user, access_token, refresh_token)
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.error || 'E-mail ou senha incorretos. Verifique suas credenciais.')
    } finally {
      setLoading(false)
    }
  }

  // Pre-fill demo credentials for quick access
  const handleQuickDemo = () => {
    setEmail('demo@capitalhub.com')
    setPassword('Demo@123456')
  }

  return (
    <div className="min-h-screen bg-dark-950 text-slate-100 flex relative overflow-hidden font-sans select-none">
      {/* Dynamic Background Glowing Orbs */}
      <div className="absolute -top-40 -left-40 w-96 h-96 bg-brand-500/20 rounded-full blur-[128px] pointer-events-none"></div>
      <div className="absolute top-1/2 -right-40 w-96 h-96 bg-purple-500/15 rounded-full blur-[140px] pointer-events-none"></div>
      <div className="absolute -bottom-40 left-1/3 w-96 h-96 bg-blue-500/15 rounded-full blur-[130px] pointer-events-none"></div>

      {/* Grid Pattern Overlay */}
      <div
        className="absolute inset-0 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage: `radial-gradient(circle at 1px 1px, white 1px, transparent 0)`,
          backgroundSize: '32px 32px',
        }}
      ></div>

      <div className="w-full max-w-7xl mx-auto flex flex-col lg:flex-row items-center justify-center p-6 lg:p-12 relative z-10 gap-12 lg:gap-16 my-auto">
        {/* Left Side: Brand Showcase & Value Propositions */}
        <div className="hidden lg:flex flex-col flex-1 max-w-lg space-y-8">
          {/* Logo & Headline */}
          <div className="space-y-4">
            <div className="inline-flex items-center gap-2.5 px-3.5 py-1.5 rounded-full bg-brand-500/10 border border-brand-500/20 text-brand-400 text-xs font-semibold tracking-wide">
              <Sparkles className="w-3.5 h-3.5 text-brand-400" />
              <span>Plataforma Inteligente de Investimentos</span>
            </div>

            <h1 className="text-4xl xl:text-5xl font-black text-white tracking-tight leading-[1.15]">
              Domine seu patrimônio com{' '}
              <span className="bg-gradient-to-r from-brand-400 via-emerald-300 to-teal-200 bg-clip-text text-transparent">
                precisão cirúrgica.
              </span>
            </h1>

            <p className="text-slate-400 text-base leading-relaxed">
              Consolidação de carteira multimoeda, cálculo automático de preço médio ponderado na B3 e fluxo de caixa em um ecossistema unificado.
            </p>
          </div>

          {/* Interactive Feature Cards */}
          <div className="space-y-3.5">
            <div className="p-4 rounded-2xl bg-dark-900/60 backdrop-blur-md border border-slate-800/80 flex items-start gap-4 hover:border-slate-700/80 transition-colors shadow-lg">
              <div className="w-10 h-10 rounded-xl bg-brand-500/15 border border-brand-500/30 flex items-center justify-center text-brand-400 shrink-0">
                <TrendingUp className="w-5 h-5" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-white">Preço Médio & Rendimentos Automáticos</h4>
                <p className="text-xs text-slate-400 mt-0.5">
                  Apuração de lucros, compras, vendas e controle rigoroso de proventos (Dividendos & JCP).
                </p>
              </div>
            </div>

            <div className="p-4 rounded-2xl bg-dark-900/60 backdrop-blur-md border border-slate-800/80 flex items-start gap-4 hover:border-slate-700/80 transition-colors shadow-lg">
              <div className="w-10 h-10 rounded-xl bg-blue-500/15 border border-blue-500/30 flex items-center justify-center text-blue-400 shrink-0">
                <PieChart className="w-5 h-5" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-white">Orçamento & Metas Mensais</h4>
                <p className="text-xs text-slate-400 mt-0.5">
                  Controle de gastos por categoria e alertas inteligentes de consumo a 80% do teto.
                </p>
              </div>
            </div>

            <div className="p-4 rounded-2xl bg-dark-900/60 backdrop-blur-md border border-slate-800/80 flex items-start gap-4 hover:border-slate-700/80 transition-colors shadow-lg">
              <div className="w-10 h-10 rounded-xl bg-purple-500/15 border border-purple-500/30 flex items-center justify-center text-purple-400 shrink-0">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <div>
                <h4 className="text-sm font-bold text-white">Segurança Bancária & Privacidade</h4>
                <p className="text-xs text-slate-400 mt-0.5">
                  Criptografia em repouso, tokens JWT com rotação e Modo Privacidade para ocultar valores na tela.
                </p>
              </div>
            </div>
          </div>

          {/* Quick Metrics Bar */}
          <div className="pt-4 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
            <div className="flex items-center gap-2">
              <Zap className="w-4 h-4 text-amber-400" />
              <span>Cotações em Tempo Real (B3 & Cripto)</span>
            </div>
            <div className="flex items-center gap-1 text-emerald-400 font-semibold">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping mr-1"></span>
              Sistemas Online
            </div>
          </div>
        </div>

        {/* Right Side: High-End Auth Glass Card */}
        <div className="w-full max-w-md">
          <div className="bg-dark-900/80 backdrop-blur-2xl border border-slate-800/90 rounded-3xl p-8 sm:p-10 shadow-2xl shadow-black/80 space-y-7 relative">
            {/* Top Glow Accent */}
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-3/4 h-[2px] bg-gradient-to-r from-transparent via-brand-500 to-transparent"></div>

            {/* Mobile Header (Shown on small screens) */}
            <div className="text-center space-y-2">
              <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-brand-500 to-emerald-600 flex items-center justify-center mx-auto shadow-xl shadow-brand-500/30">
                <TrendingUp className="w-6 h-6 text-white stroke-[2.5]" />
              </div>
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Bem-vindo de volta</h2>
              <p className="text-xs sm:text-sm text-slate-400">Insira suas credenciais para acessar sua carteira</p>
            </div>

            {/* Error Notification Alert */}
            {error && (
              <div className="flex items-start gap-3 p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs animate-shake">
                <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                <span className="leading-relaxed">{error}</span>
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Email Input */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                  E-mail
                </label>
                <div className="relative group">
                  <Mail className="w-4 h-4 absolute left-3.5 top-3.5 text-slate-500 group-focus-within:text-brand-400 transition-colors" />
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="seu.email@exemplo.com"
                    className="w-full pl-10 pr-4 py-3 bg-dark-950/90 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>
              </div>

              {/* Password Input */}
              <div className="space-y-1.5">
                <div className="flex items-center justify-between">
                  <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                    Senha
                  </label>
                  <button
                    type="button"
                    onClick={handleQuickDemo}
                    className="text-[11px] font-semibold text-brand-400 hover:text-brand-300 transition-colors"
                  >
                    Usar conta demo
                  </button>
                </div>
                <div className="relative group">
                  <Lock className="w-4 h-4 absolute left-3.5 top-3.5 text-slate-500 group-focus-within:text-brand-400 transition-colors" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••••••"
                    className="w-full pl-10 pr-11 py-3 bg-dark-950/90 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3.5 top-3.5 text-slate-500 hover:text-slate-300 transition-colors"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3.5 px-4 bg-gradient-to-r from-brand-500 to-emerald-600 hover:from-brand-600 hover:to-emerald-700 active:scale-[0.98] text-white font-bold rounded-xl text-sm shadow-xl shadow-brand-500/25 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed mt-2"
              >
                <span>{loading ? 'Autenticando...' : 'Acessar Painel Financeiro'}</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </form>

            {/* Footer / Switch to Register */}
            <div className="text-center pt-4 border-t border-slate-800/80">
              <span className="text-xs text-slate-400">Novo por aqui? </span>
              <Link
                to="/register"
                className="text-xs font-bold text-brand-400 hover:text-brand-300 hover:underline transition-all ml-1"
              >
                Criar conta gratuita &rarr;
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
