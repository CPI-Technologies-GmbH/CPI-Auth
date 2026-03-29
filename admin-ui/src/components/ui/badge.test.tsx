import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Badge } from './badge'

describe('Badge', () => {
  describe('rendering', () => {
    it('should render with text content', () => {
      render(<Badge>Active</Badge>)
      expect(screen.getByText('Active')).toBeInTheDocument()
    })

    it('should render as a div element', () => {
      render(<Badge data-testid="badge">Status</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.tagName).toBe('DIV')
    })
  })

  describe('variant rendering', () => {
    it('should apply default variant classes', () => {
      render(<Badge data-testid="badge">Default</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('bg-primary')
      expect(badge.className).toContain('text-primary-foreground')
    })

    it('should apply secondary variant classes', () => {
      render(<Badge data-testid="badge" variant="secondary">Secondary</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('bg-secondary')
      expect(badge.className).toContain('text-secondary-foreground')
    })

    it('should apply destructive variant classes', () => {
      render(<Badge data-testid="badge" variant="destructive">Error</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('bg-destructive')
      expect(badge.className).toContain('text-destructive-foreground')
    })

    it('should apply success variant classes', () => {
      render(<Badge data-testid="badge" variant="success">Success</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('bg-success')
      expect(badge.className).toContain('text-success')
    })

    it('should apply warning variant classes', () => {
      render(<Badge data-testid="badge" variant="warning">Warning</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('bg-warning')
      expect(badge.className).toContain('text-warning')
    })

    it('should apply outline variant classes', () => {
      render(<Badge data-testid="badge" variant="outline">Outline</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('text-foreground')
    })

    it('should apply muted variant classes', () => {
      render(<Badge data-testid="badge" variant="muted">Muted</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('bg-muted')
      expect(badge.className).toContain('text-muted-foreground')
    })
  })

  describe('common styling', () => {
    it('should have rounded-full class', () => {
      render(<Badge data-testid="badge">Pill</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('rounded-full')
    })

    it('should have text-xs class', () => {
      render(<Badge data-testid="badge">Small</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('text-xs')
    })

    it('should have font-semibold class', () => {
      render(<Badge data-testid="badge">Bold</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('font-semibold')
    })

    it('should have inline-flex class', () => {
      render(<Badge data-testid="badge">Inline</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('inline-flex')
    })
  })

  describe('custom className', () => {
    it('should accept additional className', () => {
      render(<Badge data-testid="badge" className="my-class">Custom</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.className).toContain('my-class')
    })
  })

  describe('HTML attributes', () => {
    it('should pass through HTML attributes', () => {
      render(<Badge data-testid="badge" title="Status badge">Status</Badge>)
      const badge = screen.getByTestId('badge')
      expect(badge.getAttribute('title')).toBe('Status badge')
    })
  })
})
