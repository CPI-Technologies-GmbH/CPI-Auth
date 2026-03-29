export interface ValidationResult {
	valid: boolean;
	error?: string;
}

export function validateEmail(email: string): ValidationResult {
	if (!email.trim()) {
		return { valid: false, error: 'validation.required' };
	}
	const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
	if (!emailRegex.test(email)) {
		return { valid: false, error: 'validation.email' };
	}
	return { valid: true };
}

export function validatePassword(password: string): ValidationResult {
	if (!password) {
		return { valid: false, error: 'validation.required' };
	}
	if (password.length < 8) {
		return { valid: false, error: 'validation.password.min' };
	}
	return { valid: true };
}

export function validatePasswordMatch(password: string, confirm: string): ValidationResult {
	if (!confirm) {
		return { valid: false, error: 'validation.required' };
	}
	if (password !== confirm) {
		return { valid: false, error: 'validation.password.match' };
	}
	return { valid: true };
}

export function validateRequired(value: string): ValidationResult {
	if (!value.trim()) {
		return { valid: false, error: 'validation.required' };
	}
	return { valid: true };
}

export function validateOTP(code: string): ValidationResult {
	if (!code) {
		return { valid: false, error: 'validation.required' };
	}
	if (!/^\d{6}$/.test(code)) {
		return { valid: false, error: 'validation.code' };
	}
	return { valid: true };
}

export interface PasswordStrength {
	score: 0 | 1 | 2 | 3 | 4;
	label: string;
	color: string;
	cssClass: string;
}

export function getPasswordStrength(password: string): PasswordStrength {
	if (!password) {
		return { score: 0, label: '', color: '', cssClass: '' };
	}

	let score = 0;

	if (password.length >= 8) score++;
	if (password.length >= 12) score++;
	if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++;
	if (/\d/.test(password)) score++;
	if (/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) score++;

	// Cap at 4
	score = Math.min(score, 4) as 0 | 1 | 2 | 3 | 4;

	const levels: Record<number, Omit<PasswordStrength, 'score'>> = {
		0: { label: '', color: '', cssClass: '' },
		1: { label: 'password.strength.weak', color: 'var(--af-color-error)', cssClass: 'strength-weak' },
		2: { label: 'password.strength.fair', color: 'var(--af-color-warning)', cssClass: 'strength-fair' },
		3: { label: 'password.strength.good', color: 'var(--af-color-info)', cssClass: 'strength-good' },
		4: { label: 'password.strength.strong', color: 'var(--af-color-success)', cssClass: 'strength-strong' }
	};

	return { score, ...levels[score] };
}
