import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from './button'

describe('Button', () => {
  describe('rendering', () => {
    it('should render with children', () => {
      render(<Button>Click me</Button>)
      expect(screen.getByRole('button', { name: 'Click me' })).toBeInTheDocument()
    })

    it('should render as a button element', () => {
      render(<Button>Test</Button>)
      const button = screen.getByRole('button')
      expect(button.tagName).toBe('BUTTON')
    })
  })

  describe('variants', () => {
    it('should apply default variant classes', () => {
      render(<Button variant="default">Default</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('bg-primary')
    })

    it('should apply destructive variant classes', () => {
      render(<Button variant="destructive">Delete</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('bg-destructive')
    })

    it('should apply outline variant classes', () => {
      render(<Button variant="outline">Outline</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('border')
    })

    it('should apply secondary variant classes', () => {
      render(<Button variant="secondary">Secondary</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('bg-secondary')
    })

    it('should apply ghost variant classes', () => {
      render(<Button variant="ghost">Ghost</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('hover:bg-accent')
    })

    it('should apply link variant classes', () => {
      render(<Button variant="link">Link</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('underline-offset-4')
    })

    it('should apply success variant classes', () => {
      render(<Button variant="success">Success</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('bg-success')
    })
  })

  describe('sizes', () => {
    it('should apply default size classes', () => {
      render(<Button size="default">Default Size</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('h-10')
      expect(button.className).toContain('px-4')
    })

    it('should apply sm size classes', () => {
      render(<Button size="sm">Small</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('h-9')
      expect(button.className).toContain('px-3')
    })

    it('should apply lg size classes', () => {
      render(<Button size="lg">Large</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('h-11')
      expect(button.className).toContain('px-8')
    })

    it('should apply icon size classes', () => {
      render(<Button size="icon">Icon</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('h-10')
      expect(button.className).toContain('w-10')
    })
  })

  describe('loading state', () => {
    it('should render loading spinner when loading is true', () => {
      render(<Button loading>Loading</Button>)
      const button = screen.getByRole('button')
      // The Loader2 icon should be present with animate-spin
      const spinner = button.querySelector('.animate-spin')
      expect(spinner).not.toBeNull()
    })

    it('should disable button when loading', () => {
      render(<Button loading>Loading</Button>)
      const button = screen.getByRole('button')
      expect(button).toBeDisabled()
    })

    it('should still show children text when loading', () => {
      render(<Button loading>Submit</Button>)
      expect(screen.getByText('Submit')).toBeInTheDocument()
    })

    it('should not show spinner when not loading', () => {
      render(<Button>Submit</Button>)
      const button = screen.getByRole('button')
      const spinner = button.querySelector('.animate-spin')
      expect(spinner).toBeNull()
    })
  })

  describe('disabled state', () => {
    it('should be disabled when disabled prop is true', () => {
      render(<Button disabled>Disabled</Button>)
      expect(screen.getByRole('button')).toBeDisabled()
    })

    it('should not be disabled by default', () => {
      render(<Button>Enabled</Button>)
      expect(screen.getByRole('button')).not.toBeDisabled()
    })

    it('should have disabled styling classes', () => {
      render(<Button disabled>Disabled</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('disabled:opacity-50')
    })
  })

  describe('click handler', () => {
    it('should call onClick when clicked', async () => {
      const handleClick = vi.fn()
      const user = userEvent.setup()

      render(<Button onClick={handleClick}>Click</Button>)
      await user.click(screen.getByRole('button'))

      expect(handleClick).toHaveBeenCalledOnce()
    })

    it('should not call onClick when disabled', async () => {
      const handleClick = vi.fn()
      const user = userEvent.setup()

      render(<Button onClick={handleClick} disabled>Click</Button>)
      await user.click(screen.getByRole('button'))

      expect(handleClick).not.toHaveBeenCalled()
    })

    it('should not call onClick when loading', async () => {
      const handleClick = vi.fn()
      const user = userEvent.setup()

      render(<Button onClick={handleClick} loading>Click</Button>)
      await user.click(screen.getByRole('button'))

      expect(handleClick).not.toHaveBeenCalled()
    })
  })

  describe('custom className', () => {
    it('should accept additional className', () => {
      render(<Button className="my-custom-class">Custom</Button>)
      const button = screen.getByRole('button')
      expect(button.className).toContain('my-custom-class')
    })
  })

  describe('ref forwarding', () => {
    it('should forward ref to button element', () => {
      const ref = vi.fn()
      render(<Button ref={ref}>Ref Test</Button>)
      expect(ref).toHaveBeenCalledWith(expect.any(HTMLButtonElement))
    })
  })

  describe('displayName', () => {
    it('should have displayName set to "Button"', () => {
      expect(Button.displayName).toBe('Button')
    })
  })
})
