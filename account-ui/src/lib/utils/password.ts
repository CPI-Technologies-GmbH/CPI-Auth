export interface PasswordStrength {
	score: number; // 0-4
	label: string;
	color: string;
	suggestions: string[];
}

export function evaluatePasswordStrength(password: string): PasswordStrength {
	if (!password) {
		return { score: 0, label: '', color: '', suggestions: [] };
	}

	let score = 0;
	const suggestions: string[] = [];

	// Length checks
	if (password.length >= 8) score++;
	if (password.length >= 12) score++;
	if (password.length < 8) suggestions.push('Use at least 8 characters');

	// Character variety
	if (/[a-z]/.test(password) && /[A-Z]/.test(password)) {
		score++;
	} else {
		suggestions.push('Use both uppercase and lowercase letters');
	}

	if (/\d/.test(password)) {
		score += 0.5;
	} else {
		suggestions.push('Add numbers');
	}

	if (/[^a-zA-Z0-9]/.test(password)) {
		score += 0.5;
	} else {
		suggestions.push('Add special characters');
	}

	// Common patterns penalty
	if (/^[a-zA-Z]+$/.test(password) || /^\d+$/.test(password)) {
		score = Math.max(0, score - 1);
		suggestions.push('Avoid using only letters or only numbers');
	}

	if (/(.)\1{2,}/.test(password)) {
		score = Math.max(0, score - 0.5);
		suggestions.push('Avoid repeating characters');
	}

	// Normalize score to 0-4
	const normalizedScore = Math.min(4, Math.round(score));

	const labels = ['Very Weak', 'Weak', 'Fair', 'Strong', 'Very Strong'];
	const colors = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#16a34a'];

	return {
		score: normalizedScore,
		label: labels[normalizedScore],
		color: colors[normalizedScore],
		suggestions: normalizedScore < 3 ? suggestions : []
	};
}
