// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		interface Error {
			message: string;
			code?: string;
		}
		interface Locals {
			csrfToken?: string;
		}
		interface PageData {
			branding?: import('$lib/api/types').BrandingConfig;
		}
		interface PageState {}
		interface Platform {}
	}
}

export {};
