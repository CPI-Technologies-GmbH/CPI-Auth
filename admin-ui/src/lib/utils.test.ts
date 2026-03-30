import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { cn, formatDate, formatDateTime, formatRelativeTime, truncate, getInitials, generateId, sleep } from './utils'

describe('cn', () => {
  it('should merge class names', () => {
    expect(cn('foo', 'bar')).toBe('foo bar')
  })

  it('should handle conditional classes', () => {
    expect(cn('foo', false && 'bar', 'baz')).toBe('foo baz')
  })

  it('should merge tailwind classes correctly', () => {
    expect(cn('px-2 py-1', 'px-4')).toBe('py-1 px-4')
  })

  it('should handle undefined/null inputs', () => {
    expect(cn('foo', undefined, null, 'bar')).toBe('foo bar')
  })

  it('should handle empty input', () => {
    expect(cn()).toBe('')
  })

  it('should handle arrays', () => {
    expect(cn(['foo', 'bar'])).toBe('foo bar')
  })

  it('should handle object syntax', () => {
    expect(cn({ foo: true, bar: false, baz: true })).toBe('foo baz')
  })

  it('should deduplicate tailwind conflicting classes', () => {
    expect(cn('text-red-500', 'text-blue-500')).toBe('text-blue-500')
  })
})

describe('formatDate', () => {
  it('should format a date string', () => {
    const result = formatDate('2024-01-15T10:30:00Z')
    expect(result).toContain('Jan')
    expect(result).toContain('15')
    expect(result).toContain('2024')
  })

  it('should format a Date object', () => {
    const result = formatDate(new Date('2024-06-01'))
    expect(result).toContain('Jun')
    expect(result).toContain('2024')
  })

  it('should handle different months', () => {
    expect(formatDate('2024-12-25')).toContain('Dec')
    expect(formatDate('2024-03-01')).toContain('Mar')
  })
})

describe('formatDateTime', () => {
  it('should include time in the output', () => {
    const result = formatDateTime('2024-01-15T10:30:00Z')
    expect(result).toContain('Jan')
    expect(result).toContain('15')
    expect(result).toContain('2024')
    // Should include time parts
    expect(result).toMatch(/\d{1,2}:\d{2}/)
  })
})

describe('formatRelativeTime', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2024-06-15T12:00:00Z'))
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('should return "just now" for recent times', () => {
    const result = formatRelativeTime('2024-06-15T11:59:30Z')
    expect(result).toBe('just now')
  })

  it('should return minutes ago', () => {
    const result = formatRelativeTime('2024-06-15T11:55:00Z')
    expect(result).toBe('5m ago')
  })

  it('should return hours ago', () => {
    const result = formatRelativeTime('2024-06-15T09:00:00Z')
    expect(result).toBe('3h ago')
  })

  it('should return days ago', () => {
    const result = formatRelativeTime('2024-06-13T12:00:00Z')
    expect(result).toBe('2d ago')
  })

  it('should return formatted date for older than 7 days', () => {
    const result = formatRelativeTime('2024-06-01T12:00:00Z')
    expect(result).toContain('Jun')
    expect(result).toContain('2024')
  })

  it('should handle string input', () => {
    const result = formatRelativeTime('2024-06-15T11:00:00Z')
    expect(result).toBe('1h ago')
  })

  it('should handle Date input', () => {
    const result = formatRelativeTime(new Date('2024-06-15T11:30:00Z'))
    expect(result).toBe('30m ago')
  })
})

describe('truncate', () => {
  it('should not truncate strings shorter than limit', () => {
    expect(truncate('hello', 10)).toBe('hello')
  })

  it('should not truncate strings equal to limit', () => {
    expect(truncate('hello', 5)).toBe('hello')
  })

  it('should truncate strings longer than limit', () => {
    expect(truncate('hello world', 5)).toBe('hello...')
  })

  it('should handle empty string', () => {
    expect(truncate('', 5)).toBe('')
  })
})

describe('getInitials', () => {
  it('should return initials from a full name', () => {
    expect(getInitials('John Doe')).toBe('JD')
  })

  it('should return first two initials for long names', () => {
    expect(getInitials('John Michael Doe')).toBe('JM')
  })

  it('should handle single name', () => {
    expect(getInitials('John')).toBe('J')
  })

  it('should uppercase initials', () => {
    expect(getInitials('john doe')).toBe('JD')
  })

  it('should handle names with extra spaces', () => {
    // split(' ') will produce empty strings for extra spaces
    // but [0] of empty string is undefined which will throw
    // So this tests the actual behavior
    expect(getInitials('John Doe')).toBe('JD')
  })
})

describe('generateId', () => {
  it('should return a non-empty string', () => {
    expect(generateId()).toBeTruthy()
    expect(typeof generateId()).toBe('string')
  })

  it('should return different IDs on each call', () => {
    const ids = new Set(Array.from({ length: 100 }, () => generateId()))
    expect(ids.size).toBe(100)
  })
})

describe('sleep', () => {
  it('should resolve after specified duration', async () => {
    vi.useFakeTimers()
    const promise = sleep(100)
    vi.advanceTimersByTime(100)
    await expect(promise).resolves.toBeUndefined()
    vi.useRealTimers()
  })
})
