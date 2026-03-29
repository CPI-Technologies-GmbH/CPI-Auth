// Simple QR code generator using SVG
// Implements a basic QR encoding algorithm sufficient for TOTP URIs

type Module = boolean;

// Simple polynomial multiplication in GF(256)
const GF256_EXP = new Uint8Array(256);
const GF256_LOG = new Uint8Array(256);

// Initialize Galois Field tables
(function initGF() {
	let x = 1;
	for (let i = 0; i < 255; i++) {
		GF256_EXP[i] = x;
		GF256_LOG[x] = i;
		x = x << 1;
		if (x & 0x100) x ^= 0x11d;
	}
	GF256_EXP[255] = GF256_EXP[0];
})();

function gfMul(a: number, b: number): number {
	if (a === 0 || b === 0) return 0;
	return GF256_EXP[(GF256_LOG[a] + GF256_LOG[b]) % 255];
}

function getErrorCorrectionCodewords(data: number[], ecLength: number): number[] {
	// Generate EC polynomial
	const genPoly: number[] = new Array(ecLength + 1).fill(0);
	genPoly[0] = 1;
	for (let i = 0; i < ecLength; i++) {
		const newPoly = new Array(ecLength + 1).fill(0);
		for (let j = 0; j <= i + 1; j++) {
			if (j > 0) newPoly[j] ^= genPoly[j - 1];
			newPoly[j] ^= gfMul(genPoly[j], GF256_EXP[i]);
		}
		for (let j = 0; j <= ecLength; j++) genPoly[j] = newPoly[j];
	}

	// Polynomial division
	const msgPoly = new Array(data.length + ecLength).fill(0);
	for (let i = 0; i < data.length; i++) msgPoly[i] = data[i];

	for (let i = 0; i < data.length; i++) {
		if (msgPoly[i] === 0) continue;
		for (let j = 0; j <= ecLength; j++) {
			msgPoly[i + j] ^= gfMul(genPoly[j], msgPoly[i]);
		}
	}

	return msgPoly.slice(data.length);
}

function encodeByteMode(text: string): number[] {
	const bytes = new TextEncoder().encode(text);
	const bits: number[] = [];

	// Mode indicator for byte mode: 0100
	bits.push(0, 1, 0, 0);

	// Character count (8 bits for versions 1-9)
	const len = bytes.length;
	for (let i = 7; i >= 0; i--) bits.push((len >> i) & 1);

	// Data
	for (const b of bytes) {
		for (let i = 7; i >= 0; i--) bits.push((b >> i) & 1);
	}

	return bits;
}

function bitsToCodewords(bits: number[], totalCodewords: number): number[] {
	// Terminator
	const maxBits = totalCodewords * 8;
	for (let i = 0; i < 4 && bits.length < maxBits; i++) bits.push(0);

	// Pad to byte boundary
	while (bits.length % 8 !== 0 && bits.length < maxBits) bits.push(0);

	// Convert to codewords
	const codewords: number[] = [];
	for (let i = 0; i < bits.length; i += 8) {
		let cw = 0;
		for (let j = 0; j < 8; j++) cw = (cw << 1) | (bits[i + j] || 0);
		codewords.push(cw);
	}

	// Pad codewords
	const padBytes = [0xec, 0x11];
	let padIdx = 0;
	while (codewords.length < totalCodewords) {
		codewords.push(padBytes[padIdx % 2]);
		padIdx++;
	}

	return codewords;
}

// QR Version configurations for byte mode (EC Level M)
interface VersionConfig {
	version: number;
	size: number;
	totalCodewords: number;
	dataCodewords: number;
	ecCodewordsPerBlock: number;
	blocks: number;
	maxBytes: number;
}

const VERSIONS: VersionConfig[] = [
	{ version: 1, size: 21, totalCodewords: 26, dataCodewords: 16, ecCodewordsPerBlock: 10, blocks: 1, maxBytes: 14 },
	{ version: 2, size: 25, totalCodewords: 44, dataCodewords: 28, ecCodewordsPerBlock: 16, blocks: 1, maxBytes: 26 },
	{ version: 3, size: 29, totalCodewords: 70, dataCodewords: 44, ecCodewordsPerBlock: 26, blocks: 1, maxBytes: 42 },
	{ version: 4, size: 33, totalCodewords: 100, dataCodewords: 64, ecCodewordsPerBlock: 18, blocks: 2, maxBytes: 62 },
	{ version: 5, size: 37, totalCodewords: 134, dataCodewords: 86, ecCodewordsPerBlock: 24, blocks: 2, maxBytes: 84 },
	{ version: 6, size: 41, totalCodewords: 172, dataCodewords: 108, ecCodewordsPerBlock: 16, blocks: 4, maxBytes: 106 },
];

function selectVersion(dataLength: number): VersionConfig {
	for (const v of VERSIONS) {
		if (dataLength <= v.maxBytes) return v;
	}
	return VERSIONS[VERSIONS.length - 1];
}

function createMatrix(size: number): Module[][] {
	return Array.from({ length: size }, () => Array(size).fill(false));
}

function addFinderPattern(matrix: Module[][], reserved: Module[][], row: number, col: number) {
	const size = matrix.length;
	for (let r = -1; r <= 7; r++) {
		for (let c = -1; c <= 7; c++) {
			const mr = row + r;
			const mc = col + c;
			if (mr < 0 || mr >= size || mc < 0 || mc >= size) continue;
			reserved[mr][mc] = true;
			if (r >= 0 && r <= 6 && c >= 0 && c <= 6) {
				const isOuter = r === 0 || r === 6 || c === 0 || c === 6;
				const isInner = r >= 2 && r <= 4 && c >= 2 && c <= 4;
				matrix[mr][mc] = isOuter || isInner;
			} else {
				matrix[mr][mc] = false;
			}
		}
	}
}

const ALIGNMENT_POSITIONS: Record<number, number[]> = {
	2: [6, 18],
	3: [6, 22],
	4: [6, 26],
	5: [6, 30],
	6: [6, 34]
};

function addAlignmentPatterns(matrix: Module[][], reserved: Module[][], version: number) {
	const positions = ALIGNMENT_POSITIONS[version];
	if (!positions) return;

	for (const row of positions) {
		for (const col of positions) {
			// Skip if overlaps with finder patterns
			if (reserved[row]?.[col]) continue;
			for (let r = -2; r <= 2; r++) {
				for (let c = -2; c <= 2; c++) {
					const mr = row + r;
					const mc = col + c;
					if (mr < 0 || mr >= matrix.length || mc < 0 || mc >= matrix.length) continue;
					reserved[mr][mc] = true;
					const isOuter = Math.abs(r) === 2 || Math.abs(c) === 2;
					const isCenter = r === 0 && c === 0;
					matrix[mr][mc] = isOuter || isCenter;
				}
			}
		}
	}
}

function addTimingPatterns(matrix: Module[][], reserved: Module[][]) {
	const size = matrix.length;
	for (let i = 8; i < size - 8; i++) {
		if (!reserved[6][i]) {
			matrix[6][i] = i % 2 === 0;
			reserved[6][i] = true;
		}
		if (!reserved[i][6]) {
			matrix[i][6] = i % 2 === 0;
			reserved[i][6] = true;
		}
	}
}

function reserveFormatInfo(reserved: Module[][], size: number) {
	// Around top-left finder
	for (let i = 0; i <= 8; i++) {
		if (i < size) reserved[8][i] = true;
		if (i < size) reserved[i][8] = true;
	}
	// Around top-right finder
	for (let i = 0; i <= 7; i++) {
		reserved[8][size - 1 - i] = true;
	}
	// Around bottom-left finder
	for (let i = 0; i <= 7; i++) {
		reserved[size - 1 - i][8] = true;
	}
	// Dark module
	reserved[size - 8][8] = true;
}

function reserveVersionInfo(reserved: Module[][], version: number, size: number) {
	if (version < 7) return;
	for (let i = 0; i < 6; i++) {
		for (let j = 0; j < 3; j++) {
			reserved[i][size - 11 + j] = true;
			reserved[size - 11 + j][i] = true;
		}
	}
}

function placeDataBits(matrix: Module[][], reserved: Module[][], dataBits: number[]) {
	const size = matrix.length;
	let bitIdx = 0;
	let upward = true;

	for (let right = size - 1; right >= 1; right -= 2) {
		if (right === 6) right = 5; // Skip timing column
		const rows = upward
			? Array.from({ length: size }, (_, i) => size - 1 - i)
			: Array.from({ length: size }, (_, i) => i);

		for (const row of rows) {
			for (const col of [right, right - 1]) {
				if (col < 0 || col >= size) continue;
				if (reserved[row][col]) continue;
				matrix[row][col] = bitIdx < dataBits.length ? dataBits[bitIdx] === 1 : false;
				bitIdx++;
			}
		}
		upward = !upward;
	}
}

// Format info for mask pattern 0, EC level M
const FORMAT_BITS_M_MASK0 = [1, 0, 1, 0, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0];

function applyFormatInfo(matrix: Module[][], size: number) {
	const bits = FORMAT_BITS_M_MASK0;

	// Top-left horizontal
	for (let i = 0; i <= 7; i++) {
		const idx = i < 6 ? i : i === 6 ? 7 : 8;
		matrix[8][idx < 6 ? idx : idx === 7 ? 7 : 8] = bits[i] === 1;
	}
	// Fix: proper placement around top-left
	const hPositions = [0, 1, 2, 3, 4, 5, 7, 8];
	const vPositions = [0, 1, 2, 3, 4, 5, 7, 8];
	for (let i = 0; i < 8; i++) {
		matrix[8][hPositions[i]] = bits[i] === 1;
	}
	for (let i = 0; i < 7; i++) {
		matrix[vPositions[6 - i]][8] = bits[8 + i] === 1;
	}

	// Top-right
	for (let i = 0; i < 8; i++) {
		matrix[8][size - 1 - i] = bits[14 - i] === 1;
	}
	// Bottom-left
	for (let i = 0; i < 7; i++) {
		matrix[size - 1 - i][8] = bits[i] === 1;
	}
	// Dark module
	matrix[size - 8][8] = true;
}

function applyMask0(matrix: Module[][], reserved: Module[][]) {
	const size = matrix.length;
	for (let r = 0; r < size; r++) {
		for (let c = 0; c < size; c++) {
			if (!reserved[r][c] && (r + c) % 2 === 0) {
				matrix[r][c] = !matrix[r][c];
			}
		}
	}
}

export function generateQRCode(text: string): Module[][] {
	const textBytes = new TextEncoder().encode(text);
	const config = selectVersion(textBytes.length);
	const size = config.size;

	// Encode data
	const dataBits = encodeByteMode(text);

	// Adjust character count indicator length for version >= 10
	const dataCodewords = bitsToCodewords(dataBits, config.dataCodewords);

	// Generate EC codewords
	const blockSize = Math.floor(config.dataCodewords / config.blocks);
	const remainder = config.dataCodewords % config.blocks;

	const dataBlocks: number[][] = [];
	const ecBlocks: number[][] = [];
	let offset = 0;

	for (let b = 0; b < config.blocks; b++) {
		const thisBlockSize = blockSize + (b >= config.blocks - remainder ? 1 : 0);
		const block = dataCodewords.slice(offset, offset + thisBlockSize);
		dataBlocks.push(block);
		ecBlocks.push(getErrorCorrectionCodewords(block, config.ecCodewordsPerBlock));
		offset += thisBlockSize;
	}

	// Interleave data codewords
	const interleaved: number[] = [];
	const maxDataBlock = Math.max(...dataBlocks.map((b) => b.length));
	for (let i = 0; i < maxDataBlock; i++) {
		for (const block of dataBlocks) {
			if (i < block.length) interleaved.push(block[i]);
		}
	}
	for (let i = 0; i < config.ecCodewordsPerBlock; i++) {
		for (const block of ecBlocks) {
			if (i < block.length) interleaved.push(block[i]);
		}
	}

	// Convert to bits
	const allBits: number[] = [];
	for (const cw of interleaved) {
		for (let i = 7; i >= 0; i--) allBits.push((cw >> i) & 1);
	}
	// Remainder bits for certain versions
	const remainderBits: Record<number, number> = { 2: 7, 3: 7, 4: 7, 5: 7, 6: 7 };
	const rb = remainderBits[config.version] || 0;
	for (let i = 0; i < rb; i++) allBits.push(0);

	// Build matrix
	const matrix = createMatrix(size);
	const reserved = createMatrix(size);

	// Add finder patterns
	addFinderPattern(matrix, reserved, 0, 0);
	addFinderPattern(matrix, reserved, 0, size - 7);
	addFinderPattern(matrix, reserved, size - 7, 0);

	// Add alignment patterns
	addAlignmentPatterns(matrix, reserved, config.version);

	// Add timing patterns
	addTimingPatterns(matrix, reserved);

	// Reserve format and version info areas
	reserveFormatInfo(reserved, size);
	reserveVersionInfo(reserved, config.version, size);

	// Place data
	placeDataBits(matrix, reserved, allBits);

	// Apply mask pattern 0 (checkerboard)
	applyMask0(matrix, reserved);

	// Write format info
	applyFormatInfo(matrix, size);

	return matrix;
}

export function qrToSvg(matrix: Module[][], moduleSize: number = 4, margin: number = 4): string {
	const size = matrix.length;
	const svgSize = (size + margin * 2) * moduleSize;

	let pathData = '';
	for (let row = 0; row < size; row++) {
		for (let col = 0; col < size; col++) {
			if (matrix[row][col]) {
				const x = (col + margin) * moduleSize;
				const y = (row + margin) * moduleSize;
				pathData += `M${x},${y}h${moduleSize}v${moduleSize}h-${moduleSize}z`;
			}
		}
	}

	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${svgSize} ${svgSize}" width="${svgSize}" height="${svgSize}">
<rect width="${svgSize}" height="${svgSize}" fill="white"/>
<path d="${pathData}" fill="black"/>
</svg>`;
}
