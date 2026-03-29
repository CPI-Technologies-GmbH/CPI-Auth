<script lang="ts">
	import { api } from '$lib/api';
	import type { MFAMethod, Passkey, RecoveryCodes } from '$lib/types';
	import PasswordInput from '$components/PasswordInput.svelte';
	import OTPInput from '$components/OTPInput.svelte';
	import QRCode from '$components/QRCode.svelte';
	import Alert from '$components/Alert.svelte';
	import Badge from '$components/Badge.svelte';
	import ConfirmDialog from '$components/ConfirmDialog.svelte';
	import EmptyState from '$components/EmptyState.svelte';
	import LoadingSpinner from '$components/LoadingSpinner.svelte';
	import { formatDate, formatRelativeTime } from '$lib/utils/format';

	// Password change
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let passwordSaving = $state(false);
	let passwordError = $state('');
	let passwordSuccess = $state('');

	// MFA
	let mfaMethods = $state<MFAMethod[]>([]);
	let mfaLoading = $state(true);
	let mfaError = $state('');

	// TOTP enrollment
	let showTOTPEnroll = $state(false);
	let totpUri = $state('');
	let totpSecret = $state('');
	let totpId = $state('');
	let totpCode = $state('');
	let totpVerifying = $state(false);
	let totpError = $state('');

	// Email/SMS OTP
	let showEmailOTP = $state(false);
	let emailOTPEnrolling = $state(false);
	let showSMSOTP = $state(false);
	let smsPhone = $state('');
	let smsOTPEnrolling = $state(false);

	// Recovery codes
	let recoveryCodes = $state<RecoveryCodes | null>(null);
	let showRecoveryCodes = $state(false);
	let recoveryCodesLoading = $state(false);
	let regeneratingCodes = $state(false);

	// MFA removal
	let mfaToRemove = $state<MFAMethod | null>(null);
	let removingMFA = $state(false);

	// Passkeys
	let passkeys = $state<Passkey[]>([]);
	let passkeysLoading = $state(true);
	let passkeysError = $state('');
	let registeringPasskey = $state(false);
	let passkeyToDelete = $state<Passkey | null>(null);
	let deletingPasskey = $state(false);
	let passkeyToRename = $state<Passkey | null>(null);
	let newPasskeyName = $state('');
	let renamingPasskey = $state(false);

	let successMessage = $state('');

	// Load MFA and Passkeys
	async function loadMFA() {
		mfaLoading = true;
		try {
			mfaMethods = await api.getMFAMethods();
		} catch (err: unknown) {
			mfaError = (err as { message?: string })?.message || 'Failed to load MFA methods.';
		} finally {
			mfaLoading = false;
		}
	}

	async function loadPasskeys() {
		passkeysLoading = true;
		try {
			passkeys = await api.getPasskeys();
		} catch (err: unknown) {
			passkeysError = (err as { message?: string })?.message || 'Failed to load passkeys.';
		} finally {
			passkeysLoading = false;
		}
	}

	$effect(() => {
		loadMFA();
		loadPasskeys();
	});

	// Password change handler
	async function handleChangePassword() {
		passwordError = '';
		passwordSuccess = '';

		if (newPassword !== confirmPassword) {
			passwordError = 'Passwords do not match.';
			return;
		}
		if (newPassword.length < 8) {
			passwordError = 'Password must be at least 8 characters.';
			return;
		}

		passwordSaving = true;
		try {
			await api.changePassword({
				current_password: currentPassword,
				new_password: newPassword
			});
			passwordSuccess = 'Password changed successfully.';
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		} catch (err: unknown) {
			passwordError = (err as { message?: string })?.message || 'Failed to change password.';
		} finally {
			passwordSaving = false;
		}
	}

	// TOTP enrollment
	async function handleEnrollTOTP() {
		showTOTPEnroll = true;
		totpError = '';
		try {
			const enrollment = await api.enrollTOTP();
			totpUri = enrollment.uri;
			totpSecret = enrollment.secret;
			totpId = enrollment.id;
		} catch (err: unknown) {
			totpError = (err as { message?: string })?.message || 'Failed to start TOTP enrollment.';
		}
	}

	async function handleVerifyTOTP(code: string) {
		totpVerifying = true;
		totpError = '';
		try {
			await api.verifyTOTP(totpId, code);
			successMessage = 'Authenticator app added successfully.';
			showTOTPEnroll = false;
			totpCode = '';
			await loadMFA();
		} catch (err: unknown) {
			totpError = (err as { message?: string })?.message || 'Invalid verification code.';
		} finally {
			totpVerifying = false;
		}
	}

	// Email OTP
	async function handleEnrollEmailOTP() {
		emailOTPEnrolling = true;
		try {
			await api.enrollEmailOTP();
			successMessage = 'Email OTP has been enabled.';
			showEmailOTP = false;
			await loadMFA();
		} catch (err: unknown) {
			mfaError = (err as { message?: string })?.message || 'Failed to enable email OTP.';
		} finally {
			emailOTPEnrolling = false;
		}
	}

	// SMS OTP
	async function handleEnrollSMSOTP() {
		if (!smsPhone) return;
		smsOTPEnrolling = true;
		try {
			await api.enrollSMSOTP(smsPhone);
			successMessage = 'SMS OTP has been enabled.';
			showSMSOTP = false;
			smsPhone = '';
			await loadMFA();
		} catch (err: unknown) {
			mfaError = (err as { message?: string })?.message || 'Failed to enable SMS OTP.';
		} finally {
			smsOTPEnrolling = false;
		}
	}

	// Remove MFA
	async function handleRemoveMFA() {
		if (!mfaToRemove) return;
		removingMFA = true;
		try {
			await api.removeMFA(mfaToRemove.id);
			successMessage = `${mfaToRemove.type.toUpperCase()} has been removed.`;
			mfaToRemove = null;
			await loadMFA();
		} catch (err: unknown) {
			mfaError = (err as { message?: string })?.message || 'Failed to remove MFA method.';
		} finally {
			removingMFA = false;
		}
	}

	// Recovery codes
	async function handleViewRecoveryCodes() {
		recoveryCodesLoading = true;
		showRecoveryCodes = true;
		try {
			recoveryCodes = await api.getRecoveryCodes();
		} catch (err: unknown) {
			mfaError = (err as { message?: string })?.message || 'Failed to load recovery codes.';
			showRecoveryCodes = false;
		} finally {
			recoveryCodesLoading = false;
		}
	}

	async function handleRegenerateRecoveryCodes() {
		regeneratingCodes = true;
		try {
			recoveryCodes = await api.regenerateRecoveryCodes();
			successMessage = 'Recovery codes regenerated. Please save them securely.';
		} catch (err: unknown) {
			mfaError = (err as { message?: string })?.message || 'Failed to regenerate recovery codes.';
		} finally {
			regeneratingCodes = false;
		}
	}

	// Passkey registration
	async function handleRegisterPasskey() {
		registeringPasskey = true;
		passkeysError = '';
		try {
			const { options } = await api.beginPasskeyRegistration();
			const credential = await navigator.credentials.create({
				publicKey: options as PublicKeyCredentialCreationOptions
			});
			if (credential) {
				await api.finishPasskeyRegistration(credential);
				successMessage = 'Passkey registered successfully.';
				await loadPasskeys();
			}
		} catch (err: unknown) {
			if ((err as Error)?.name !== 'NotAllowedError') {
				passkeysError = (err as { message?: string })?.message || 'Failed to register passkey.';
			}
		} finally {
			registeringPasskey = false;
		}
	}

	// Passkey rename
	async function handleRenamePasskey() {
		if (!passkeyToRename || !newPasskeyName) return;
		renamingPasskey = true;
		try {
			await api.renamePasskey(passkeyToRename.id, newPasskeyName);
			successMessage = 'Passkey renamed successfully.';
			passkeyToRename = null;
			newPasskeyName = '';
			await loadPasskeys();
		} catch (err: unknown) {
			passkeysError = (err as { message?: string })?.message || 'Failed to rename passkey.';
		} finally {
			renamingPasskey = false;
		}
	}

	// Passkey delete
	async function handleDeletePasskey() {
		if (!passkeyToDelete) return;
		deletingPasskey = true;
		try {
			await api.deletePasskey(passkeyToDelete.id);
			successMessage = 'Passkey deleted successfully.';
			passkeyToDelete = null;
			await loadPasskeys();
		} catch (err: unknown) {
			passkeysError = (err as { message?: string })?.message || 'Failed to delete passkey.';
		} finally {
			deletingPasskey = false;
		}
	}

	function getMFATypeLabel(type: string): string {
		switch (type) {
			case 'totp': return 'Authenticator App';
			case 'email': return 'Email OTP';
			case 'sms': return 'SMS OTP';
			default: return type;
		}
	}

	function getMFAIcon(type: string): string {
		switch (type) {
			case 'totp': return 'M12 18h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z';
			case 'email': return 'M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2zM22 6l-10 7L2 6';
			case 'sms': return 'M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z';
			default: return 'M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z';
		}
	}
</script>

<svelte:head>
	<title>Security - CPI Auth Account</title>
</svelte:head>

<div class="page">
	<div class="page-header">
		<h1>Security</h1>
		<p>Manage your password, multi-factor authentication, and passkeys.</p>
	</div>

	{#if successMessage}
		<Alert variant="success" dismissible ondismiss={() => (successMessage = '')}>{successMessage}</Alert>
	{/if}

	<!-- Change Password -->
	<section class="card">
		<h2 class="card-title">Change Password</h2>
		<p class="card-description">Choose a strong password that you don't use on other sites.</p>

		{#if passwordSuccess}
			<Alert variant="success" dismissible ondismiss={() => (passwordSuccess = '')}>{passwordSuccess}</Alert>
		{/if}
		{#if passwordError}
			<Alert variant="danger" dismissible ondismiss={() => (passwordError = '')}>{passwordError}</Alert>
		{/if}

		<form class="form" onsubmit={(e) => { e.preventDefault(); handleChangePassword(); }}>
			<PasswordInput
				bind:value={currentPassword}
				label="Current Password"
				name="current-password"
				autocomplete="current-password"
				required
			/>
			<PasswordInput
				bind:value={newPassword}
				label="New Password"
				name="new-password"
				autocomplete="new-password"
				showStrength
				required
			/>
			<PasswordInput
				bind:value={confirmPassword}
				label="Confirm New Password"
				name="confirm-password"
				autocomplete="new-password"
				error={confirmPassword && newPassword !== confirmPassword ? 'Passwords do not match' : ''}
				required
			/>
			<div class="form-actions">
				<button type="submit" class="btn btn-primary" disabled={passwordSaving || !currentPassword || !newPassword || !confirmPassword}>
					{#if passwordSaving}
						<LoadingSpinner size={16} color="white" />
					{/if}
					Update Password
				</button>
			</div>
		</form>
	</section>

	<!-- Multi-Factor Authentication -->
	<section class="card">
		<div class="card-header-row">
			<div>
				<h2 class="card-title">Multi-Factor Authentication</h2>
				<p class="card-description">Add extra security to your account with additional authentication methods.</p>
			</div>
		</div>

		{#if mfaError}
			<Alert variant="danger" dismissible ondismiss={() => (mfaError = '')}>{mfaError}</Alert>
		{/if}

		{#if mfaLoading}
			<div class="loading-center">
				<LoadingSpinner size={32} />
			</div>
		{:else}
			<!-- Current MFA methods -->
			{#if mfaMethods.length > 0}
				<div class="mfa-list">
					{#each mfaMethods as method}
						<div class="mfa-item">
							<div class="mfa-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
									<path d={getMFAIcon(method.type)} />
								</svg>
							</div>
							<div class="mfa-info">
								<div class="mfa-name">
									{getMFATypeLabel(method.type)}
									{#if method.verified}
										<Badge variant="success">Active</Badge>
									{:else}
										<Badge variant="warning">Pending</Badge>
									{/if}
								</div>
								<span class="mfa-meta">Added {formatDate(method.created_at)}</span>
							</div>
							<button class="btn btn-danger-outline btn-sm" onclick={() => (mfaToRemove = method)}>
								Remove
							</button>
						</div>
					{/each}
				</div>
			{/if}

			<!-- Add MFA buttons -->
			<div class="mfa-actions">
				<button class="btn btn-secondary" onclick={handleEnrollTOTP} disabled={showTOTPEnroll}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon">
						<path d="M12 18h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
					</svg>
					Add Authenticator App
				</button>
				<button class="btn btn-secondary" onclick={() => (showEmailOTP = true)} disabled={showEmailOTP}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon">
						<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2zM22 6l-10 7L2 6" />
					</svg>
					Add Email OTP
				</button>
				<button class="btn btn-secondary" onclick={() => (showSMSOTP = true)} disabled={showSMSOTP}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon">
						<path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z" />
					</svg>
					Add SMS OTP
				</button>
			</div>

			<!-- TOTP enrollment -->
			{#if showTOTPEnroll}
				<div class="enroll-section">
					<h3 class="enroll-title">Set Up Authenticator App</h3>
					{#if totpUri}
						<div class="totp-setup">
							<div class="totp-qr">
								<QRCode data={totpUri} size={180} />
							</div>
							<div class="totp-instructions">
								<p>Scan the QR code with your authenticator app (e.g., Google Authenticator, Authy).</p>
								<div class="totp-secret">
									<span class="secret-label">Or enter this key manually:</span>
									<code class="secret-value">{totpSecret}</code>
								</div>
								<div class="totp-verify">
									<p class="verify-label">Enter the 6-digit code from your app:</p>
									{#if totpError}
										<Alert variant="danger" dismissible ondismiss={() => (totpError = '')}>{totpError}</Alert>
									{/if}
									<OTPInput bind:value={totpCode} onsubmit={handleVerifyTOTP} error="" />
									<div class="enroll-actions">
										<button class="btn btn-secondary" onclick={() => { showTOTPEnroll = false; totpCode = ''; }}>Cancel</button>
										<button class="btn btn-primary" onclick={() => handleVerifyTOTP(totpCode)} disabled={totpVerifying || totpCode.length !== 6}>
											{#if totpVerifying}
												<LoadingSpinner size={16} color="white" />
											{/if}
											Verify
										</button>
									</div>
								</div>
							</div>
						</div>
					{:else}
						<div class="loading-center">
							<LoadingSpinner size={24} />
						</div>
					{/if}
				</div>
			{/if}

			<!-- Email OTP enrollment -->
			{#if showEmailOTP}
				<div class="enroll-section">
					<h3 class="enroll-title">Enable Email OTP</h3>
					<p>A verification code will be sent to your email address when you sign in.</p>
					<div class="enroll-actions">
						<button class="btn btn-secondary" onclick={() => (showEmailOTP = false)}>Cancel</button>
						<button class="btn btn-primary" onclick={handleEnrollEmailOTP} disabled={emailOTPEnrolling}>
							{#if emailOTPEnrolling}
								<LoadingSpinner size={16} color="white" />
							{/if}
							Enable Email OTP
						</button>
					</div>
				</div>
			{/if}

			<!-- SMS OTP enrollment -->
			{#if showSMSOTP}
				<div class="enroll-section">
					<h3 class="enroll-title">Enable SMS OTP</h3>
					<form onsubmit={(e) => { e.preventDefault(); handleEnrollSMSOTP(); }}>
						<div class="field">
							<label for="sms-phone" class="field-label">Phone Number</label>
							<input id="sms-phone" type="tel" class="input" bind:value={smsPhone} placeholder="+1 (555) 000-0000" required />
						</div>
						<div class="enroll-actions">
							<button type="button" class="btn btn-secondary" onclick={() => { showSMSOTP = false; smsPhone = ''; }}>Cancel</button>
							<button type="submit" class="btn btn-primary" disabled={smsOTPEnrolling || !smsPhone}>
								{#if smsOTPEnrolling}
									<LoadingSpinner size={16} color="white" />
								{/if}
								Enable SMS OTP
							</button>
						</div>
					</form>
				</div>
			{/if}

			<!-- Recovery Codes -->
			{#if mfaMethods.length > 0}
				<div class="recovery-section">
					<h3 class="section-subtitle">Recovery Codes</h3>
					<p class="card-description">Recovery codes can be used to access your account if you lose your MFA device.</p>
					{#if showRecoveryCodes && recoveryCodes}
						<div class="recovery-codes-box">
							<div class="recovery-codes-grid">
								{#each recoveryCodes.codes as code}
									<code class="recovery-code">{code}</code>
								{/each}
							</div>
							<p class="recovery-warning">Store these codes in a safe place. Each code can only be used once.</p>
							<div class="recovery-actions">
								<button class="btn btn-secondary" onclick={() => (showRecoveryCodes = false)}>Hide</button>
								<button class="btn btn-danger-outline" onclick={handleRegenerateRecoveryCodes} disabled={regeneratingCodes}>
									{#if regeneratingCodes}
										<LoadingSpinner size={16} />
									{/if}
									Regenerate Codes
								</button>
							</div>
						</div>
					{:else}
						<button class="btn btn-secondary" onclick={handleViewRecoveryCodes} disabled={recoveryCodesLoading}>
							{#if recoveryCodesLoading}
								<LoadingSpinner size={16} />
							{/if}
							View Recovery Codes
						</button>
					{/if}
				</div>
			{/if}
		{/if}
	</section>

	<!-- Passkeys -->
	<section class="card">
		<div class="card-header-row">
			<div>
				<h2 class="card-title">Passkeys</h2>
				<p class="card-description">Sign in securely with fingerprint, face, or device PIN.</p>
			</div>
			<button class="btn btn-primary" onclick={handleRegisterPasskey} disabled={registeringPasskey}>
				{#if registeringPasskey}
					<LoadingSpinner size={16} color="white" />
				{/if}
				Add Passkey
			</button>
		</div>

		{#if passkeysError}
			<Alert variant="danger" dismissible ondismiss={() => (passkeysError = '')}>{passkeysError}</Alert>
		{/if}

		{#if passkeysLoading}
			<div class="loading-center">
				<LoadingSpinner size={32} />
			</div>
		{:else if passkeys.length === 0}
			<EmptyState
				icon="key"
				title="No passkeys registered"
				description="Add a passkey to sign in without a password using your device's biometrics or PIN."
			/>
		{:else}
			<div class="passkeys-list">
				{#each passkeys as passkey}
					<div class="passkey-item">
						<div class="passkey-icon">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
								<path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
							</svg>
						</div>
						<div class="passkey-info">
							<span class="passkey-name">{passkey.name}</span>
							<span class="passkey-meta">
								Created {formatDate(passkey.created_at)}
								{#if passkey.last_used_at}
									&middot; Last used {formatRelativeTime(passkey.last_used_at)}
								{/if}
							</span>
						</div>
						<div class="passkey-actions">
							<button
								class="btn btn-ghost btn-sm"
								onclick={() => { passkeyToRename = passkey; newPasskeyName = passkey.name; }}
								title="Rename"
							>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon-only">
									<path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
									<path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
								</svg>
							</button>
							<button
								class="btn btn-ghost btn-sm"
								onclick={() => (passkeyToDelete = passkey)}
								title="Delete"
							>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon-only delete-icon">
									<polyline points="3,6 5,6 21,6" />
									<path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" />
								</svg>
							</button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</section>

	<!-- Rename Passkey Dialog -->
	{#if passkeyToRename}
		<ConfirmDialog
			open={true}
			title="Rename Passkey"
			confirmText="Rename"
			variant="primary"
			loading={renamingPasskey}
			onconfirm={handleRenamePasskey}
			oncancel={() => { passkeyToRename = null; newPasskeyName = ''; }}
		>
			<div class="field">
				<label for="passkey-name" class="field-label">Passkey Name</label>
				<input id="passkey-name" type="text" class="input" bind:value={newPasskeyName} placeholder="e.g., MacBook Pro" />
			</div>
		</ConfirmDialog>
	{/if}

	<!-- Delete Passkey Dialog -->
	<ConfirmDialog
		open={!!passkeyToDelete}
		title="Delete Passkey"
		confirmText="Delete"
		variant="danger"
		loading={deletingPasskey}
		onconfirm={handleDeletePasskey}
		oncancel={() => (passkeyToDelete = null)}
	>
		<p>Are you sure you want to delete the passkey "{passkeyToDelete?.name}"? You won't be able to use it to sign in anymore.</p>
	</ConfirmDialog>

	<!-- Remove MFA Dialog -->
	<ConfirmDialog
		open={!!mfaToRemove}
		title="Remove MFA Method"
		confirmText="Remove"
		variant="danger"
		loading={removingMFA}
		onconfirm={handleRemoveMFA}
		oncancel={() => (mfaToRemove = null)}
	>
		<p>Are you sure you want to remove {getMFATypeLabel(mfaToRemove?.type || '')}? This will reduce the security of your account.</p>
	</ConfirmDialog>
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
		margin: 0 0 0.25rem;
	}

	.card-description {
		font-size: 0.8125rem;
		color: var(--color-text-secondary);
		margin: 0 0 1rem;
	}

	.card-header-row {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		padding-top: 0.5rem;
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

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.8125rem;
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

	.btn-danger-outline {
		background: transparent;
		border-color: var(--color-danger);
		color: var(--color-danger);
	}

	.btn-danger-outline:hover:not(:disabled) {
		background-color: color-mix(in srgb, var(--color-danger) 10%, transparent);
	}

	.btn-ghost {
		background: transparent;
		border: none;
		color: var(--color-text-secondary);
		padding: 0.375rem;
	}

	.btn-ghost:hover:not(:disabled) {
		background-color: var(--color-bg-tertiary);
		color: var(--color-text);
	}

	.btn-icon {
		width: 1rem;
		height: 1rem;
	}

	.btn-icon-only {
		width: 1rem;
		height: 1rem;
	}

	.delete-icon:hover {
		color: var(--color-danger);
	}

	.loading-center {
		display: flex;
		justify-content: center;
		padding: 2rem 0;
	}

	/* MFA List */
	.mfa-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.mfa-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.875rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
	}

	.mfa-icon {
		width: 2.25rem;
		height: 2.25rem;
		padding: 0.375rem;
		border-radius: var(--radius-md);
		background-color: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
		flex-shrink: 0;
	}

	.mfa-icon svg {
		width: 100%;
		height: 100%;
	}

	.mfa-info {
		flex: 1;
		min-width: 0;
	}

	.mfa-name {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.mfa-meta {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.mfa-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}

	/* Enrollment Sections */
	.enroll-section {
		margin-top: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--color-border);
	}

	.enroll-title {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 0.75rem;
	}

	.enroll-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: flex-end;
		margin-top: 1rem;
	}

	.totp-setup {
		display: flex;
		gap: 2rem;
		align-items: flex-start;
		flex-wrap: wrap;
	}

	.totp-qr {
		flex-shrink: 0;
	}

	.totp-instructions {
		flex: 1;
		min-width: 14rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.totp-instructions p {
		font-size: 0.875rem;
		color: var(--color-text-secondary);
		margin: 0;
	}

	.totp-secret {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.secret-label {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.secret-value {
		font-family: monospace;
		font-size: 0.8125rem;
		padding: 0.5rem 0.75rem;
		background-color: var(--color-bg-tertiary);
		border-radius: var(--radius-sm);
		word-break: break-all;
		user-select: all;
	}

	.totp-verify {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.verify-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
		margin: 0;
	}

	/* Recovery Codes */
	.recovery-section {
		margin-top: 1.5rem;
		padding-top: 1.5rem;
		border-top: 1px solid var(--color-border);
	}

	.section-subtitle {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--color-text);
		margin: 0 0 0.25rem;
	}

	.recovery-codes-box {
		margin-top: 1rem;
		padding: 1rem;
		background-color: var(--color-bg-secondary);
		border-radius: var(--radius-lg);
	}

	.recovery-codes-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
		gap: 0.5rem;
		margin-bottom: 0.75rem;
	}

	.recovery-code {
		font-family: monospace;
		font-size: 0.875rem;
		padding: 0.375rem 0.75rem;
		background-color: var(--color-bg);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		text-align: center;
	}

	.recovery-warning {
		font-size: 0.75rem;
		color: var(--color-warning);
		margin: 0 0 0.75rem;
	}

	.recovery-actions {
		display: flex;
		gap: 0.75rem;
	}

	/* Passkeys */
	.passkeys-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.passkey-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.875rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
	}

	.passkey-icon {
		width: 2.25rem;
		height: 2.25rem;
		padding: 0.375rem;
		border-radius: var(--radius-md);
		background-color: var(--color-bg-tertiary);
		color: var(--color-text-secondary);
		flex-shrink: 0;
	}

	.passkey-icon svg {
		width: 100%;
		height: 100%;
	}

	.passkey-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.passkey-name {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--color-text);
	}

	.passkey-meta {
		font-size: 0.75rem;
		color: var(--color-text-secondary);
	}

	.passkey-actions {
		display: flex;
		gap: 0.25rem;
	}
</style>
