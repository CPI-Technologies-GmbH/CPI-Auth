import { create } from 'zustand'
import type { AdminUser, Tenant } from '@/types'
import { api } from '@/lib/api'

interface AuthState {
  user: AdminUser | null
  isAuthenticated: boolean
  isLoading: boolean
  activeTenantId: string | null
  /** Cached list of tenants the current user can switch between (super-admin only). */
  availableTenants: Tenant[]
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  loadUser: () => Promise<void>
  setUser: (user: AdminUser | null) => void
  setActiveTenant: (id: string | null) => void
  /**
   * Sync the active tenant from a URL slug. Looks the slug up in the
   * tenant cache; if missing, fetches the tenant list. Used by the
   * <TenantSync> wrapper around /t/:slug/ routes so the URL is the
   * source of truth instead of localStorage.
   */
  setActiveTenantBySlug: (slug: string) => Promise<void>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: !!localStorage.getItem('access_token'),
  isLoading: true,
  activeTenantId: localStorage.getItem('active_tenant_id'),
  availableTenants: [],

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
      set({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        activeTenantId: null,
        availableTenants: [],
      })
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

  setActiveTenantBySlug: async (slug: string) => {
    let tenants = get().availableTenants
    if (tenants.length === 0) {
      try {
        tenants = await api.getTenants()
        set({ availableTenants: tenants })
      } catch {
        return
      }
    }
    const match = tenants.find((t) => t.slug === slug)
    if (!match) return
    if (get().activeTenantId !== match.id) {
      localStorage.setItem('active_tenant_id', match.id)
      set({ activeTenantId: match.id })
    }
  },
}))
