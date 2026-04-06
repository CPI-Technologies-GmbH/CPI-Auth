// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Error {
			message: string;
			code?: string;
		}
		interface Locals {
			csrfToken?: string;
			/** Tenant slug extracted from /t/{slug}/ URL prefix, set by hooks. */
			tenantSlug?: string;
			/** Path prefix the browser sees, e.g. "/t/lastsoftware". Empty when no prefix. */
			tenantBase?: string;
		}
		interface PageData {
			branding?: import('$lib/api/types').BrandingConfig;
			tenantSlug?: string;
			tenantBase?: string;
		}
		interface PageState {}
		interface Platform {}
	}
}

export {};
