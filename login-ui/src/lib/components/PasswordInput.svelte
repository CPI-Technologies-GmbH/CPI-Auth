<script lang="ts">
	import { t } from '$lib/i18n';
	import { getPasswordStrength } from '$lib/utils/validation';

	interface Props {
		value?: string;
		name?: string;
		id?: string;
		placeholder?: string;
		label?: string;
		error?: string;
		showStrength?: boolean;
		disabled?: boolean;
		autocomplete?: string;
		required?: boolean;
		oninput?: (e: Event) => void;
	}

	let {
		value = $bindable(''),
		name = 'password',
		id = name,
		placeholder = '',
		label = '',
		error = '',
		showStrength = false,
		disabled = false,
		autocomplete = 'current-password',
		required = false,
		oninput
	}: Props = $props();

	let showPassword = $state(false);
	let strength = $derived(showStrength ? getPasswordStrength(value) : null);

	function toggleVisibility() {
		showPassword = !showPassword;
	}
</script>

<div class="w-full">
	{#if label}
		<label for={id} class="af-label">{label}</label>
	{/if}

	<div class="relative">
		<input
			{id}
			{name}
			type={showPassword ? 'text' : 'password'}
			bind:value
			{placeholder}
			{disabled}
			{autocomplete}
			{required}
			{oninput}
			class="af-input pr-10"
			class:error={!!error}
			aria-invalid={!!error}
			aria-describedby={error ? `${id}-error` : undefined}
		/>

		<button
			type="button"
			onclick={toggleVisibility}
			class="absolute top-1/2 right-3 -translate-y-1/2 cursor-pointer"
			style="color: var(--af-color-text-muted)"
			aria-label={showPassword ? 'Hide password' : 'Show password'}
			tabindex="-1"
		>
			{#if showPassword}
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
					/>
				</svg>
			{:else}
				<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
					/>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
					/>
				</svg>
			{/if}
		</button>
	</div>

	{#if showStrength && strength && strength.score > 0}
		<div class="mt-2">
			<div class="flex gap-1">
				{#each Array(4) as _, i}
					<div
						class="h-1 flex-1 rounded-full transition-all duration-300"
						style="background-color: {i < strength.score
							? strength.color
							: 'var(--af-color-border)'}"
					></div>
				{/each}
			</div>
			<p class="mt-1 text-xs" style="color: {strength.color}">
				{$t(strength.label)}
			</p>
		</div>
	{/if}

	{#if error}
		<p id="{id}-error" class="mt-1 text-sm" style="color: var(--af-color-error)" role="alert">
			{$t(error)}
		</p>
	{/if}
</div>
