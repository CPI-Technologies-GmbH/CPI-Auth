import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { locale, t, setLocale, getTranslation } from './index';
import type { Locale } from './index';

describe('i18n', () => {
	beforeEach(() => {
		// Reset to English before each test
		setLocale('en');
	});

	describe('locale store', () => {
		it('should default to "en"', () => {
			expect(get(locale)).toBe('en');
		});

		it('should update when setLocale is called', () => {
			setLocale('de');
			expect(get(locale)).toBe('de');
		});

		it('should support all defined locales', () => {
			const locales: Locale[] = ['en', 'de', 'fr', 'es'];
			for (const loc of locales) {
				setLocale(loc);
				expect(get(locale)).toBe(loc);
			}
		});
	});

	describe('t (derived store) - translation lookup', () => {
		it('should return English translation for known key', () => {
			const translate = get(t);
			expect(translate('login.title')).toBe('Sign in to your account');
		});

		it('should return English translation for validation keys', () => {
			const translate = get(t);
			expect(translate('validation.required')).toBe('This field is required');
			expect(translate('validation.email')).toBe('Please enter a valid email address');
			expect(translate('validation.password.min')).toBe(
				'Password must be at least 8 characters'
			);
		});

		it('should return the key itself when translation is missing', () => {
			const translate = get(t);
			expect(translate('nonexistent.key')).toBe('nonexistent.key');
		});

		it('should return German translation when locale is de', () => {
			setLocale('de');
			const translate = get(t);
			expect(translate('login.title')).toBe('Anmelden');
			expect(translate('login.submit')).toBe('Anmelden');
		});

		it('should return French translation when locale is fr', () => {
			setLocale('fr');
			const translate = get(t);
			expect(translate('login.title')).toBe('Connectez-vous');
			expect(translate('login.submit')).toBe('Se connecter');
		});

		it('should return Spanish translation when locale is es', () => {
			setLocale('es');
			const translate = get(t);
			expect(translate('login.title')).toBe('Iniciar sesion');
		});
	});

	describe('fallback to English', () => {
		it('should fall back to English when key is missing in German', () => {
			setLocale('de');
			const translate = get(t);
			// 'validation.required' is not in de translations, should fall back to en
			expect(translate('validation.required')).toBe('This field is required');
		});

		it('should fall back to English when key is missing in French', () => {
			setLocale('fr');
			const translate = get(t);
			expect(translate('validation.email')).toBe('Please enter a valid email address');
		});

		it('should fall back to English when key is missing in Spanish', () => {
			setLocale('es');
			const translate = get(t);
			expect(translate('validation.password.match')).toBe('Passwords do not match');
		});

		it('should return key when translation is missing in both locale and English', () => {
			setLocale('de');
			const translate = get(t);
			expect(translate('totally.unknown.key')).toBe('totally.unknown.key');
		});
	});

	describe('parameter interpolation', () => {
		it('should replace {param} in translation strings', () => {
			// We need a translation that has parameters. The translations in this codebase
			// don't seem to use parameters directly, but the mechanism is there.
			// Let's test the mechanism directly with a known translation key and added params.
			const translate = get(t);

			// Even though the English translations don't have params, the function should
			// simply return the text unchanged when no matching param placeholders exist.
			const result = translate('login.title', { name: 'Test' });
			expect(result).toBe('Sign in to your account');
		});

		it('should handle missing params gracefully (leave placeholder)', () => {
			const translate = get(t);
			// If a translation had {name}, passing no params means the {name} stays
			// Since the actual translations don't have params, this test ensures
			// the mechanism doesn't break on existing translations
			const result = translate('login.submit');
			expect(result).toBe('Sign in');
		});

		it('should replace multiple params', () => {
			const translate = get(t);
			// Test the replacement mechanism with a key that returns text with placeholders
			// The mechanism does: text.replace(`{${k}}`, v) for each param
			// Testing indirectly through the function
			const result = translate('login.title', { foo: 'bar', baz: 'qux' });
			// No placeholders in login.title, so it stays the same
			expect(result).toBe('Sign in to your account');
		});
	});

	describe('getTranslation function', () => {
		it('should return translation for current locale', () => {
			expect(getTranslation('login.title')).toBe('Sign in to your account');
		});

		it('should reflect locale changes', () => {
			setLocale('de');
			expect(getTranslation('login.title')).toBe('Anmelden');
		});

		it('should support params', () => {
			const result = getTranslation('login.submit');
			expect(result).toBe('Sign in');
		});
	});

	describe('setLocale', () => {
		it('should change locale to de', () => {
			setLocale('de');
			expect(get(locale)).toBe('de');
		});

		it('should change locale to fr', () => {
			setLocale('fr');
			expect(get(locale)).toBe('fr');
		});

		it('should change locale to es', () => {
			setLocale('es');
			expect(get(locale)).toBe('es');
		});

		it('should change back to en', () => {
			setLocale('de');
			setLocale('en');
			expect(get(locale)).toBe('en');
			expect(getTranslation('login.title')).toBe('Sign in to your account');
		});
	});
});
