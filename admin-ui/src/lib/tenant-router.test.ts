import { describe, it, expect } from 'vitest'
import {
  parseTenantRoute,
  buildTenantPath,
  buildSystemPath,
  switchTenantPath,
} from './tenant-router'

describe('parseTenantRoute', () => {
  it('extracts the slug from /t/{slug}/...', () => {
    expect(parseTenantRoute('/t/lastsoftware/users')).toEqual({
      slug: 'lastsoftware',
      subPath: '/users',
    })
  })

  it('handles /t/{slug} with no trailing path', () => {
    expect(parseTenantRoute('/t/lastsoftware')).toEqual({
      slug: 'lastsoftware',
      subPath: '/',
    })
  })

  it('returns empty slug for system routes', () => {
    expect(parseTenantRoute('/system')).toEqual({ slug: '', subPath: '/system' })
    expect(parseTenantRoute('/system/tenants')).toEqual({
      slug: '',
      subPath: '/system/tenants',
    })
  })

  it('returns empty slug for login route', () => {
    expect(parseTenantRoute('/login')).toEqual({ slug: '', subPath: '/login' })
  })

  it('rejects invalid slug formats', () => {
    expect(parseTenantRoute('/t/Bad_Slug/users')).toEqual({
      slug: '',
      subPath: '/t/Bad_Slug/users',
    })
    expect(parseTenantRoute('/t/-bad/users')).toEqual({
      slug: '',
      subPath: '/t/-bad/users',
    })
  })

  it('preserves nested sub-paths', () => {
    expect(parseTenantRoute('/t/lastsoftware/users/abc-123')).toEqual({
      slug: 'lastsoftware',
      subPath: '/users/abc-123',
    })
  })
})

describe('buildTenantPath', () => {
  it('builds /t/{slug} for the root', () => {
    expect(buildTenantPath('lastsoftware')).toBe('/t/lastsoftware')
    expect(buildTenantPath('lastsoftware', '/')).toBe('/t/lastsoftware')
  })

  it('appends sub-paths', () => {
    expect(buildTenantPath('lastsoftware', '/users')).toBe('/t/lastsoftware/users')
    expect(buildTenantPath('default', '/applications/123')).toBe(
      '/t/default/applications/123'
    )
  })

  it('normalizes leading slash', () => {
    expect(buildTenantPath('default', 'users')).toBe('/t/default/users')
  })
})

describe('buildSystemPath', () => {
  it('returns /system for root', () => {
    expect(buildSystemPath()).toBe('/system')
    expect(buildSystemPath('')).toBe('/system')
    expect(buildSystemPath('/')).toBe('/system')
  })

  it('appends sub-paths', () => {
    expect(buildSystemPath('/tenants')).toBe('/system/tenants')
    expect(buildSystemPath('settings')).toBe('/system/settings')
  })
})

describe('switchTenantPath', () => {
  it('keeps the sub-path when switching tenant', () => {
    expect(switchTenantPath('/t/lastsoftware/users', 'default')).toBe(
      '/t/default/users'
    )
    expect(switchTenantPath('/t/lastsoftware/applications/abc', 'default')).toBe(
      '/t/default/applications/abc'
    )
  })

  it('lands on the tenant root when coming from a system route', () => {
    expect(switchTenantPath('/system/tenants', 'lastsoftware')).toBe(
      '/t/lastsoftware'
    )
  })

  it('lands on the tenant root when coming from /login', () => {
    expect(switchTenantPath('/login', 'lastsoftware')).toBe('/t/lastsoftware')
  })

  it('normalizes when target is the same tenant', () => {
    expect(switchTenantPath('/t/lastsoftware/users', 'lastsoftware')).toBe(
      '/t/lastsoftware/users'
    )
  })
})
