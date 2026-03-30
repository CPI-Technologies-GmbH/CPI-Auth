import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { branding, darkMode, applyBranding, toggleDarkMode } from './theme';
import type { BrandingConfig } from '$lib/api/types';

// Mock $lib/api/types — it's a type-only import, no need for runtime mock

describe('Theme store', () => {
	beforeEach(() => {
		// Reset stores
		branding.set(null);
		darkMode.set(false);

		// Reset document class list and styles
		document.documentElement.classList.remove('dark');
		document.documentElement.style.cssText = '';

		// Clean up any injected styles
		const customStyle = document.getElementById('af-custom-css');
		if (customStyle) customStyle.remove();

		// Clean up any injected link elements
		const links = document.head.querySelectorAll('link[rel="stylesheet"]');
		links.forEach((link) => link.remove());
	});

	function createBrandingConfig(overrides: Partial<BrandingConfig> = {}): BrandingConfig {
		return {
			primary_color: '#3b82f6',
			primary_hover_color: '#2563eb',
			primary_light_color: '#dbeafe',
			secondary_color: '#6b7280',
			background_color: '#ffffff',
			surface_color: '#f9fafb',
			text_color: '#111827',
			text_muted_color: '#6b7280',
			text_inverted_color: '#ffffff',
			border_color: '#e5e7eb',
			error_color: '#ef4444',
			success_color: '#22c55e',
			warning_color: '#f59e0b',
			info_color: '#3b82f6',
			logo_url: 'https://example.com/logo.png',
			logo_dark_url: 'https://example.com/logo-dark.png',
			font_family: 'Inter',
			border_radius: '8px',
			layout_mode: 'centered',
			custom_css: '',
			dark_mode: false,
			app_name: 'TestApp',
			social_providers: [],
			custom_fields: [],
			passkeys_enabled: false,
			magic_link_enabled: false,
			mfa_enabled: false,
			...overrides
		};
	}

	describe('applyBranding', () => {
		it('should set the branding store', () => {
			const config = createBrandingConfig();
			applyBranding(config);
			expect(get(branding)).toEqual(config);
		});

		it('should set CSS custom properties on document root', () => {
			const config = createBrandingConfig({
				primary_color: '#ff0000',
				background_color: '#000000',
				text_color: '#ffffff'
			});

			applyBranding(config);

			const style = document.documentElement.style;
			expect(style.getPropertyValue('--af-color-primary')).toBe('#ff0000');
			expect(style.getPropertyValue('--af-color-background')).toBe('#000000');
			expect(style.getPropertyValue('--af-color-text')).toBe('#ffffff');
		});

		it('should set all color properties', () => {
			const config = createBrandingConfig();
			applyBranding(config);

			const style = document.documentElement.style;
			expect(style.getPropertyValue('--af-color-primary')).toBe('#3b82f6');
			expect(style.getPropertyValue('--af-color-primary-hover')).toBe('#2563eb');
			expect(style.getPropertyValue('--af-color-primary-light')).toBe('#dbeafe');
			expect(style.getPropertyValue('--af-color-secondary')).toBe('#6b7280');
			expect(style.getPropertyValue('--af-color-surface')).toBe('#f9fafb');
			expect(style.getPropertyValue('--af-color-text-muted')).toBe('#6b7280');
			expect(style.getPropertyValue('--af-color-text-inverted')).toBe('#ffffff');
			expect(style.getPropertyValue('--af-color-border')).toBe('#e5e7eb');
			expect(style.getPropertyValue('--af-color-error')).toBe('#ef4444');
			expect(style.getPropertyValue('--af-color-success')).toBe('#22c55e');
			expect(style.getPropertyValue('--af-color-warning')).toBe('#f59e0b');
			expect(style.getPropertyValue('--af-color-info')).toBe('#3b82f6');
		});

		it('should set border_radius property', () => {
			const config = createBrandingConfig({ border_radius: '12px' });
			applyBranding(config);
			expect(document.documentElement.style.getPropertyValue('--af-border-radius')).toBe(
				'12px'
			);
		});

		it('should set font_family property', () => {
			const config = createBrandingConfig({ font_family: 'Roboto, sans-serif' });
			applyBranding(config);
			expect(document.documentElement.style.getPropertyValue('--af-font-family')).toBe(
				'Roboto, sans-serif'
			);
		});

		it('should enable dark mode when config.dark_mode is true', () => {
			const config = createBrandingConfig({ dark_mode: true });
			applyBranding(config);

			expect(get(darkMode)).toBe(true);
			expect(document.documentElement.classList.contains('dark')).toBe(true);
		});

		it('should not enable dark mode when config.dark_mode is false', () => {
			const config = createBrandingConfig({ dark_mode: false });
			applyBranding(config);

			expect(get(darkMode)).toBe(false);
			expect(document.documentElement.classList.contains('dark')).toBe(false);
		});

		it('should inject custom CSS', () => {
			const config = createBrandingConfig({
				custom_css: '.my-class { color: red; }'
			});
			applyBranding(config);

			const styleElement = document.getElementById('af-custom-css');
			expect(styleElement).not.toBeNull();
			expect(styleElement?.textContent).toBe('.my-class { color: red; }');
		});

		it('should replace existing custom CSS', () => {
			const config1 = createBrandingConfig({ custom_css: '.old { color: blue; }' });
			applyBranding(config1);

			const config2 = createBrandingConfig({ custom_css: '.new { color: green; }' });
			applyBranding(config2);

			const styles = document.querySelectorAll('#af-custom-css');
			expect(styles.length).toBe(1);
			expect(styles[0].textContent).toBe('.new { color: green; }');
		});

		it('should not inject custom CSS when custom_css is empty', () => {
			const config = createBrandingConfig({ custom_css: '' });
			applyBranding(config);

			const styleElement = document.getElementById('af-custom-css');
			expect(styleElement).toBeNull();
		});

		it('should load Google Font for non-Inter fonts', () => {
			const config = createBrandingConfig({ font_family: 'Roboto, sans-serif' });
			applyBranding(config);

			const links = document.head.querySelectorAll('link[rel="stylesheet"]');
			const fontLink = Array.from(links).find((l) =>
				l.getAttribute('href')?.includes('fonts.googleapis.com')
			);
			expect(fontLink).not.toBeNull();
			expect(fontLink?.getAttribute('href')).toContain('Roboto');
		});

		it('should not load Google Font for Inter (default)', () => {
			const config = createBrandingConfig({ font_family: 'Inter' });
			applyBranding(config);

			const links = document.head.querySelectorAll('link[rel="stylesheet"]');
			const fontLink = Array.from(links).find((l) =>
				l.getAttribute('href')?.includes('fonts.googleapis.com')
			);
			expect(fontLink).toBeUndefined();
		});
	});

	describe('toggleDarkMode', () => {
		it('should toggle dark mode on', () => {
			expect(get(darkMode)).toBe(false);

			toggleDarkMode();

			expect(get(darkMode)).toBe(true);
			expect(document.documentElement.classList.contains('dark')).toBe(true);
		});

		it('should toggle dark mode off', () => {
			darkMode.set(true);
			document.documentElement.classList.add('dark');

			toggleDarkMode();

			expect(get(darkMode)).toBe(false);
			expect(document.documentElement.classList.contains('dark')).toBe(false);
		});

		it('should toggle multiple times correctly', () => {
			toggleDarkMode(); // on
			expect(get(darkMode)).toBe(true);

			toggleDarkMode(); // off
			expect(get(darkMode)).toBe(false);

			toggleDarkMode(); // on
			expect(get(darkMode)).toBe(true);
		});
	});

	describe('branding store', () => {
		it('should initialize as null', () => {
			expect(get(branding)).toBeNull();
		});

		it('should be updatable', () => {
			const config = createBrandingConfig();
			branding.set(config);
			expect(get(branding)).toEqual(config);
		});
	});

	describe('darkMode store', () => {
		it('should initialize as false', () => {
			expect(get(darkMode)).toBe(false);
		});

		it('should be settable directly', () => {
			darkMode.set(true);
			expect(get(darkMode)).toBe(true);
		});
	});
});
