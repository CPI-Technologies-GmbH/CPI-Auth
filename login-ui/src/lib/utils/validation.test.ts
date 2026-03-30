import { describe, it, expect } from 'vitest';
import {
	validateEmail,
	validatePassword,
	validatePasswordMatch,
	validateRequired,
	validateOTP,
	getPasswordStrength
} from './validation';

describe('validateEmail', () => {
	it('should accept a valid email', () => {
		expect(validateEmail('user@example.com')).toEqual({ valid: true });
	});

	it('should accept emails with subdomains', () => {
		expect(validateEmail('user@mail.example.com')).toEqual({ valid: true });
	});

	it('should accept emails with plus addressing', () => {
		expect(validateEmail('user+tag@example.com')).toEqual({ valid: true });
	});

	it('should reject empty string', () => {
		const result = validateEmail('');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should reject whitespace-only string', () => {
		const result = validateEmail('   ');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should reject email without @', () => {
		const result = validateEmail('userexample.com');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.email');
	});

	it('should reject email without domain', () => {
		const result = validateEmail('user@');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.email');
	});

	it('should reject email without TLD', () => {
		const result = validateEmail('user@example');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.email');
	});

	it('should reject email with spaces', () => {
		const result = validateEmail('user @example.com');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.email');
	});

	it('should reject email with multiple @ signs', () => {
		const result = validateEmail('user@@example.com');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.email');
	});
});

describe('validatePassword', () => {
	it('should accept a password with 8 or more characters', () => {
		expect(validatePassword('password')).toEqual({ valid: true });
	});

	it('should accept a long password', () => {
		expect(validatePassword('thisIsAVeryLongPasswordThatShouldBeValid')).toEqual({ valid: true });
	});

	it('should reject empty password', () => {
		const result = validatePassword('');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should reject password shorter than 8 characters', () => {
		const result = validatePassword('short');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.password.min');
	});

	it('should reject password with exactly 7 characters', () => {
		const result = validatePassword('1234567');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.password.min');
	});

	it('should accept password with exactly 8 characters', () => {
		expect(validatePassword('12345678')).toEqual({ valid: true });
	});
});

describe('validatePasswordMatch', () => {
	it('should pass when passwords match', () => {
		expect(validatePasswordMatch('Password1!', 'Password1!')).toEqual({ valid: true });
	});

	it('should fail when passwords do not match', () => {
		const result = validatePasswordMatch('Password1!', 'Password2!');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.password.match');
	});

	it('should fail when confirm is empty', () => {
		const result = validatePasswordMatch('Password1!', '');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should match when both are same empty-like strings after non-empty check', () => {
		expect(validatePasswordMatch('abc', 'abc')).toEqual({ valid: true });
	});
});

describe('validateRequired', () => {
	it('should accept non-empty value', () => {
		expect(validateRequired('hello')).toEqual({ valid: true });
	});

	it('should reject empty string', () => {
		const result = validateRequired('');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should reject whitespace-only string', () => {
		const result = validateRequired('   ');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should accept string with leading/trailing whitespace and non-whitespace', () => {
		expect(validateRequired('  hello  ')).toEqual({ valid: true });
	});
});

describe('validateOTP', () => {
	it('should accept valid 6-digit code', () => {
		expect(validateOTP('123456')).toEqual({ valid: true });
	});

	it('should accept 6-digit code with leading zeros', () => {
		expect(validateOTP('000000')).toEqual({ valid: true });
	});

	it('should reject empty code', () => {
		const result = validateOTP('');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.required');
	});

	it('should reject code with fewer than 6 digits', () => {
		const result = validateOTP('12345');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.code');
	});

	it('should reject code with more than 6 digits', () => {
		const result = validateOTP('1234567');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.code');
	});

	it('should reject non-numeric code', () => {
		const result = validateOTP('abcdef');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.code');
	});

	it('should reject code with mixed alphanumeric', () => {
		const result = validateOTP('12ab56');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.code');
	});

	it('should reject code with spaces', () => {
		const result = validateOTP('12 456');
		expect(result.valid).toBe(false);
		expect(result.error).toBe('validation.code');
	});
});

describe('getPasswordStrength', () => {
	it('should return score 0 for empty password', () => {
		const result = getPasswordStrength('');
		expect(result.score).toBe(0);
		expect(result.label).toBe('');
		expect(result.color).toBe('');
		expect(result.cssClass).toBe('');
	});

	it('should return weak (score 1) for short simple password', () => {
		// 8 chars = +1, lowercase only = 0, no digits = 0, no special = 0 => score 1
		const result = getPasswordStrength('abcdefgh');
		expect(result.score).toBe(1);
		expect(result.label).toBe('password.strength.weak');
		expect(result.cssClass).toBe('strength-weak');
	});

	it('should return fair (score 2) for medium password', () => {
		// 8 chars = +1, mixed case = +1, no digits = 0, no special = 0 => score 2
		const result = getPasswordStrength('AbCdEfGh');
		expect(result.score).toBe(2);
		expect(result.label).toBe('password.strength.fair');
		expect(result.cssClass).toBe('strength-fair');
	});

	it('should return good (score 3) for decent password', () => {
		// 8 chars = +1, mixed case = +1, has digit = +1, no special = 0 => score 3
		const result = getPasswordStrength('AbCdEf1h');
		expect(result.score).toBe(3);
		expect(result.label).toBe('password.strength.good');
		expect(result.cssClass).toBe('strength-good');
	});

	it('should return strong (score 4) for complex password', () => {
		// 12+ chars = +2, mixed case = +1, digit = +1, special = +1 => score 5, capped at 4
		const result = getPasswordStrength('MyPassword1!abc');
		expect(result.score).toBe(4);
		expect(result.label).toBe('password.strength.strong');
		expect(result.cssClass).toBe('strength-strong');
	});

	it('should cap score at 4', () => {
		const result = getPasswordStrength('VeryStrongP@ssw0rd123!');
		expect(result.score).toBe(4);
	});

	it('should give higher score to longer passwords', () => {
		// 8 chars lowercase = score 1
		const short = getPasswordStrength('abcdefgh');
		// 12 chars lowercase = score 2 (length >= 8 = +1, length >= 12 = +1)
		const longer = getPasswordStrength('abcdefghijkl');
		expect(longer.score).toBeGreaterThan(short.score);
	});

	it('should give higher score for mixed case', () => {
		const lower = getPasswordStrength('abcdefgh');
		const mixed = getPasswordStrength('AbCdEfGh');
		expect(mixed.score).toBeGreaterThan(lower.score);
	});

	it('should give higher score for digits', () => {
		const noDigits = getPasswordStrength('abcdefgh');
		const withDigits = getPasswordStrength('abcdef1h');
		expect(withDigits.score).toBeGreaterThan(noDigits.score);
	});

	it('should give higher score for special characters', () => {
		const noSpecial = getPasswordStrength('abcdefgh');
		const withSpecial = getPasswordStrength('abcdef!h');
		expect(withSpecial.score).toBeGreaterThan(noSpecial.score);
	});
});
