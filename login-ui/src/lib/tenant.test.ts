import { describe, it, expect } from 'vitest';
import { parseTenantFromPath, stripTenantPrefix, tenantPath } from './tenant';

describe('parseTenantFromPath', () => {
	it('extracts the slug from /t/{slug}/...', () => {
		expect(parseTenantFromPath('/t/lastsoftware/login')).toEqual({
			slug: 'lastsoftware',
			base: '/t/lastsoftware'
		});
	});

	it('handles /t/{slug} with no trailing slash', () => {
		expect(parseTenantFromPath('/t/lastsoftware')).toEqual({
			slug: 'lastsoftware',
			base: '/t/lastsoftware'
		});
	});

	it('returns empty for non-tenant paths', () => {
		expect(parseTenantFromPath('/login')).toEqual({ slug: '', base: '' });
		expect(parseTenantFromPath('/')).toEqual({ slug: '', base: '' });
	});

	it('rejects invalid slug formats', () => {
		// Uppercase, underscores, leading dash — invalid slug shapes are
		// treated as "no tenant" so we don't accidentally route to a fake tenant.
		expect(parseTenantFromPath('/t/Bad_Slug/login')).toEqual({ slug: '', base: '' });
		expect(parseTenantFromPath('/t/-bad/login')).toEqual({ slug: '', base: '' });
		expect(parseTenantFromPath('/t//login')).toEqual({ slug: '', base: '' });
	});

	it('accepts dashed lowercase slugs', () => {
		expect(parseTenantFromPath('/t/foo-bar-1/login')).toEqual({
			slug: 'foo-bar-1',
			base: '/t/foo-bar-1'
		});
	});
});

describe('stripTenantPrefix', () => {
	it('removes the /t/{slug} prefix', () => {
		expect(stripTenantPrefix('/t/lastsoftware/login')).toBe('/login');
		expect(stripTenantPrefix('/t/lastsoftware/oauth/authorize')).toBe('/oauth/authorize');
	});

	it('returns / for /t/{slug} alone', () => {
		expect(stripTenantPrefix('/t/lastsoftware')).toBe('/');
	});

	it('returns the original path for non-tenant URLs', () => {
		expect(stripTenantPrefix('/login')).toBe('/login');
		expect(stripTenantPrefix('/')).toBe('/');
	});

	it('returns the original path for invalid slugs', () => {
		expect(stripTenantPrefix('/t/Bad_Slug/login')).toBe('/t/Bad_Slug/login');
	});
});

describe('tenantPath', () => {
	it('prepends the base to a same-origin path', () => {
		expect(tenantPath('/t/lastsoftware', '/login')).toBe('/t/lastsoftware/login');
		expect(tenantPath('/t/lastsoftware', '/forgot-password')).toBe(
			'/t/lastsoftware/forgot-password'
		);
	});

	it('returns the path unchanged when there is no base', () => {
		expect(tenantPath('', '/login')).toBe('/login');
		expect(tenantPath(undefined, '/login')).toBe('/login');
	});

	it('does not double-prefix if the base is already present', () => {
		expect(tenantPath('/t/lastsoftware', '/t/lastsoftware/login')).toBe(
			'/t/lastsoftware/login'
		);
	});

	it('leaves absolute URLs alone', () => {
		expect(tenantPath('/t/lastsoftware', 'https://elsewhere.com/cb')).toBe(
			'https://elsewhere.com/cb'
		);
	});

	it('leaves protocol-relative URLs alone', () => {
		expect(tenantPath('/t/lastsoftware', '//elsewhere.com/cb')).toBe('//elsewhere.com/cb');
	});
});
