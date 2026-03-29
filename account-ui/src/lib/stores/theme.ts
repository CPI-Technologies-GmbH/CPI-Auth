import { writable } from 'svelte/store';
import type { Branding } from '$lib/types';

export const branding = writable<Branding | null>(null);

export function applyBranding(b: Branding) {
	if (typeof document === 'undefined') return;

	const root = document.documentElement;

	if (b.primary_color) {
		root.style.setProperty('--color-primary', b.primary_color);
	}
	if (b.background_color) {
		root.style.setProperty('--color-bg', b.background_color);
	}
	if (b.font_family) {
		root.style.setProperty('--font-sans', b.font_family);
	}
	if (b.favicon_url) {
		const link = document.querySelector("link[rel='icon']") as HTMLLinkElement;
		if (link) {
			link.href = b.favicon_url;
		}
	}
}
