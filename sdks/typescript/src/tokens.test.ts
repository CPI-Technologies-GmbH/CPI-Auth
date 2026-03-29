import { describe, it, expect } from 'vitest'
import { buildCSS, parseHexColor, contrastRatio, validateContrasts } from './tokens.js'
import type { DesignTokens } from './types.js'

describe('buildCSS', () => {
  it('should generate CSS custom properties from color tokens', () => {
    const tokens: DesignTokens = {
      colors: { primary: '#6366f1', background: '#0f172a' },
    }
    const css = buildCSS(tokens)
    expect(css).toContain(':root {')
    expect(css).toContain('--af-color-primary: #6366f1;')
    expect(css).toContain('--af-color-background: #0f172a;')
    expect(css).toContain('}')
  })

  it('should generate spacing tokens', () => {
    const tokens: DesignTokens = {
      spacing: { sm: '0.5rem', md: '1rem', lg: '1.5rem' },
    }
    const css = buildCSS(tokens)
    expect(css).toContain('--af-spacing-sm: 0.5rem;')
    expect(css).toContain('--af-spacing-md: 1rem;')
    expect(css).toContain('--af-spacing-lg: 1.5rem;')
  })

  it('should generate radius tokens', () => {
    const tokens: DesignTokens = {
      radius: { sm: '4px', md: '8px' },
    }
    const css = buildCSS(tokens)
    expect(css).toContain('--af-radius-sm: 4px;')
    expect(css).toContain('--af-radius-md: 8px;')
  })

  it('should generate typography tokens', () => {
    const tokens: DesignTokens = {
      typography: { 'font-family': 'Inter, sans-serif', 'font-size-base': '16px' },
    }
    const css = buildCSS(tokens)
    expect(css).toContain('--af-font-family: Inter, sans-serif;')
    expect(css).toContain('--af-font-size-base: 16px;')
  })

  it('should combine all token types', () => {
    const tokens: DesignTokens = {
      colors: { primary: '#ff0000' },
      spacing: { md: '1rem' },
      radius: { md: '8px' },
      typography: { 'font-family': 'Arial' },
    }
    const css = buildCSS(tokens)
    const lines = css.split('\n').filter((l) => l.includes('--af-'))
    expect(lines).toHaveLength(4)
  })

  it('should return empty root for empty tokens', () => {
    const css = buildCSS({})
    expect(css).toBe(':root {\n}')
  })
})

describe('parseHexColor', () => {
  it('should parse 6-digit hex with #', () => {
    expect(parseHexColor('#ff0000')).toEqual({ r: 255, g: 0, b: 0 })
  })

  it('should parse 6-digit hex without #', () => {
    expect(parseHexColor('00ff00')).toEqual({ r: 0, g: 255, b: 0 })
  })

  it('should parse mixed case', () => {
    expect(parseHexColor('#FfAa33')).toEqual({ r: 255, g: 170, b: 51 })
  })

  it('should return null for invalid hex', () => {
    expect(parseHexColor('not-a-color')).toBeNull()
    expect(parseHexColor('#fff')).toBeNull() // 3-digit not supported
    expect(parseHexColor('')).toBeNull()
  })
})

describe('contrastRatio', () => {
  it('should return 21:1 for black on white', () => {
    const ratio = contrastRatio('#000000', '#ffffff')
    expect(ratio).toBeCloseTo(21, 0)
  })

  it('should return 1:1 for same color', () => {
    const ratio = contrastRatio('#ff0000', '#ff0000')
    expect(ratio).toBeCloseTo(1, 0)
  })

  it('should be symmetric (order independent)', () => {
    const r1 = contrastRatio('#000000', '#ffffff')
    const r2 = contrastRatio('#ffffff', '#000000')
    expect(r1).toEqual(r2)
  })

  it('should return null for invalid colors', () => {
    expect(contrastRatio('invalid', '#ffffff')).toBeNull()
    expect(contrastRatio('#000000', 'bad')).toBeNull()
  })

  it('should calculate realistic contrast for indigo on dark bg', () => {
    const ratio = contrastRatio('#6366f1', '#0f172a')
    expect(ratio).toBeGreaterThan(3) // AA Large passes
    expect(ratio).toBeLessThan(5) // AA normal fails
  })
})

describe('validateContrasts', () => {
  it('should check text-on-background pairs', () => {
    const tokens: DesignTokens = {
      colors: {
        text: '#e2e8f0',
        'text-muted': '#94a3b8',
        background: '#0f172a',
        surface: '#1e293b',
        primary: '#6366f1',
        error: '#f87171',
        success: '#34d399',
      },
    }
    const results = validateContrasts(tokens)
    expect(results.length).toBeGreaterThan(0)

    // text on background should pass AA
    const textOnBg = results.find((r) => r.pair[0] === 'text' && r.pair[1] === 'background')
    expect(textOnBg).toBeDefined()
    expect(textOnBg!.aa).toBe(true)
    expect(textOnBg!.ratio).toBeGreaterThan(4.5)

    // primary on background should fail AA normal
    const primaryOnBg = results.find((r) => r.pair[0] === 'primary' && r.pair[1] === 'background')
    expect(primaryOnBg).toBeDefined()
    expect(primaryOnBg!.aa).toBe(false)
    expect(primaryOnBg!.aaLarge).toBe(true) // But passes AA Large
  })

  it('should return empty for missing color tokens', () => {
    expect(validateContrasts({})).toEqual([])
    expect(validateContrasts({ colors: { primary: '#000' } })).toEqual([])
  })

  it('should identify AAA compliance', () => {
    const tokens: DesignTokens = {
      colors: { text: '#000000', background: '#ffffff' },
    }
    const results = validateContrasts(tokens)
    const textOnBg = results.find((r) => r.pair[0] === 'text')
    expect(textOnBg!.aaa).toBe(true) // 21:1 ratio
  })
})
