<script lang="ts">
	import { evaluatePasswordStrength } from '$lib/utils/password';

	let {
		value = $bindable(''),
		label = 'Password',
		placeholder = '',
		showStrength = false,
		error = '',
		name = 'password',
		autocomplete = 'current-password',
		required = false
	}: {
		value: string;
		label?: string;
		placeholder?: string;
		showStrength?: boolean;
		error?: string;
		name?: string;
		autocomplete?: string;
		required?: boolean;
	} = $props();

	let showPassword = $state(false);

	const strength = $derived(showStrength ? evaluatePasswordStrength(value) : null);
</script>

<div class="field">
	<label for={name} class="field-label">{label}</label>
	<div class="input-wrapper">
		<input
			id={name}
			{name}
			type={showPassword ? 'text' : 'password'}
			bind:value
			{placeholder}
			{autocomplete}
			{required}
			class="input"
			class:input-error={!!error}
		/>
		<button
			type="button"
			class="toggle-visibility"
			onclick={() => (showPassword = !showPassword)}
			aria-label={showPassword ? 'Hide password' : 'Show password'}
		>
			{#if showPassword}
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24" />
					<line x1="1" y1="1" x2="23" y2="23" />
				</svg>
			{:else}
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
					<circle cx="12" cy="12" r="3" />
				</svg>
			{/if}
		</button>
	</div>

	{#if showStrength && value && strength}
		<div class="strength-meter">
			<div class="strength-bars">
				{#each Array(4) as _, i}
					<div
						class="strength-bar"
						style="background-color: {i < strength.score ? strength.color : 'var(--color-border)'}"
					></div>
				{/each}
			</div>
			<span class="strength-label" style="color: {strength.color}">{strength.label}</span>
		</div>
		{#if strength.suggestions.length > 0}
			<ul class="strength-suggestions">
				{#each strength.suggestions as suggestion}
					<li>{suggestion}</li>
				{/each}
			</ul>
		{/if}
	{/if}

	{#if error}
		<p class="field-error">{error}</p>
	{/if}
</div>

<style>
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

	.input-wrapper {
		position: relative;
		display: flex;
	}

	.input {
		width: 100%;
		padding: 0.625rem 2.75rem 0.625rem 0.875rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-bg);
		color: var(--color-text);
		font-size: 0.875rem;
		outline: none;
		transition: border-color 0.15s;
	}

	.input:focus {
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
	}

	.input-error {
		border-color: var(--color-danger);
	}

	.toggle-visibility {
		position: absolute;
		right: 0.5rem;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.25rem;
		color: var(--color-text-tertiary);
		display: flex;
		align-items: center;
	}

	.toggle-visibility:hover {
		color: var(--color-text-secondary);
	}

	.toggle-visibility svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	.strength-meter {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.25rem;
	}

	.strength-bars {
		display: flex;
		gap: 0.25rem;
		flex: 1;
	}

	.strength-bar {
		height: 4px;
		flex: 1;
		border-radius: 2px;
		transition: background-color 0.2s;
	}

	.strength-label {
		font-size: 0.75rem;
		font-weight: 500;
		white-space: nowrap;
	}

	.strength-suggestions {
		margin: 0.25rem 0 0;
		padding: 0 0 0 1.25rem;
		font-size: 0.75rem;
		color: var(--color-text-tertiary);
		list-style: disc;
	}

	.strength-suggestions li {
		margin-bottom: 0.125rem;
	}

	.field-error {
		font-size: 0.75rem;
		color: var(--color-danger);
		margin: 0;
	}
</style>
