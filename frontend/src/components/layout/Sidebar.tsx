import React from 'react'
import { NavLink } from 'react-router-dom'
import { LayoutDashboard, Wallet, CreditCard, TrendingUp, LogOut } from 'lucide-react'
import { useAuthStore } from '../../stores/auth'

export const Sidebar: React.FC = () => {
  const logout = useAuthStore((state) => state.logout)

  const navItems = [
    { to: '/', label: 'Dashboard', icon: LayoutDashboard },
    { to: '/finances', label: 'Finanças & Orçamento', icon: Wallet },
    { to: '/credit-cards', label: 'Cartões & Faturas', icon: CreditCard },
    { to: '/investments', label: 'Investimentos', icon: TrendingUp },
  ]

  return (
    <aside className="w-64 bg-dark-900 border-r border-slate-800 flex flex-col h-screen select-none">
      {/* Brand Header */}
      <div className="h-16 flex items-center gap-3 px-6 border-b border-slate-800">
        <div className="w-9 h-9 rounded-xl bg-brand-500 flex items-center justify-center shadow-lg shadow-brand-500/20">
          <TrendingUp className="w-5 h-5 text-white stroke-[2.5]" />
        </div>
        <div>
          <h1 className="font-bold text-white tracking-wide text-base">CapitalHub</h1>
          <span className="text-xs text-brand-500 font-medium tracking-wider uppercase">Finanças & Portfólio</span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-6 space-y-1.5 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all ${
                  isActive
                    ? 'bg-brand-500/10 text-brand-500 border border-brand-500/20'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
                }`
              }
            >
              <Icon className="w-4 h-4" />
              <span>{item.label}</span>
            </NavLink>
          )
        })}
      </nav>

      {/* User Footer */}
      <div className="p-4 border-t border-slate-800 flex items-center justify-between">
        <button
          onClick={logout}
          className="flex items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium text-slate-400 hover:text-red-400 hover:bg-red-500/10 transition-colors w-full"
        >
          <LogOut className="w-4 h-4" />
          <span>Sair da Conta</span>
        </button>
      </div>
    </aside>
  )
}
