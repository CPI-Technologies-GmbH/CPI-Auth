import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
	formatDate,
	formatDateTime,
	formatRelativeTime,
	getInitials,
	parseUserAgent,
	getDeviceIcon,
	getLocationString
} from './format';

describe('formatDate', () => {
	it('should format an ISO date string', () => {
		const result = formatDate('2024-01-15T10:30:00Z');
		expect(result).toContain('Jan');
		expect(result).toContain('15');
		expect(result).toContain('2024');
	});

	it('should format different months correctly', () => {
		expect(formatDate('2024-06-01T00:00:00Z')).toContain('Jun');
		expect(formatDate('2024-12-25T00:00:00Z')).toContain('Dec');
		expect(formatDate('2024-03-15T00:00:00Z')).toContain('Mar');
	});

	it('should include the year', () => {
		const result = formatDate('2023-01-01T00:00:00Z');
		expect(result).toContain('2023');
	});
});

describe('formatDateTime', () => {
	it('should include date and time', () => {
		const result = formatDateTime('2024-01-15T10:30:00Z');
		expect(result).toContain('Jan');
		expect(result).toContain('15');
		expect(result).toContain('2024');
		// Should include time
		expect(result).toMatch(/\d{1,2}:\d{2}/);
	});
});

describe('formatRelativeTime', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		vi.setSystemTime(new Date('2024-06-15T12:00:00Z'));
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('should return "Just now" for times less than 60 seconds ago', () => {
		expect(formatRelativeTime('2024-06-15T11:59:30Z')).toBe('Just now');
		expect(formatRelativeTime('2024-06-15T11:59:55Z')).toBe('Just now');
	});

	it('should return minutes for times 1-59 minutes ago', () => {
		expect(formatRelativeTime('2024-06-15T11:55:00Z')).toBe('5m ago');
		expect(formatRelativeTime('2024-06-15T11:30:00Z')).toBe('30m ago');
		expect(formatRelativeTime('2024-06-15T11:59:00Z')).toBe('1m ago');
	});

	it('should return hours for times 1-23 hours ago', () => {
		expect(formatRelativeTime('2024-06-15T11:00:00Z')).toBe('1h ago');
		expect(formatRelativeTime('2024-06-15T09:00:00Z')).toBe('3h ago');
		expect(formatRelativeTime('2024-06-14T13:00:00Z')).toBe('23h ago');
	});

	it('should return days for times 1-6 days ago', () => {
		expect(formatRelativeTime('2024-06-14T12:00:00Z')).toBe('1d ago');
		expect(formatRelativeTime('2024-06-13T12:00:00Z')).toBe('2d ago');
		expect(formatRelativeTime('2024-06-09T12:00:00Z')).toBe('6d ago');
	});

	it('should return formatted date for times 7+ days ago', () => {
		const result = formatRelativeTime('2024-06-01T12:00:00Z');
		expect(result).toContain('Jun');
		expect(result).toContain('2024');
	});
});

describe('getInitials', () => {
	it('should return initials from name', () => {
		expect(getInitials('John Doe')).toBe('JD');
	});

	it('should return first two initials for long names', () => {
		expect(getInitials('John Michael Doe')).toBe('JM');
	});

	it('should return single initial for single name', () => {
		expect(getInitials('John')).toBe('J');
	});

	it('should uppercase initials', () => {
		expect(getInitials('john doe')).toBe('JD');
	});

	it('should fall back to email initial when name is undefined', () => {
		expect(getInitials(undefined, 'john@example.com')).toBe('J');
	});

	it('should uppercase email initial', () => {
		expect(getInitials(undefined, 'alice@example.com')).toBe('A');
	});

	it('should return ? when neither name nor email is provided', () => {
		expect(getInitials()).toBe('?');
		expect(getInitials(undefined, undefined)).toBe('?');
	});

	it('should prefer name over email', () => {
		expect(getInitials('Alice Bob', 'charlie@example.com')).toBe('AB');
	});

	it('should handle empty name string', () => {
		// Empty string is falsy, so it falls through to email check
		expect(getInitials('', 'bob@example.com')).toBe('B');
	});
});

describe('parseUserAgent', () => {
	it('should detect Chrome browser', () => {
		const ua =
			'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36';
		const result = parseUserAgent(ua);
		expect(result.browser).toBe('Chrome');
		expect(result.os).toBe('Windows');
	});

	it('should detect Firefox browser', () => {
		const ua =
			'Mozilla/5.0 (X11; Linux x86_64; rv:121.0) Gecko/20100101 Firefox/121.0';
		const result = parseUserAgent(ua);
		expect(result.browser).toBe('Firefox');
		expect(result.os).toBe('Linux');
	});

	it('should detect Safari browser', () => {
		const ua =
			'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15';
		const result = parseUserAgent(ua);
		expect(result.browser).toBe('Safari');
		expect(result.os).toBe('macOS');
	});

	it('should detect Edge browser', () => {
		const ua =
			'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0';
		const result = parseUserAgent(ua);
		expect(result.browser).toBe('Edge');
		expect(result.os).toBe('Windows');
	});

	it('should detect Opera browser when OPR is present without Chrome', () => {
		// Note: in the source code, Chrome/ is checked before Opera/OPR,
		// so modern Opera UAs (which include Chrome/) match as Chrome.
		// Opera is only detected when Chrome/ is absent.
		const ua = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Opera/106.0.0.0';
		const result = parseUserAgent(ua);
		expect(result.browser).toBe('Opera');
	});

	it('should detect Chrome when Opera UA also includes Chrome/', () => {
		// Modern Opera includes Chrome/ in the UA, which matches Chrome first
		const ua =
			'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0';
		const result = parseUserAgent(ua);
		expect(result.browser).toBe('Chrome');
	});

	it('should detect Windows OS', () => {
		const ua = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0';
		expect(parseUserAgent(ua).os).toBe('Windows');
	});

	it('should detect macOS', () => {
		const ua = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2)';
		expect(parseUserAgent(ua).os).toBe('macOS');
	});

	it('should detect Linux', () => {
		const ua = 'Mozilla/5.0 (X11; Linux x86_64)';
		expect(parseUserAgent(ua).os).toBe('Linux');
	});

	it('should detect Linux for Android UA (Android check comes after Linux)', () => {
		// Note: in the source code, Linux is checked before Android,
		// so Android UAs (which include "Linux") match as Linux.
		const ua = 'Mozilla/5.0 (Linux; Android 14)';
		expect(parseUserAgent(ua).os).toBe('Linux');
	});

	it('should detect Android when Linux is not in the UA', () => {
		const ua = 'Mozilla/5.0 (Android 14; Mobile)';
		expect(parseUserAgent(ua).os).toBe('Android');
	});

	it('should detect macOS for iPhone UA (Mac OS X check comes before iPhone)', () => {
		// Note: in the source code, "Mac OS X" is checked before "iPhone",
		// so iPhone UAs (which include "Mac OS X") match as macOS.
		const ua = 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X)';
		expect(parseUserAgent(ua).os).toBe('macOS');
	});

	it('should detect macOS for iPad UA (Mac OS X check comes before iPad)', () => {
		const ua = 'Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X)';
		expect(parseUserAgent(ua).os).toBe('macOS');
	});

	it('should detect iOS when Mac OS X is not in the UA', () => {
		const ua = 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_2)';
		expect(parseUserAgent(ua).os).toBe('iOS');
	});

	it('should return Unknown for unrecognized user agents', () => {
		const result = parseUserAgent('SomeRandomBot/1.0');
		expect(result.browser).toBe('Unknown Browser');
		expect(result.os).toBe('Unknown OS');
	});
});

describe('getDeviceIcon', () => {
	it('should return "smartphone" for mobile', () => {
		expect(getDeviceIcon('mobile')).toBe('smartphone');
	});

	it('should return "tablet" for tablet', () => {
		expect(getDeviceIcon('tablet')).toBe('tablet');
	});

	it('should return "monitor" for desktop', () => {
		expect(getDeviceIcon('desktop')).toBe('monitor');
	});

	it('should return "monitor" for undefined', () => {
		expect(getDeviceIcon(undefined)).toBe('monitor');
	});

	it('should return "monitor" for unknown device types', () => {
		expect(getDeviceIcon('unknown')).toBe('monitor');
	});

	it('should be case insensitive', () => {
		expect(getDeviceIcon('Mobile')).toBe('smartphone');
		expect(getDeviceIcon('TABLET')).toBe('tablet');
		expect(getDeviceIcon('Desktop')).toBe('monitor');
	});
});

describe('getLocationString', () => {
	it('should return full location with city, region, country', () => {
		const result = getLocationString({
			city: 'San Francisco',
			region: 'California',
			country: 'US'
		});
		expect(result).toBe('San Francisco, California, US');
	});

	it('should return partial location', () => {
		const result = getLocationString({ city: 'Berlin', country: 'DE' });
		expect(result).toBe('Berlin, DE');
	});

	it('should return country only', () => {
		const result = getLocationString({ country: 'US' });
		expect(result).toBe('US');
	});

	it('should return "Unknown location" for undefined', () => {
		expect(getLocationString(undefined)).toBe('Unknown location');
	});

	it('should return "Unknown location" for empty object', () => {
		expect(getLocationString({})).toBe('Unknown location');
	});
});
