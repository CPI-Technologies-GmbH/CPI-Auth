import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

// Mock the api module
vi.mock('@/lib/api', () => ({
  api: {
    login: vi.fn(),
    logout: vi.fn(),
    getMe: vi.fn(),
  },
}))

import { useAuthStore } from './auth'
import { api } from '@/lib/api'

describe('useAuthStore', () => {
  let mockLocalStorage: Record<string, string>

  beforeEach(() => {
    // Mock localStorage
    mockLocalStorage = {}
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => mockLocalStorage[key] ?? null,
      setItem: (key: string, value: string) => {
        mockLocalStorage[key] = value
      },
      removeItem: (key: string) => {
        delete mockLocalStorage[key]
      },
    })

    // Reset the store state
    useAuthStore.setState({
      user: null,
      isAuthenticated: false,
      isLoading: true,
    })

    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('initial state', () => {
    it('should have null user', () => {
      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
    })

    it('should have isLoading true initially', () => {
      const state = useAuthStore.getState()
      expect(state.isLoading).toBe(true)
    })
  })

  describe('login', () => {
    it('should call api.login and api.getMe, then set user and isAuthenticated', async () => {
      const mockUser = { id: '1', email: 'admin@test.com', name: 'Admin', role: 'admin' as const, tenant_id: 't1', created_at: '2024-01-01' }
      vi.mocked(api.login).mockResolvedValueOnce({
        access_token: 'at',
        refresh_token: 'rt',
        expires_in: 3600,
        token_type: 'Bearer',
      })
      vi.mocked(api.getMe).mockResolvedValueOnce(mockUser)

      await useAuthStore.getState().login('admin@test.com', 'password')

      expect(api.login).toHaveBeenCalledWith('admin@test.com', 'password')
      expect(api.getMe).toHaveBeenCalled()

      const state = useAuthStore.getState()
      expect(state.user).toEqual(mockUser)
      expect(state.isAuthenticated).toBe(true)
      expect(state.isLoading).toBe(false)
    })

    it('should propagate login errors', async () => {
      vi.mocked(api.login).mockRejectedValueOnce(new Error('Invalid credentials'))

      await expect(
        useAuthStore.getState().login('admin@test.com', 'wrong')
      ).rejects.toThrow('Invalid credentials')
    })
  })

  describe('logout', () => {
    it('should call api.logout and reset state', async () => {
      // Set up authenticated state
      useAuthStore.setState({
        user: { id: '1', email: 'admin@test.com', name: 'Admin', role: 'admin', tenant_id: 't1', created_at: '2024-01-01' },
        isAuthenticated: true,
        isLoading: false,
      })

      vi.mocked(api.logout).mockResolvedValueOnce(undefined)

      await useAuthStore.getState().logout()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isAuthenticated).toBe(false)
    })

    it('should reset state even if api.logout throws', async () => {
      useAuthStore.setState({
        user: { id: '1', email: 'admin@test.com', name: 'Admin', role: 'admin', tenant_id: 't1', created_at: '2024-01-01' },
        isAuthenticated: true,
        isLoading: false,
      })

      vi.mocked(api.logout).mockRejectedValueOnce(new Error('Network error'))

      // The try/finally in logout lets the error propagate, so we catch it here
      try {
        await useAuthStore.getState().logout()
      } catch {
        // Expected to throw
      }

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isAuthenticated).toBe(false)
    })
  })

  describe('loadUser', () => {
    it('should load user when access_token exists', async () => {
      mockLocalStorage['access_token'] = 'at123'
      const mockUser = { id: '1', email: 'admin@test.com', name: 'Admin', role: 'admin' as const, tenant_id: 't1', created_at: '2024-01-01' }
      vi.mocked(api.getMe).mockResolvedValueOnce(mockUser)

      await useAuthStore.getState().loadUser()

      const state = useAuthStore.getState()
      expect(state.user).toEqual(mockUser)
      expect(state.isAuthenticated).toBe(true)
      expect(state.isLoading).toBe(false)
    })

    it('should set isLoading false when no access_token', async () => {
      await useAuthStore.getState().loadUser()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isLoading).toBe(false)
    })

    it('should reset auth state when getMe fails', async () => {
      mockLocalStorage['access_token'] = 'expired_token'
      vi.mocked(api.getMe).mockRejectedValueOnce(new Error('Unauthorized'))

      await useAuthStore.getState().loadUser()

      const state = useAuthStore.getState()
      expect(state.user).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.isLoading).toBe(false)
    })
  })

  describe('setUser', () => {
    it('should update the user', () => {
      const mockUser = { id: '1', email: 'admin@test.com', name: 'Admin', role: 'admin' as const, tenant_id: 't1', created_at: '2024-01-01' }
      useAuthStore.getState().setUser(mockUser)

      expect(useAuthStore.getState().user).toEqual(mockUser)
    })

    it('should allow setting user to null', () => {
      useAuthStore.getState().setUser(null)

      expect(useAuthStore.getState().user).toBeNull()
    })
  })
})
