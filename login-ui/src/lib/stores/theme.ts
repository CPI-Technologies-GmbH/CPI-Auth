import { writable, get } from 'svelte/store';
import type { BrandingConfig } from '$lib/api/types';

export const branding = writable<BrandingConfig | null>(null);
export const darkMode = writable<boolean>(false);

export function applyBranding(config: BrandingConfig) {
	branding.set(config);

	if (typeof document === 'undefined') return;

	const root = document.documentElement;
	const style = root.style;

	if (config.primary_color) style.setProperty('--af-color-primary', config.primary_color);
	if (config.primary_hover_color)
		style.setProperty('--af-color-primary-hover', config.primary_hover_color);
	if (config.primary_light_color)
		style.setProperty('--af-color-primary-light', config.primary_light_color);
	if (config.secondary_color) style.setProperty('--af-color-secondary', config.secondary_color);
	if (config.background_color)
		style.setProperty('--af-color-background', config.background_color);
	if (config.surface_color) style.setProperty('--af-color-surface', config.surface_color);
	if (config.text_color) style.setProperty('--af-color-text', config.text_color);
	if (config.text_muted_color)
		style.setProperty('--af-color-text-muted', config.text_muted_color);
	if (config.text_inverted_color)
		style.setProperty('--af-color-text-inverted', config.text_inverted_color);
	if (config.border_color) style.setProperty('--af-color-border', config.border_color);
	if (config.error_color) style.setProperty('--af-color-error', config.error_color);
	if (config.success_color) style.setProperty('--af-color-success', config.success_color);
	if (config.warning_color) style.setProperty('--af-color-warning', config.warning_color);
	if (config.info_color) style.setProperty('--af-color-info', config.info_color);
	if (config.border_radius) style.setProperty('--af-border-radius', config.border_radius);

	if (config.font_family) {
		style.setProperty('--af-font-family', config.font_family);
		loadGoogleFont(config.font_family);
	}

	if (config.dark_mode) {
		root.classList.add('dark');
		darkMode.set(true);
	}

	if (config.custom_css) {
		injectCustomCSS(config.custom_css);
	}
}

export function toggleDarkMode() {
	const isDark = !get(darkMode);
	darkMode.set(isDark);

	if (typeof document === 'undefined') return;

	if (isDark) {
		document.documentElement.classList.add('dark');
	} else {
		document.documentElement.classList.remove('dark');
	}
}

function loadGoogleFont(fontFamily: string) {
	const primaryFont = fontFamily.split(',')[0].trim().replace(/['"]/g, '');
	if (!primaryFont || primaryFont === 'Inter') return;

	const link = document.createElement('link');
	link.rel = 'stylesheet';
	link.href = `https://fonts.googleapis.com/css2?family=${encodeURIComponent(primaryFont)}:wght@300;400;500;600;700&display=swap`;
	document.head.appendChild(link);
}

function injectCustomCSS(css: string) {
	const existingStyle = document.getElementById('af-custom-css');
	if (existingStyle) existingStyle.remove();

	const style = document.createElement('style');
	style.id = 'af-custom-css';
	style.textContent = css;
	document.head.appendChild(style);
}
