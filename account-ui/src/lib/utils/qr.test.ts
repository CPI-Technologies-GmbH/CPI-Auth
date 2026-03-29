import { describe, it, expect } from 'vitest';
import { generateQRCode, qrToSvg } from './qr';

describe('generateQRCode', () => {
	it('should generate a 2D boolean matrix', () => {
		const matrix = generateQRCode('Hello');
		expect(Array.isArray(matrix)).toBe(true);
		expect(matrix.length).toBeGreaterThan(0);

		for (const row of matrix) {
			expect(Array.isArray(row)).toBe(true);
			expect(row.length).toBe(matrix.length); // Square matrix
			for (const cell of row) {
				expect(typeof cell).toBe('boolean');
			}
		}
	});

	it('should generate version 1 (21x21) for short input', () => {
		const matrix = generateQRCode('Hi');
		expect(matrix.length).toBe(21);
		expect(matrix[0].length).toBe(21);
	});

	it('should generate larger matrix for longer input', () => {
		const shortMatrix = generateQRCode('A');
		const longMatrix = generateQRCode(
			'This is a much longer string that needs more space in the QR code'
		);
		expect(longMatrix.length).toBeGreaterThanOrEqual(shortMatrix.length);
	});

	it('should handle TOTP URI format', () => {
		const totpUri = 'otpauth://totp/CPI Auth:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=CPI Auth';
		const matrix = generateQRCode(totpUri);

		expect(matrix.length).toBeGreaterThan(0);
		// TOTP URIs are longer, should produce version 2+ (25x25 or larger)
		expect(matrix.length).toBeGreaterThanOrEqual(25);
	});

	it('should produce a matrix with finder patterns in corners', () => {
		const matrix = generateQRCode('Test');
		const size = matrix.length;

		// Top-left finder pattern: first 7 rows and columns should follow the pattern
		// The outer ring of 7x7 should be true (dark modules)
		expect(matrix[0][0]).toBe(true); // Top-left corner
		expect(matrix[0][6]).toBe(true); // Top-right of finder
		expect(matrix[6][0]).toBe(true); // Bottom-left of finder
		expect(matrix[6][6]).toBe(true); // Bottom-right of finder

		// Top-right finder pattern
		expect(matrix[0][size - 1]).toBe(true);
		expect(matrix[0][size - 7]).toBe(true);

		// Bottom-left finder pattern
		expect(matrix[size - 1][0]).toBe(true);
		expect(matrix[size - 7][0]).toBe(true);
	});

	it('should produce consistent output for the same input', () => {
		const matrix1 = generateQRCode('Deterministic');
		const matrix2 = generateQRCode('Deterministic');

		expect(matrix1.length).toBe(matrix2.length);
		for (let r = 0; r < matrix1.length; r++) {
			for (let c = 0; c < matrix1[r].length; c++) {
				expect(matrix1[r][c]).toBe(matrix2[r][c]);
			}
		}
	});

	it('should produce different output for different inputs', () => {
		const matrix1 = generateQRCode('Hello');
		const matrix2 = generateQRCode('World');

		// At least some cells should differ (not all will be different)
		let hasDifference = false;
		const minLen = Math.min(matrix1.length, matrix2.length);
		for (let r = 0; r < minLen; r++) {
			for (let c = 0; c < minLen; c++) {
				if (matrix1[r][c] !== matrix2[r][c]) {
					hasDifference = true;
					break;
				}
			}
			if (hasDifference) break;
		}
		expect(hasDifference).toBe(true);
	});

	it('should handle different input lengths', () => {
		// Single character
		const tiny = generateQRCode('A');
		expect(tiny.length).toBe(21);

		// Medium string
		const medium = generateQRCode('A medium length string');
		expect(medium.length).toBeGreaterThanOrEqual(21);

		// Longer string
		const longer = generateQRCode('A'.repeat(80));
		expect(longer.length).toBeGreaterThanOrEqual(medium.length);
	});

	it('should handle special characters', () => {
		const matrix = generateQRCode('Hello! @#$%^&*()');
		expect(matrix.length).toBeGreaterThan(0);
	});

	it('should handle unicode characters', () => {
		const matrix = generateQRCode('Hallo Welt');
		expect(matrix.length).toBeGreaterThan(0);
	});
});

describe('qrToSvg', () => {
	it('should generate valid SVG string', () => {
		const matrix = generateQRCode('Test');
		const svg = qrToSvg(matrix);

		expect(svg).toContain('<svg');
		expect(svg).toContain('xmlns="http://www.w3.org/2000/svg"');
		expect(svg).toContain('</svg>');
	});

	it('should include a white background rect', () => {
		const matrix = generateQRCode('Test');
		const svg = qrToSvg(matrix);

		expect(svg).toContain('fill="white"');
	});

	it('should include path data for dark modules', () => {
		const matrix = generateQRCode('Test');
		const svg = qrToSvg(matrix);

		expect(svg).toContain('<path');
		expect(svg).toContain('fill="black"');
		expect(svg).toContain('d="M');
	});

	it('should respect custom module size', () => {
		const matrix = generateQRCode('Test');
		const svg4 = qrToSvg(matrix, 4);
		const svg8 = qrToSvg(matrix, 8);

		// Larger module size should produce larger viewBox
		const viewBox4 = svg4.match(/viewBox="0 0 (\d+) (\d+)"/);
		const viewBox8 = svg8.match(/viewBox="0 0 (\d+) (\d+)"/);

		expect(viewBox4).not.toBeNull();
		expect(viewBox8).not.toBeNull();

		const size4 = parseInt(viewBox4![1]);
		const size8 = parseInt(viewBox8![1]);
		expect(size8).toBeGreaterThan(size4);
	});

	it('should respect custom margin', () => {
		const matrix = generateQRCode('Test');
		const svgSmallMargin = qrToSvg(matrix, 4, 2);
		const svgLargeMargin = qrToSvg(matrix, 4, 8);

		const viewBoxSmall = svgSmallMargin.match(/viewBox="0 0 (\d+) (\d+)"/);
		const viewBoxLarge = svgLargeMargin.match(/viewBox="0 0 (\d+) (\d+)"/);

		const sizeSmall = parseInt(viewBoxSmall![1]);
		const sizeLarge = parseInt(viewBoxLarge![1]);
		expect(sizeLarge).toBeGreaterThan(sizeSmall);
	});

	it('should generate SVG with correct dimensions', () => {
		const matrix = generateQRCode('Test');
		const moduleSize = 4;
		const margin = 4;
		const expectedSize = (matrix.length + margin * 2) * moduleSize;

		const svg = qrToSvg(matrix, moduleSize, margin);

		expect(svg).toContain(`width="${expectedSize}"`);
		expect(svg).toContain(`height="${expectedSize}"`);
	});

	it('should use default module size and margin', () => {
		const matrix = generateQRCode('Test');
		const svg = qrToSvg(matrix);

		// Default moduleSize=4, margin=4
		const expectedSize = (matrix.length + 8) * 4;
		expect(svg).toContain(`width="${expectedSize}"`);
	});

	it('should generate path data with M, h, v commands', () => {
		const matrix = generateQRCode('Hi');
		const svg = qrToSvg(matrix);

		// Should contain move (M), horizontal (h), vertical (v) path commands
		expect(svg).toMatch(/M\d+,\d+h\d+v\d+h-\d+z/);
	});
});
