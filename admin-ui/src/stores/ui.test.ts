import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { useUIStore } from './ui'

describe('useUIStore', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    // Reset store to initial state
    useUIStore.setState({
      sidebarOpen: true,
      sidebarCollapsed: false,
      toasts: [],
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('sidebar', () => {
    it('should start with sidebar open', () => {
      expect(useUIStore.getState().sidebarOpen).toBe(true)
    })

    it('should start with sidebar not collapsed', () => {
      expect(useUIStore.getState().sidebarCollapsed).toBe(false)
    })

    it('should toggle sidebar', () => {
      useUIStore.getState().toggleSidebar()
      expect(useUIStore.getState().sidebarOpen).toBe(false)

      useUIStore.getState().toggleSidebar()
      expect(useUIStore.getState().sidebarOpen).toBe(true)
    })

    it('should set sidebar collapsed', () => {
      useUIStore.getState().setSidebarCollapsed(true)
      expect(useUIStore.getState().sidebarCollapsed).toBe(true)

      useUIStore.getState().setSidebarCollapsed(false)
      expect(useUIStore.getState().sidebarCollapsed).toBe(false)
    })
  })

  describe('toasts', () => {
    it('should start with empty toasts', () => {
      expect(useUIStore.getState().toasts).toEqual([])
    })

    it('should add a toast', () => {
      useUIStore.getState().addToast({
        title: 'Success',
        description: 'User created',
        variant: 'success',
      })

      const toasts = useUIStore.getState().toasts
      expect(toasts).toHaveLength(1)
      expect(toasts[0].title).toBe('Success')
      expect(toasts[0].description).toBe('User created')
      expect(toasts[0].variant).toBe('success')
      expect(toasts[0].id).toBeTruthy()
    })

    it('should add multiple toasts', () => {
      useUIStore.getState().addToast({ title: 'Toast 1', variant: 'default' })
      useUIStore.getState().addToast({ title: 'Toast 2', variant: 'error' })
      useUIStore.getState().addToast({ title: 'Toast 3', variant: 'warning' })

      expect(useUIStore.getState().toasts).toHaveLength(3)
    })

    it('should assign unique IDs to each toast', () => {
      useUIStore.getState().addToast({ title: 'Toast 1', variant: 'default' })
      useUIStore.getState().addToast({ title: 'Toast 2', variant: 'default' })

      const toasts = useUIStore.getState().toasts
      expect(toasts[0].id).not.toBe(toasts[1].id)
    })

    it('should remove a toast by ID', () => {
      useUIStore.getState().addToast({ title: 'Toast 1', variant: 'default' })
      useUIStore.getState().addToast({ title: 'Toast 2', variant: 'default' })

      const toastId = useUIStore.getState().toasts[0].id
      useUIStore.getState().removeToast(toastId)

      const toasts = useUIStore.getState().toasts
      expect(toasts).toHaveLength(1)
      expect(toasts[0].title).toBe('Toast 2')
    })

    it('should auto-remove toast after default duration (5000ms)', () => {
      useUIStore.getState().addToast({ title: 'Auto', variant: 'default' })

      expect(useUIStore.getState().toasts).toHaveLength(1)

      vi.advanceTimersByTime(5000)

      expect(useUIStore.getState().toasts).toHaveLength(0)
    })

    it('should auto-remove toast after custom duration', () => {
      useUIStore.getState().addToast({
        title: 'Custom',
        variant: 'default',
        duration: 2000,
      })

      expect(useUIStore.getState().toasts).toHaveLength(1)

      vi.advanceTimersByTime(2000)

      expect(useUIStore.getState().toasts).toHaveLength(0)
    })

    it('should not remove toast before duration elapses', () => {
      useUIStore.getState().addToast({
        title: 'Wait',
        variant: 'default',
        duration: 3000,
      })

      vi.advanceTimersByTime(2000)

      expect(useUIStore.getState().toasts).toHaveLength(1)
    })

    it('should handle removing non-existent toast gracefully', () => {
      useUIStore.getState().addToast({ title: 'Toast', variant: 'default' })

      useUIStore.getState().removeToast('nonexistent-id')

      // Should still have the original toast
      expect(useUIStore.getState().toasts).toHaveLength(1)
    })

    it('should support all toast variants', () => {
      const variants = ['default', 'success', 'error', 'warning'] as const
      for (const variant of variants) {
        useUIStore.getState().addToast({ title: `${variant} toast`, variant })
      }

      const toasts = useUIStore.getState().toasts
      expect(toasts).toHaveLength(4)
      expect(toasts.map((t) => t.variant)).toEqual(['default', 'success', 'error', 'warning'])
    })
  })
})
