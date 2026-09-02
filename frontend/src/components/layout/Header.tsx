import React from 'react'
import { Eye, EyeOff, User as UserIcon } from 'lucide-react'
import { useAuthStore } from '../../stores/auth'
import { usePrivacyStore } from '../../stores/privacy'

export const Header: React.FC = () => {
  const user = useAuthStore((state) => state.user)
  const { hideValues, toggleHideValues } = usePrivacyStore()

  return (
    <header className="h-16 bg-dark-900/80 backdrop-blur-md border-b border-slate-800 px-8 flex items-center justify-between sticky top-0 z-20">
      <div className="flex items-center gap-2">
        <span className="text-xs text-slate-500 font-medium uppercase tracking-wider">Ambiente Ativo:</span>
        <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
          Produção
        </span>
      </div>

      <div className="flex items-center gap-4">
        {/* Privacy Toggle Button */}
        <button
          onClick={toggleHideValues}
          title={hideValues ? 'Mostrar Valores' : 'Ocultar Valores (Modo Privacidade)'}
          className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-800/60 hover:bg-slate-700/80 text-slate-300 text-xs font-medium transition-colors"
        >
          {hideValues ? <EyeOff className="w-3.5 h-3.5 text-amber-400" /> : <Eye className="w-3.5 h-3.5 text-slate-400" />}
          <span>{hideValues ? 'Valores Ocultos' : 'Ocultar'}</span>
        </button>

        {/* User Badge */}
        <div className="flex items-center gap-3 pl-2 border-l border-slate-800">
          <div className="w-8 h-8 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-300">
            <UserIcon className="w-4 h-4" />
          </div>
          <div className="flex flex-col text-left">
            <span className="text-sm font-semibold text-slate-200 leading-tight">{user?.name || 'Investidor'}</span>
            <span className="text-xs text-slate-400">{user?.email}</span>
          </div>
        </div>
      </div>
    </header>
  )
}
