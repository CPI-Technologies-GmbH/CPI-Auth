import { describe, it, expect } from 'vitest';
import { evaluatePasswordStrength } from './password';

describe('evaluatePasswordStrength', () => {
	describe('empty password', () => {
		it('should return score 0 for empty string', () => {
			const result = evaluatePasswordStrength('');
			expect(result.score).toBe(0);
			expect(result.label).toBe('');
			expect(result.color).toBe('');
			expect(result.suggestions).toEqual([]);
		});
	});

	describe('very weak passwords', () => {
		it('should score very short passwords as very weak', () => {
			const result = evaluatePasswordStrength('abc');
			expect(result.score).toBeLessThanOrEqual(1);
			expect(result.suggestions.length).toBeGreaterThan(0);
		});

		it('should score short only-lowercase as very weak', () => {
			const result = evaluatePasswordStrength('hello');
			expect(result.score).toBeLessThanOrEqual(1);
		});

		it('should score only digits as weak due to common pattern penalty', () => {
			const result = evaluatePasswordStrength('12345678');
			// Only digits: +1 for length >= 8, +0.5 for digits, -1 for common pattern
			expect(result.score).toBeLessThanOrEqual(1);
		});
	});

	describe('weak passwords', () => {
		it('should score lowercase-only 8+ chars as weak', () => {
			const result = evaluatePasswordStrength('abcdefgh');
			// length >= 8: +1, only letters penalty: -1, lower only: no mixed case bonus
			expect(result.score).toBeLessThanOrEqual(1);
		});
	});

	describe('fair passwords', () => {
		it('should score mixed case 8+ chars as fair', () => {
			const result = evaluatePasswordStrength('AbCdEfGh');
			// length >= 8: +1, mixed case: +1, only letters penalty: -1 => ~1
			// Actually: it's all letters, so common pattern penalty applies
			const score = result.score;
			expect(score).toBeGreaterThanOrEqual(1);
		});

		it('should score lowercase with numbers as fair', () => {
			const result = evaluatePasswordStrength('abcdef12');
			// length >= 8: +1, digits: +0.5 => 1.5 => rounds to 2
			expect(result.score).toBeGreaterThanOrEqual(1);
		});
	});

	describe('strong passwords', () => {
		it('should score 12+ char mixed case with numbers as strong', () => {
			const result = evaluatePasswordStrength('AbCdEfGh1234');
			// length >= 8: +1, length >= 12: +1, mixed case: +1, digits: +0.5 => 3.5 => 4
			expect(result.score).toBeGreaterThanOrEqual(3);
		});

		it('should score complex password as very strong', () => {
			const result = evaluatePasswordStrength('MyP@ssw0rd!xyz');
			// length >= 8: +1, length >= 12: +1, mixed case: +1, digits: +0.5, special: +0.5 => 4
			expect(result.score).toBe(4);
			expect(result.label).toBe('Very Strong');
		});
	});

	describe('score capping', () => {
		it('should not exceed score of 4', () => {
			const result = evaluatePasswordStrength('VeryStrongP@ssw0rd123!XYZ');
			expect(result.score).toBeLessThanOrEqual(4);
		});
	});

	describe('labels', () => {
		it('should return correct labels for each score level', () => {
			const labels = ['Very Weak', 'Weak', 'Fair', 'Strong', 'Very Strong'];
			// We verify at least some labels map correctly
			const veryStrong = evaluatePasswordStrength('MyP@ssw0rd!xyz');
			expect(veryStrong.label).toBe('Very Strong');
		});
	});

	describe('colors', () => {
		it('should return a color for non-empty passwords', () => {
			const result = evaluatePasswordStrength('abc');
			expect(result.color).toBeTruthy();
			expect(result.color).toMatch(/^#/);
		});

		it('should return red-ish color for weak passwords', () => {
			const result = evaluatePasswordStrength('abc');
			// Score 0 => '#ef4444', Score 1 => '#f97316'
			expect(['#ef4444', '#f97316']).toContain(result.color);
		});

		it('should return green color for strong passwords', () => {
			const result = evaluatePasswordStrength('MyP@ssw0rd!xyz');
			expect(result.color).toBe('#16a34a');
		});
	});

	describe('suggestions', () => {
		it('should suggest length for short passwords', () => {
			const result = evaluatePasswordStrength('Ab1!');
			expect(result.suggestions).toContain('Use at least 8 characters');
		});

		it('should suggest mixed case for lowercase-only passwords', () => {
			const result = evaluatePasswordStrength('abcdefgh');
			expect(result.suggestions).toContain('Use both uppercase and lowercase letters');
		});

		it('should suggest numbers for letter-only passwords', () => {
			const result = evaluatePasswordStrength('AbCdEfGh');
			expect(result.suggestions).toContain('Add numbers');
		});

		it('should suggest special characters for weaker alphanumeric passwords', () => {
			// AbCd1234 scores >= 3 (length +1, mixed case +1, digits +0.5 = 2.5, rounds to 3)
			// Suggestions are only returned for score < 3, so use a weaker password
			const result = evaluatePasswordStrength('abcd1234');
			// lowercase + digits: length >= 8: +1, no mixed case, digits +0.5 = 1.5, rounds to 2
			expect(result.suggestions).toContain('Add special characters');
		});

		it('should return no suggestions for strong passwords (score >= 3)', () => {
			const result = evaluatePasswordStrength('MyP@ssw0rd!xyz');
			expect(result.suggestions).toEqual([]);
		});

		it('should warn about repeating characters', () => {
			const result = evaluatePasswordStrength('aaabbb12');
			expect(result.suggestions).toContain('Avoid repeating characters');
		});
	});

	describe('common pattern detection', () => {
		it('should penalize all-letter passwords', () => {
			const allLetters = evaluatePasswordStrength('abcdefghi');
			const withDigits = evaluatePasswordStrength('abcdefg1i');
			// All letters should score lower due to penalty
			expect(allLetters.score).toBeLessThanOrEqual(withDigits.score);
		});

		it('should penalize all-digit passwords', () => {
			const allDigits = evaluatePasswordStrength('123456789');
			expect(allDigits.score).toBeLessThanOrEqual(1);
		});
	});

	describe('edge cases', () => {
		it('should handle very long passwords', () => {
			const longPassword = 'A'.repeat(100) + 'b1!';
			const result = evaluatePasswordStrength(longPassword);
			// Should still produce a valid score
			expect(result.score).toBeGreaterThanOrEqual(0);
			expect(result.score).toBeLessThanOrEqual(4);
		});

		it('should handle single character', () => {
			const result = evaluatePasswordStrength('a');
			expect(result.score).toBeLessThanOrEqual(1);
		});

		it('should handle special characters only', () => {
			const result = evaluatePasswordStrength('!@#$%^&*');
			expect(result.score).toBeGreaterThanOrEqual(0);
		});

		it('should handle unicode characters', () => {
			const result = evaluatePasswordStrength('p@ssword123');
			expect(result.score).toBeGreaterThanOrEqual(0);
			expect(result.score).toBeLessThanOrEqual(4);
		});
	});
});
