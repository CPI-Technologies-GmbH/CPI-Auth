import { create } from 'zustand'
import type { AdminUser } from '@/types'
import { api } from '@/lib/api'

interface AuthState {
  user: AdminUser | null
  isAuthenticated: boolean
  isLoading: boolean
  activeTenantId: string | null
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  loadUser: () => Promise<void>
  setUser: (user: AdminUser | null) => void
  setActiveTenant: (id: string | null) => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: !!localStorage.getItem('access_token'),
  isLoading: true,
  activeTenantId: localStorage.getItem('active_tenant_id'),

  login: async (email: string, password: string) => {
    await api.login(email, password)
    const user = await api.getMe()
    set({ user, isAuthenticated: true, isLoading: false })
  },

  logout: async () => {
    try {
      await api.logout()
    } finally {
      localStorage.removeItem('active_tenant_id')
      set({ user: null, isAuthenticated: false, isLoading: false, activeTenantId: null })
    }
  },

  loadUser: async () => {
    try {
      if (!localStorage.getItem('access_token')) {
        set({ isLoading: false })
        return
      }
      const user = await api.getMe()
      const stored = localStorage.getItem('active_tenant_id')
      set({
        user,
        isAuthenticated: true,
        isLoading: false,
        activeTenantId: stored || user.tenant_id,
      })
    } catch {
      set({ user: null, isAuthenticated: false, isLoading: false })
    }
  },

  setUser: (user) => set({ user }),

  setActiveTenant: (id) => {
    if (id) {
      localStorage.setItem('active_tenant_id', id)
    } else {
      localStorage.removeItem('active_tenant_id')
    }
    set({ activeTenantId: id })
  },
}))
