<script lang="ts">
	let {
		value = $bindable(''),
		length = 6,
		error = '',
		onsubmit
	}: {
		value: string;
		length?: number;
		error?: string;
		onsubmit?: (code: string) => void;
	} = $props();

	let inputs: HTMLInputElement[] = $state([]);
	const digits = $derived(value.padEnd(length, '').split('').slice(0, length));

	function handleInput(index: number, event: Event) {
		const target = event.target as HTMLInputElement;
		const char = target.value.slice(-1);

		if (!/^\d$/.test(char) && char !== '') {
			target.value = '';
			return;
		}

		const newDigits = [...digits];
		newDigits[index] = char;
		value = newDigits.join('').replace(/\s/g, '');

		if (char && index < length - 1) {
			inputs[index + 1]?.focus();
		}

		if (value.length === length && onsubmit) {
			onsubmit(value);
		}
	}

	function handleKeydown(index: number, event: KeyboardEvent) {
		if (event.key === 'Backspace') {
			if (!digits[index] && index > 0) {
				const newDigits = [...digits];
				newDigits[index - 1] = '';
				value = newDigits.join('').replace(/\s/g, '');
				inputs[index - 1]?.focus();
			} else {
				const newDigits = [...digits];
				newDigits[index] = '';
				value = newDigits.join('').replace(/\s/g, '');
			}
		} else if (event.key === 'ArrowLeft' && index > 0) {
			inputs[index - 1]?.focus();
		} else if (event.key === 'ArrowRight' && index < length - 1) {
			inputs[index + 1]?.focus();
		}
	}

	function handlePaste(event: ClipboardEvent) {
		event.preventDefault();
		const pasted = event.clipboardData?.getData('text') || '';
		const digits = pasted.replace(/\D/g, '').slice(0, length);
		value = digits;

		if (digits.length > 0) {
			const focusIndex = Math.min(digits.length, length - 1);
			inputs[focusIndex]?.focus();
		}

		if (digits.length === length && onsubmit) {
			onsubmit(digits);
		}
	}
</script>

<div class="otp-container">
	<div class="otp-inputs" role="group" aria-label="Verification code">
		{#each { length } as _, i}
			<input
				bind:this={inputs[i]}
				type="text"
				inputmode="numeric"
				maxlength="2"
				class="otp-input"
				class:otp-error={!!error}
				value={digits[i]?.trim() || ''}
				oninput={(e) => handleInput(i, e)}
				onkeydown={(e) => handleKeydown(i, e)}
				onpaste={handlePaste}
				aria-label={`Digit ${i + 1}`}
			/>
			{#if i === 2}
				<span class="otp-separator">-</span>
			{/if}
		{/each}
	</div>
	{#if error}
		<p class="otp-error-text">{error}</p>
	{/if}
</div>

<style>
	.otp-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
	}

	.otp-inputs {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.otp-input {
		width: 2.75rem;
		height: 3rem;
		text-align: center;
		font-size: 1.25rem;
		font-weight: 600;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background-color: var(--color-bg);
		color: var(--color-text);
		outline: none;
		transition: border-color 0.15s;
	}

	.otp-input:focus {
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
	}

	.otp-input.otp-error {
		border-color: var(--color-danger);
	}

	.otp-separator {
		font-size: 1.5rem;
		color: var(--color-text-tertiary);
		margin: 0 0.125rem;
	}

	.otp-error-text {
		font-size: 0.75rem;
		color: var(--color-danger);
		margin: 0;
	}
</style>
