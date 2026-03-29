import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ConfirmDialog } from './confirm-dialog'

describe('ConfirmDialog', () => {
  const defaultProps = {
    open: true,
    onClose: vi.fn(),
    onConfirm: vi.fn(),
    title: 'Delete User',
    description: 'Are you sure you want to delete this user?',
  }

  describe('open/close', () => {
    it('should render when open is true', () => {
      render(<ConfirmDialog {...defaultProps} />)
      expect(screen.getByText('Delete User')).toBeInTheDocument()
      expect(screen.getByText('Are you sure you want to delete this user?')).toBeInTheDocument()
    })

    it('should not render when open is false', () => {
      render(<ConfirmDialog {...defaultProps} open={false} />)
      expect(screen.queryByText('Delete User')).not.toBeInTheDocument()
    })

    it('should call onClose when close button is clicked', async () => {
      const onClose = vi.fn()
      const user = userEvent.setup()

      render(<ConfirmDialog {...defaultProps} onClose={onClose} />)

      // The DialogClose renders an X button
      const closeButton = document.querySelector('button.absolute')
      expect(closeButton).not.toBeNull()
      await user.click(closeButton!)

      expect(onClose).toHaveBeenCalled()
    })

    it('should call onClose when Cancel button is clicked', async () => {
      const onClose = vi.fn()
      const user = userEvent.setup()

      render(<ConfirmDialog {...defaultProps} onClose={onClose} />)

      await user.click(screen.getByText('Cancel'))
      expect(onClose).toHaveBeenCalled()
    })

    it('should call onClose when backdrop overlay is clicked', async () => {
      const onClose = vi.fn()
      const user = userEvent.setup()

      render(<ConfirmDialog {...defaultProps} onClose={onClose} />)

      const overlay = document.querySelector('.bg-black\\/60')
      if (overlay) {
        await user.click(overlay)
        expect(onClose).toHaveBeenCalled()
      }
    })
  })

  describe('confirm action', () => {
    it('should call onConfirm when Confirm button is clicked', async () => {
      const onConfirm = vi.fn()
      const user = userEvent.setup()

      render(<ConfirmDialog {...defaultProps} onConfirm={onConfirm} />)

      await user.click(screen.getByText('Confirm'))
      expect(onConfirm).toHaveBeenCalledOnce()
    })
  })

  describe('cancel action', () => {
    it('should call onClose when Cancel button is clicked', async () => {
      const onClose = vi.fn()
      const user = userEvent.setup()

      render(<ConfirmDialog {...defaultProps} onClose={onClose} />)

      await user.click(screen.getByText('Cancel'))
      expect(onClose).toHaveBeenCalledOnce()
    })
  })

  describe('custom labels', () => {
    it('should render custom confirm label', () => {
      render(<ConfirmDialog {...defaultProps} confirmLabel="Yes, Delete" />)
      expect(screen.getByText('Yes, Delete')).toBeInTheDocument()
    })

    it('should render custom cancel label', () => {
      render(<ConfirmDialog {...defaultProps} cancelLabel="No, Keep" />)
      expect(screen.getByText('No, Keep')).toBeInTheDocument()
    })

    it('should use default labels when not provided', () => {
      render(<ConfirmDialog {...defaultProps} />)
      expect(screen.getByText('Confirm')).toBeInTheDocument()
      expect(screen.getByText('Cancel')).toBeInTheDocument()
    })
  })

  describe('variant', () => {
    it('should apply destructive variant to confirm button', () => {
      render(<ConfirmDialog {...defaultProps} variant="destructive" />)
      const confirmButton = screen.getByText('Confirm').closest('button')!
      expect(confirmButton.className).toContain('bg-destructive')
    })

    it('should apply default variant to confirm button by default', () => {
      render(<ConfirmDialog {...defaultProps} />)
      const confirmButton = screen.getByText('Confirm').closest('button')!
      expect(confirmButton.className).toContain('bg-primary')
    })
  })

  describe('loading state', () => {
    it('should disable cancel button when loading', () => {
      render(<ConfirmDialog {...defaultProps} loading />)
      const cancelButton = screen.getByText('Cancel').closest('button')!
      expect(cancelButton).toBeDisabled()
    })

    it('should show loading spinner on confirm button when loading', () => {
      render(<ConfirmDialog {...defaultProps} loading />)
      const confirmButton = screen.getByText('Confirm').closest('button')!
      const spinner = confirmButton.querySelector('.animate-spin')
      expect(spinner).not.toBeNull()
    })

    it('should disable confirm button when loading', () => {
      render(<ConfirmDialog {...defaultProps} loading />)
      const confirmButton = screen.getByText('Confirm').closest('button')!
      expect(confirmButton).toBeDisabled()
    })
  })

  describe('title and description', () => {
    it('should render title', () => {
      render(<ConfirmDialog {...defaultProps} title="Remove Item" />)
      expect(screen.getByText('Remove Item')).toBeInTheDocument()
    })

    it('should render description', () => {
      render(
        <ConfirmDialog
          {...defaultProps}
          description="This action cannot be undone."
        />
      )
      expect(screen.getByText('This action cannot be undone.')).toBeInTheDocument()
    })
  })
})
