import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  TrendingUp,
  Lock,
  Mail,
  User as UserIcon,
  ArrowRight,
  AlertCircle,
  Eye,
  EyeOff,
  Sparkles,
  CheckCircle2
} from 'lucide-react'
import api from '../../services/api'
import { useAuthStore } from '../../stores/auth'
import { AuthResponse } from '../../types'

export const Register: React.FC = () => {
  const [name, setName] = useState('')
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
      const res = await api.post<AuthResponse>('/auth/register', { name, email, password })
      const { user, access_token, refresh_token } = res.data
      setAuth(user, access_token, refresh_token)
      navigate('/')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Erro ao criar conta. Tente novamente.')
    } finally {
      setLoading(false)
    }
  }

  // Password requirements feedback
  const hasMinLength = password.length >= 6
  const hasUppercase = /[A-Z]/.test(password)
  const hasNumber = /[0-9]/.test(password)

  return (
    <div className="min-h-screen bg-dark-950 text-slate-100 flex relative overflow-hidden font-sans select-none">
      {/* Ambient Glowing Background Orbs */}
      <div className="absolute -top-40 -left-40 w-96 h-96 bg-brand-500/20 rounded-full blur-[128px] pointer-events-none"></div>
      <div className="absolute top-1/2 -right-40 w-96 h-96 bg-purple-500/15 rounded-full blur-[140px] pointer-events-none"></div>
      <div className="absolute -bottom-40 left-1/3 w-96 h-96 bg-blue-500/15 rounded-full blur-[130px] pointer-events-none"></div>

      {/* Grid Overlay */}
      <div
        className="absolute inset-0 opacity-[0.03] pointer-events-none"
        style={{
          backgroundImage: `radial-gradient(circle at 1px 1px, white 1px, transparent 0)`,
          backgroundSize: '32px 32px',
        }}
      ></div>

      <div className="w-full max-w-7xl mx-auto flex flex-col lg:flex-row items-center justify-center p-6 lg:p-12 relative z-10 gap-12 lg:gap-16 my-auto">
        {/* Left Side: Brand Value Props */}
        <div className="hidden lg:flex flex-col flex-1 max-w-lg space-y-8">
          <div className="space-y-4">
            <div className="inline-flex items-center gap-2.5 px-3.5 py-1.5 rounded-full bg-brand-500/10 border border-brand-500/20 text-brand-400 text-xs font-semibold tracking-wide">
              <Sparkles className="w-3.5 h-3.5 text-brand-400" />
              <span>Cadastro 100% Gratuito</span>
            </div>

            <h1 className="text-4xl xl:text-5xl font-black text-white tracking-tight leading-[1.15]">
              Comece a investir com{' '}
              <span className="bg-gradient-to-r from-brand-400 via-emerald-300 to-teal-200 bg-clip-text text-transparent">
                clareza total.
              </span>
            </h1>

            <p className="text-slate-400 text-base leading-relaxed">
              Junte-se a investidores inteligentes que utilizam o CapitalHub para consolidar carteiras da B3, criptoativos e orçamentos mensais.
            </p>
          </div>

          {/* Benefits list */}
          <div className="space-y-4 pt-2">
            {[
              'Cálculo automático de Preço Médio móvel ponderado (padrão B3/IRPF)',
              'Gestão de fluxo de caixa e orçamentos mensais categorizados',
              'Acompanhamento de proventos e calendário de dividendos',
              'Controle de orçamento mensal com alertas inteligentes de gastos',
            ].map((benefit, i) => (
              <div key={i} className="flex items-center gap-3 text-sm text-slate-300">
                <div className="w-5 h-5 rounded-full bg-brand-500/20 text-brand-400 flex items-center justify-center shrink-0">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                </div>
                <span>{benefit}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Right Side: Register Card */}
        <div className="w-full max-w-md">
          <div className="bg-dark-900/80 backdrop-blur-2xl border border-slate-800/90 rounded-3xl p-8 sm:p-10 shadow-2xl shadow-black/80 space-y-6 relative">
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-3/4 h-[2px] bg-gradient-to-r from-transparent via-brand-500 to-transparent"></div>

            <div className="text-center space-y-2">
              <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-brand-500 to-emerald-600 flex items-center justify-center mx-auto shadow-xl shadow-brand-500/30">
                <TrendingUp className="w-6 h-6 text-white stroke-[2.5]" />
              </div>
              <h2 className="text-2xl sm:text-3xl font-extrabold text-white tracking-tight">Criar Nova Conta</h2>
              <p className="text-xs sm:text-sm text-slate-400">Preencha seus dados para iniciar instantaneamente</p>
            </div>

            {error && (
              <div className="flex items-start gap-3 p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs animate-shake">
                <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
                <span className="leading-relaxed">{error}</span>
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              {/* Name */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                  Nome Completo
                </label>
                <div className="relative group">
                  <UserIcon className="w-4 h-4 absolute left-3.5 top-3.5 text-slate-500 group-focus-within:text-brand-400 transition-colors" />
                  <input
                    type="text"
                    required
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="Lucas Silva"
                    className="w-full pl-10 pr-4 py-3 bg-dark-950/90 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>
              </div>

              {/* Email */}
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
                    placeholder="lucas@exemplo.com"
                    className="w-full pl-10 pr-4 py-3 bg-dark-950/90 border border-slate-800 rounded-xl text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-brand-500 focus:ring-1 focus:ring-brand-500/50 transition-all"
                  />
                </div>
              </div>

              {/* Password */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider block">
                  Senha
                </label>
                <div className="relative group">
                  <Lock className="w-4 h-4 absolute left-3.5 top-3.5 text-slate-500 group-focus-within:text-brand-400 transition-colors" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    required
                    minLength={6}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="Mínimo 6 caracteres"
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

                {/* Password validation indicators */}
                {password.length > 0 && (
                  <div className="flex gap-2 text-[10px] pt-1 text-slate-400">
                    <span className={hasMinLength ? 'text-emerald-400 font-semibold' : 'text-slate-500'}>
                      ✓ 6+ dígitos
                    </span>
                    <span className={hasUppercase ? 'text-emerald-400 font-semibold' : 'text-slate-500'}>
                      ✓ Letra maiúscula
                    </span>
                    <span className={hasNumber ? 'text-emerald-400 font-semibold' : 'text-slate-500'}>
                      ✓ Número
                    </span>
                  </div>
                )}
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3.5 px-4 bg-gradient-to-r from-brand-500 to-emerald-600 hover:from-brand-600 hover:to-emerald-700 active:scale-[0.98] text-white font-bold rounded-xl text-sm shadow-xl shadow-brand-500/25 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed mt-2"
              >
                <span>{loading ? 'Criando conta...' : 'Cadastrar Gratuitamente'}</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </form>

            <div className="text-center pt-4 border-t border-slate-800/80">
              <span className="text-xs text-slate-400">Já tem uma conta? </span>
              <Link
                to="/login"
                className="text-xs font-bold text-brand-400 hover:text-brand-300 hover:underline transition-all ml-1"
              >
                Fazer login &rarr;
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
