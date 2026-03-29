export {
	validateEmail,
	validatePassword,
	validatePasswordMatch,
	validateRequired,
	validateOTP,
	getPasswordStrength,
	type PasswordStrength,
	type ValidationResult
} from './validation';

/**
 * Redirect to a URL handling both client-side and server-side scenarios.
 */
export function redirectTo(url: string) {
	if (typeof window !== 'undefined') {
		window.location.href = url;
	}
}

/**
 * Extract error message from various error types.
 */
export function getErrorMessage(error: unknown): string {
	if (error instanceof Error) {
		return error.message;
	}
	if (typeof error === 'string') {
		return error;
	}
	if (error && typeof error === 'object' && 'message' in error) {
		return String((error as { message: unknown }).message);
	}
	return 'An unexpected error occurred';
}

/**
 * Format a scope name into a human-readable description.
 */
export function formatScope(scope: string): string {
	const scopeDescriptions: Record<string, string> = {
		openid: 'Verify your identity',
		profile: 'Access your profile information',
		email: 'Access your email address',
		phone: 'Access your phone number',
		address: 'Access your address',
		offline_access: 'Maintain access when you are not present'
	};
	return scopeDescriptions[scope] || scope;
}

/**
 * Debounce a function call.
 */
export function debounce<T extends (...args: unknown[]) => unknown>(
	fn: T,
	delay: number
): (...args: Parameters<T>) => void {
	let timeoutId: ReturnType<typeof setTimeout>;
	return (...args: Parameters<T>) => {
		clearTimeout(timeoutId);
		timeoutId = setTimeout(() => fn(...args), delay);
	};
}

/**
 * Generate a CSRF token.
 */
export function generateCSRFToken(): string {
	if (typeof crypto !== 'undefined' && crypto.randomUUID) {
		return crypto.randomUUID();
	}
	return Math.random().toString(36).substring(2) + Date.now().toString(36);
}
