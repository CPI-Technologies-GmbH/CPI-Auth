export function formatDate(dateString: string): string {
	const date = new Date(dateString);
	return new Intl.DateTimeFormat('en-US', {
		year: 'numeric',
		month: 'short',
		day: 'numeric'
	}).format(date);
}

export function formatDateTime(dateString: string): string {
	const date = new Date(dateString);
	return new Intl.DateTimeFormat('en-US', {
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: '2-digit',
		minute: '2-digit'
	}).format(date);
}

export function formatRelativeTime(dateString: string): string {
	const date = new Date(dateString);
	const now = new Date();
	const diffMs = now.getTime() - date.getTime();
	const diffSeconds = Math.floor(diffMs / 1000);
	const diffMinutes = Math.floor(diffSeconds / 60);
	const diffHours = Math.floor(diffMinutes / 60);
	const diffDays = Math.floor(diffHours / 24);

	if (diffSeconds < 60) return 'Just now';
	if (diffMinutes < 60) return `${diffMinutes}m ago`;
	if (diffHours < 24) return `${diffHours}h ago`;
	if (diffDays < 7) return `${diffDays}d ago`;
	return formatDate(dateString);
}

export function getInitials(name?: string, email?: string): string {
	if (name) {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.toUpperCase()
			.slice(0, 2);
	}
	if (email) {
		return email[0].toUpperCase();
	}
	return '?';
}

export function parseUserAgent(ua: string): { browser: string; os: string } {
	let browser = 'Unknown Browser';
	let os = 'Unknown OS';

	// Browser detection
	if (ua.includes('Firefox/')) browser = 'Firefox';
	else if (ua.includes('Edg/')) browser = 'Edge';
	else if (ua.includes('Chrome/')) browser = 'Chrome';
	else if (ua.includes('Safari/') && !ua.includes('Chrome')) browser = 'Safari';
	else if (ua.includes('Opera') || ua.includes('OPR')) browser = 'Opera';

	// OS detection
	if (ua.includes('Windows NT')) os = 'Windows';
	else if (ua.includes('Mac OS X')) os = 'macOS';
	else if (ua.includes('Linux')) os = 'Linux';
	else if (ua.includes('Android')) os = 'Android';
	else if (ua.includes('iPhone') || ua.includes('iPad')) os = 'iOS';

	return { browser, os };
}

export function getDeviceIcon(deviceType?: string): string {
	switch (deviceType?.toLowerCase()) {
		case 'mobile':
			return 'smartphone';
		case 'tablet':
			return 'tablet';
		case 'desktop':
		default:
			return 'monitor';
	}
}

export function getLocationString(location?: {
	city?: string;
	region?: string;
	country?: string;
}): string {
	if (!location) return 'Unknown location';
	const parts = [location.city, location.region, location.country].filter(Boolean);
	return parts.length > 0 ? parts.join(', ') : 'Unknown location';
}
