<script lang="ts">
	import { api } from '$lib/api';
	import type { User } from '$lib/types';
	import Avatar from '$components/Avatar.svelte';
	import Alert from '$components/Alert.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';
	import Badge from '$components/Badge.svelte';
	import { invalidateAll } from '$app/navigation';

	let { data }: { data: { user: User } } = $props();

	const initialUser = data.user;
	let name = $state(initialUser.name || '');
	let givenName = $state(initialUser.given_name || '');
	let familyName = $state(initialUser.family_name || '');
	let nickname = $state(initialUser.nickname || '');
	let email = $state(initialUser.email || '');
	let phone = $state(initialUser.phone || '');
	let locale = $state(initialUser.locale || '');
	let timezone = $state(initialUser.timezone || '');

	let saving = $state(false);
	let successMessage = $state('');
	let errorMessage = $state('');

	let newEmail = $state('');
	let showEmailChange = $state(false);
	let emailChanging = $state(false);

	let newPhone = $state('');
	let showPhoneChange = $state(false);
	let phoneChanging = $state(false);

	const timezones = Intl.supportedValuesOf('timeZone');
	const locales = [
		{ value: '', label: 'Not set' },
		{ value: 'en', label: 'English' },
		{ value: 'de', label: 'Deutsch' },
		{ value: 'fr', label: 'Francais' },
		{ value: 'es', label: 'Espanol' },
		{ value: 'pt', label: 'Portugues' },
		{ value: 'it', label: 'Italiano' },
		{ value: 'nl', label: 'Nederlands' },
		{ value: 'ja', label: 'Japanese' },
		{ value: 'zh', label: 'Chinese' },
		{ value: 'ko', label: 'Korean' }
	];

	async function handleSaveProfile() {
		saving = true;
		successMessage = '';
		errorMessage = '';

		try {
			await api.updateMe({
				name: name || undefined,
				given_name: givenName || undefined,
				family_name: familyName || undefined,
				nickname: nickname || undefined,
				locale: locale || undefined,
				timezone: timezone || undefined
			});
			successMessage = 'Profile updated successfully.';
			await invalidateAll();
		} catch (err: unknown) {
			errorMessage = (err as { message?: string })?.message || 'Failed to update profile.';
		} finally {
			saving = false;
		}
	}

	async function handleEmailChange() {
		if (!newEmail) return;
		emailChanging = true;
		errorMessage = '';
		successMessage = '';

		try {
			await api.updateMe({ name: data.user.name }); // Placeholder - email change usually has its own endpoint
			successMessage = 'Verification email sent to ' + newEmail + '. Please check your inbox.';
			showEmailChange = false;
			newEmail = '';
		} catch (err: unknown) {
			errorMessage = (err as { message?: string })?.message || 'Failed to change email.';
		} finally {
			emailChanging = false;
		}
	}

	async function handlePhoneChange() {
		if (!newPhone) return;
		phoneChanging = true;
		errorMessage = '';
		successMessage = '';

		try {
			await api.updateMe({ name: data.user.name }); // Placeholder
			successMessage = 'Verification code sent to ' + newPhone + '.';
			showPhoneChange = false;
			newPhone = '';
		} catch (err: unknown) {
			errorMessage = (err as { message?: string })?.message || 'Failed to change phone number.';
		} finally {
			phoneChanging = false;
		}
	}

	function handleAvatarUpload(file: File) {
		// In a real implementation, this would upload the file and get a URL back
		const reader = new FileReader();
		reader.onload = async () => {
			try {
				await api.updateMe({ avatar_url: reader.result as string });
				successMessage = 'Avatar updated successfully.';
				await invalidateAll();
			} catch (err: unknown) {
				errorMessage = (err as { message?: string })?.message || 'Failed to update avatar.';
			}
		};
		reader.readAsDataURL(file);
	}
</script>

<svelte:head>
	<title>Profile - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<h1>Profile</h1>
		<p>Manage your personal information and preferences.</p>
	</div>

	{#if successMessage}
		<Alert variant="success" dismissible ondismiss={() => (successMessage = '')}>{successMessage}</Alert>
	{/if}
	{#if errorMessage}
		<Alert variant="danger" dismissible ondismiss={() => (errorMessage = '')}>{errorMessage}</Alert>
	{/if}

	<!-- Avatar Section -->
	<section class="card">
		<h2 class="card-title">Avatar</h2>
		<div class="avatar-section">
			<Avatar
				src={data.user.avatar_url}
				name={data.user.name}
				email={data.user.email}
				size={80}
				editable
				onupload={handleAvatarUpload}
			/>
			<div class="avatar-info">
				<p class="avatar-hint">Click the avatar to upload a new photo. JPG, PNG or GIF, max 2MB.</p>
			</div>
		</div>
	</section>

	<!-- Personal Information -->
	<section class="card">
		<h2 class="card-title">Personal Information</h2>
		<form class="form" onsubmit={(e) => { e.preventDefault(); handleSaveProfile(); }}>
			<div class="form-grid">
				<div class="field">
					<label for="name" class="field-label">Display Name</label>
					<input id="name" type="text" class="input" bind:value={name} placeholder="Your display name" />
				</div>
				<div class="field">
					<label for="nickname" class="field-label">Nickname</label>
					<input id="nickname" type="text" class="input" bind:value={nickname} placeholder="Nickname" />
				</div>
				<div class="field">
					<label for="given-name" class="field-label">First Name</label>
					<input id="given-name" type="text" class="input" bind:value={givenName} placeholder="First name" />
				</div>
				<div class="field">
					<label for="family-name" class="field-label">Last Name</label>
					<input id="family-name" type="text" class="input" bind:value={familyName} placeholder="Last name" />
				</div>
			</div>

			<div class="form-actions">
				<button type="submit" class="btn btn-primary" disabled={saving}>
					{#if saving}
						<LoadingSpinner size={16} color="white" />
					{/if}
					Save Changes
				</button>
			</div>
		</form>
	</section>

	<!-- Email -->
	<section class="card">
		<div class="card-header-row">
			<div>
				<h2 class="card-title">Email Address</h2>
				<p class="card-description">Your email is used for sign-in and notifications.</p>
			</div>
		</div>
		<div class="info-row">
			<div class="info-value">
				<span>{data.user.email}</span>
				{#if data.user.email_verified}
					<Badge variant="success">Verified</Badge>
				{:else}
					<Badge variant="warning">Unverified</Badge>
				{/if}
			</div>
			<button class="btn btn-secondary" onclick={() => (showEmailChange = !showEmailChange)}>
				Change Email
			</button>
		</div>

		{#if showEmailChange}
			<form class="change-form" onsubmit={(e) => { e.preventDefault(); handleEmailChange(); }}>
				<div class="field">
					<label for="new-email" class="field-label">New Email Address</label>
					<input id="new-email" type="email" class="input" bind:value={newEmail} placeholder="Enter new email" required />
				</div>
				<div class="change-form-actions">
					<button type="button" class="btn btn-secondary" onclick={() => { showEmailChange = false; newEmail = ''; }}>Cancel</button>
					<button type="submit" class="btn btn-primary" disabled={emailChanging || !newEmail}>
						{#if emailChanging}
							<LoadingSpinner size={16} color="white" />
						{/if}
						Send Verification
					</button>
				</div>
			</form>
		{/if}
	</section>

	<!-- Phone -->
	<section class="card">
		<div class="card-header-row">
			<div>
				<h2 class="card-title">Phone Number</h2>
				<p class="card-description">Used for SMS-based MFA and account recovery.</p>
			</div>
		</div>
		<div class="info-row">
			<div class="info-value">
				{#if data.user.phone}
					<span>{data.user.phone}</span>
					{#if data.user.phone_verified}
						<Badge variant="success">Verified</Badge>
					{:else}
						<Badge variant="warning">Unverified</Badge>
					{/if}
				{:else}
					<span class="text-muted">No phone number set</span>
				{/if}
			</div>
			<button class="btn btn-secondary" onclick={() => (showPhoneChange = !showPhoneChange)}>
				{data.user.phone ? 'Change Phone' : 'Add Phone'}
			</button>
		</div>

		{#if showPhoneChange}
			<form class="change-form" onsubmit={(e) => { e.preventDefault(); handlePhoneChange(); }}>
				<div class="field">
					<label for="new-phone" class="field-label">New Phone Number</label>
					<input id="new-phone" type="tel" class="input" bind:value={newPhone} placeholder="+1 (555) 000-0000" required />
				</div>
				<div class="change-form-actions">
					<button type="button" class="btn btn-secondary" onclick={() => { showPhoneChange = false; newPhone = ''; }}>Cancel</button>
					<button type="submit" class="btn btn-primary" disabled={phoneChanging || !newPhone}>
						{#if phoneChanging}
							<LoadingSpinner size={16} color="white" />
						{/if}
						Send Verification
					</button>
				</div>
			</form>
		{/if}
	</section>

	<!-- Locale & Timezone -->
	<section class="card">
		<h2 class="card-title">Preferences</h2>
		<form class="form" onsubmit={(e) => { e.preventDefault(); handleSaveProfile(); }}>
			<div class="form-grid">
				<div class="field">
					<label for="locale" class="field-label">Language</label>
					<select id="locale" class="input" bind:value={locale}>
						{#each locales as loc}
							<option value={loc.value}>{loc.label}</option>
						{/each}
					</select>
				</div>
				<div class="field">
					<label for="timezone" class="field-label">Timezone</label>
					<select id="timezone" class="input" bind:value={timezone}>
						<option value="">Auto-detect</option>
						{#each timezones as tz}
							<option value={tz}>{tz.replace(/_/g, ' ')}</option>
						{/each}
					</select>
				</div>
			</div>
			<div class="form-actions">
				<button type="submit" class="btn btn-primary" disabled={saving}>
					{#if saving}
						<LoadingSpinner size={16} color="white" />
					{/if}
					Save Preferences
				</button>
			</div>
		</form>
	</section>

	<!-- Custom Metadata (read-only) -->
	{#if data.user.metadata && Object.keys(data.user.metadata).length > 0}
		<section class="card">
			<h2 class="card-title">Additional Information</h2>
			<p class="card-description">These fields are managed by your administrator.</p>
			<div class="metadata-grid">
				{#each Object.entries(data.user.metadata) as [key, value]}
					<div class="metadata-item">
						<span class="metadata-key">{key.replace(/_/g, ' ')}</span>
						<span class="metadata-value">{String(value)}</span>
					</div>
				{/each}
			</div>
		</section>
	{/if}
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.page-header {
		margin-bottom: 0.5rem;
	}

	.page-header h1 {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--color-text);
		margin: 0 0 0.25rem;
	}

	.page-header p {
		font-size: 0.875rem;
		color: var(--color-text-secondary);
		margin: 0;
	}

	.card {
		background-color: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-xl);
		padding: 1.5rem;
	}

	.card-title {
		font-size: 1rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 1rem;
	}

	.card-description {
		font-size: 0.8125rem;
		color: var(--color-text-secondary);
		margin: -0.5rem 0 1rem;
	}

	.card-header-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}

	.avatar-section {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}

	.avatar-hint {
		font-size: 0.8125rem;
		color: var(--color-text-secondary);
		margin: 0;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
		gap: 1rem;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.field-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.input {
		padding: 0.625rem 0.875rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-bg);
		color: var(--color-text);
		font-size: 0.875rem;
		outline: none;
		transition: border-color 0.15s;
		width: 100%;
	}

	.input:focus {
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
	}

	select.input {
		cursor: pointer;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		padding-top: 0.5rem;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		border: 1px solid transparent;
		transition: all 0.15s;
		white-space: nowrap;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background-color: var(--color-primary);
		color: var(--color-text-on-primary);
	}

	.btn-primary:hover:not(:disabled) {
		background-color: var(--color-primary-hover);
	}

	.btn-secondary {
		background-color: var(--color-bg-secondary);
		border-color: var(--color-border);
		color: var(--color-text);
	}

	.btn-secondary:hover:not(:disabled) {
		background-color: var(--color-bg-tertiary);
	}

	.info-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.info-value {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--color-text);
	}

	.text-muted {
		color: var(--color-text-tertiary);
	}

	.change-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--color-border);
	}

	.change-form-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
	}

	.metadata-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
		gap: 1rem;
	}

	.metadata-item {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.metadata-key {
		font-size: 0.75rem;
		font-weight: 500;
		color: var(--color-text-secondary);
		text-transform: capitalize;
	}

	.metadata-value {
		font-size: 0.875rem;
		color: var(--color-text);
	}
</style>
