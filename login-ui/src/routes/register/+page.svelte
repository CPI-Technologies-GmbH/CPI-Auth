<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { t } from '$lib/i18n';
	import { api, ApiClientError } from '$lib/api';
	import { branding, extractOAuthParams, oauthParams } from '$lib/stores';
	import {
		validateEmail,
		validatePassword,
		validatePasswordMatch,
		validateRequired
	} from '$lib/utils/validation';
	import { redirectTo, getErrorMessage } from '$lib/utils';
	import AuthLayout from '$lib/components/AuthLayout.svelte';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import SocialButtons from '$lib/components/SocialButtons.svelte';
	import Alert from '$lib/components/Alert.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let termsAccepted = $state(false);
	let loading = $state(false);
	let error = $state('');
	let nameError = $state('');
	let emailError = $state('');
	let passwordError = $state('');
	let confirmError = $state('');
	let termsError = $state('');
	let customFieldValues = $state<Record<string, string>>({});
	let customFieldErrors = $state<Record<string, string>>({});

	let oauth = $derived(extractOAuthParams(new URL($page.url)));
	let providers = $derived($branding?.social_providers || []);
	let customFields = $derived($branding?.custom_fields || []);

	$effect(() => {
		if (oauth) oauthParams.set(oauth);
	});

	function validate(): boolean {
		let valid = true;

		const nameResult = validateRequired(name);
		nameError = nameResult.valid ? '' : (nameResult.error || '');
		if (!nameResult.valid) valid = false;

		const emailResult = validateEmail(email);
		emailError = emailResult.valid ? '' : (emailResult.error || '');
		if (!emailResult.valid) valid = false;

		const passResult = validatePassword(password);
		passwordError = passResult.valid ? '' : (passResult.error || '');
		if (!passResult.valid) valid = false;

		const confirmResult = validatePasswordMatch(password, confirmPassword);
		confirmError = confirmResult.valid ? '' : (confirmResult.error || '');
		if (!confirmResult.valid) valid = false;

		if (!termsAccepted) {
			termsError = 'validation.terms';
			valid = false;
		} else {
			termsError = '';
		}

		// Validate custom fields
		const cfErrors: Record<string, string> = {};
		for (const field of customFields) {
			if (field.required && !customFieldValues[field.name]?.trim()) {
				cfErrors[field.name] = 'validation.required';
				valid = false;
			}
		}
		customFieldErrors = cfErrors;

		return valid;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (!validate()) return;

		loading = true;

		try {
			const result = await api.register({
				email,
				password,
				name,
				custom_fields: Object.keys(customFieldValues).length > 0 ? customFieldValues : undefined,
				...(oauth && {
					client_id: oauth.client_id,
					redirect_uri: oauth.redirect_uri
				})
			});

			if (result.redirect_url) {
				redirectTo(result.redirect_url);
			} else {
				// Redirect to login, preserving OAuth params so the flow continues
				const params = new URLSearchParams({ registered: 'true' });
				if (oauth) {
					params.set('client_id', oauth.client_id);
					params.set('redirect_uri', oauth.redirect_uri);
					params.set('scope', oauth.scope || '');
					params.set('state', oauth.state || '');
					if (oauth.code_challenge) params.set('code_challenge', oauth.code_challenge);
					if (oauth.code_challenge_method) params.set('code_challenge_method', oauth.code_challenge_method);
					if (oauth.response_type) params.set('response_type', oauth.response_type);
				}
				await goto(`/login?${params.toString()}`);
			}
		} catch (err) {
			if (err instanceof ApiClientError) {
				error = err.message;
			} else {
				error = getErrorMessage(err);
			}
		} finally {
			loading = false;
		}
	}

	function buildLink(path: string): string {
		const params = new URLSearchParams();
		if (oauth) {
			params.set('client_id', oauth.client_id);
			params.set('redirect_uri', oauth.redirect_uri);
			params.set('scope', oauth.scope);
			params.set('state', oauth.state);
		}
		const qs = params.toString();
		return qs ? `${path}?${qs}` : path;
	}
</script>

<svelte:head>
	<title>{$t('register.title')}</title>
</svelte:head>

<AuthLayout maxWidth="md">
	<div class="space-y-6">
		<div class="text-center">
			<h1 class="text-2xl font-bold" style="color: var(--af-color-text)">
				{$t('register.title')}
			</h1>
		</div>

		{#if error}
			<Alert type="error" message={error} dismissible />
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4" novalidate>
			<div>
				<label for="name" class="af-label">{$t('register.name')}</label>
				<input
					id="name"
					name="name"
					type="text"
					bind:value={name}
					placeholder={$t('register.name.placeholder')}
					autocomplete="name"
					required
					disabled={loading}
					class="af-input"
					class:error={!!nameError}
					aria-invalid={!!nameError}
					aria-describedby={nameError ? 'name-error' : undefined}
				/>
				{#if nameError}
					<p id="name-error" class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
						{$t(nameError)}
					</p>
				{/if}
			</div>

			<div>
				<label for="email" class="af-label">{$t('register.email')}</label>
				<input
					id="email"
					name="email"
					type="email"
					bind:value={email}
					placeholder={$t('register.email.placeholder')}
					autocomplete="email"
					required
					disabled={loading}
					class="af-input"
					class:error={!!emailError}
					aria-invalid={!!emailError}
					aria-describedby={emailError ? 'email-error' : undefined}
				/>
				{#if emailError}
					<p id="email-error" class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
						{$t(emailError)}
					</p>
				{/if}
			</div>

			<PasswordInput
				bind:value={password}
				name="password"
				label={$t('register.password')}
				placeholder={$t('register.password.placeholder')}
				error={passwordError}
				showStrength
				disabled={loading}
				autocomplete="new-password"
			/>

			<PasswordInput
				bind:value={confirmPassword}
				name="confirmPassword"
				id="confirmPassword"
				label={$t('register.confirm_password')}
				placeholder={$t('register.confirm_password.placeholder')}
				error={confirmError}
				disabled={loading}
				autocomplete="new-password"
			/>

			{#each customFields as field}
				<div>
					<label for="custom-{field.name}" class="af-label">
						{field.label}
						{#if field.required}<span style="color: var(--af-color-error)">*</span>{/if}
					</label>
					{#if field.type === 'select' && field.options}
						<select
							id="custom-{field.name}"
							bind:value={customFieldValues[field.name]}
							disabled={loading}
							class="af-input"
							class:error={!!customFieldErrors[field.name]}
						>
							<option value="">Select...</option>
							{#each field.options as option}
								<option value={option}>{option}</option>
							{/each}
						</select>
					{:else if field.type === 'checkbox'}
						<label class="flex items-center gap-2 text-sm">
							<input
								type="checkbox"
								checked={customFieldValues[field.name] === 'true'}
								onchange={(e) => {
									customFieldValues[field.name] = (e.target as HTMLInputElement).checked ? 'true' : 'false';
								}}
								disabled={loading}
								class="af-checkbox"
							/>
							{field.label}
						</label>
					{:else}
						<input
							id="custom-{field.name}"
							type={field.type}
							bind:value={customFieldValues[field.name]}
							placeholder={field.placeholder || ''}
							disabled={loading}
							class="af-input"
							class:error={!!customFieldErrors[field.name]}
						/>
					{/if}
					{#if customFieldErrors[field.name]}
						<p class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
							{$t(customFieldErrors[field.name])}
						</p>
					{/if}
				</div>
			{/each}

			<div>
				<label class="flex items-start gap-2 text-sm" style="color: var(--af-color-text)">
					<input
						type="checkbox"
						bind:checked={termsAccepted}
						disabled={loading}
						class="af-checkbox mt-0.5"
					/>
					<span>
						{$t('register.terms')}
						<a href="/terms" class="af-link" target="_blank" rel="noopener noreferrer">
							{$t('register.terms_link')}
						</a>
						{$t('register.and')}
						<a href="/privacy" class="af-link" target="_blank" rel="noopener noreferrer">
							{$t('register.privacy_link')}
						</a>
					</span>
				</label>
				{#if termsError}
					<p class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
						{$t(termsError)}
					</p>
				{/if}
			</div>

			<button type="submit" class="af-btn af-btn-primary w-full" disabled={loading}>
				{#if loading}
					<LoadingSpinner size="sm" />
				{/if}
				{$t('register.submit')}
			</button>
		</form>

		{#if providers.length > 0}
			<div class="af-divider">{$t('register.or')}</div>
			<SocialButtons {providers} oauthParams={oauth} mode="register" />
		{/if}

		<p class="text-center text-sm" style="color: var(--af-color-text-muted)">
			{$t('register.has_account')}
			<a href={buildLink('/login')} class="af-link">{$t('register.signin_link')}</a>
		</p>
	</div>
</AuthLayout>
