import type { DesignTokens, ContrastResult } from './types.js'

const PREFIX = 'af'

export function buildCSS(tokens: DesignTokens): string {
  const lines: string[] = [':root {']

  if (tokens.colors) {
    for (const [key, value] of Object.entries(tokens.colors)) {
      lines.push(`  --${PREFIX}-color-${key}: ${value};`)
    }
  }
  if (tokens.spacing) {
    for (const [key, value] of Object.entries(tokens.spacing)) {
      lines.push(`  --${PREFIX}-spacing-${key}: ${value};`)
    }
  }
  if (tokens.radius) {
    for (const [key, value] of Object.entries(tokens.radius)) {
      lines.push(`  --${PREFIX}-radius-${key}: ${value};`)
    }
  }
  if (tokens.typography) {
    for (const [key, value] of Object.entries(tokens.typography)) {
      lines.push(`  --${PREFIX}-${key}: ${value};`)
    }
  }

  lines.push('}')
  return lines.join('\n')
}

export function parseHexColor(hex: string): { r: number; g: number; b: number } | null {
  const match = hex.match(/^#?([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i)
  if (!match) return null
  return { r: parseInt(match[1], 16), g: parseInt(match[2], 16), b: parseInt(match[3], 16) }
}

function luminance(r: number, g: number, b: number): number {
  const [rs, gs, bs] = [r, g, b].map((c) => {
    c /= 255
    return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
  })
  return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs
}

export function contrastRatio(hex1: string, hex2: string): number | null {
  const c1 = parseHexColor(hex1)
  const c2 = parseHexColor(hex2)
  if (!c1 || !c2) return null
  const l1 = luminance(c1.r, c1.g, c1.b)
  const l2 = luminance(c2.r, c2.g, c2.b)
  const lighter = Math.max(l1, l2)
  const darker = Math.min(l1, l2)
  return (lighter + 0.05) / (darker + 0.05)
}

export function validateContrasts(tokens: DesignTokens): ContrastResult[] {
  const results: ContrastResult[] = []
  const colors = tokens.colors
  if (!colors) return results

  const textPairs: [string, string][] = [
    ['text', 'background'],
    ['text', 'surface'],
    ['text-muted', 'background'],
    ['text-muted', 'surface'],
    ['primary', 'background'],
    ['primary', 'surface'],
    ['error', 'background'],
    ['success', 'background'],
  ]

  for (const [fg, bg] of textPairs) {
    if (!colors[fg] || !colors[bg]) continue
    const ratio = contrastRatio(colors[fg], colors[bg])
    if (ratio === null) continue
    results.push({
      pair: [fg, bg],
      ratio: Math.round(ratio * 100) / 100,
      aa: ratio >= 4.5,
      aaLarge: ratio >= 3,
      aaa: ratio >= 7,
    })
  }

  return results
}
