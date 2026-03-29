<script lang="ts">
	import { t } from '$lib/i18n';

	interface Props {
		value?: string;
		length?: number;
		label?: string;
		error?: string;
		disabled?: boolean;
	}

	let { value = $bindable(''), length = 6, label = '', error = '', disabled = false }: Props = $props();

	let inputs: HTMLInputElement[] = $state([]);
	let digits: string[] = $derived.by(() => {
		const d = Array(length).fill('');
		if (value) {
			value.split('').forEach((char, i) => {
				if (i < length) d[i] = char;
			});
		}
		return d;
	});

	function setDigit(index: number, val: string) {
		const arr = [...digits];
		arr[index] = val;
		value = arr.join('');
	}

	function setDigits(newDigits: string[]) {
		value = newDigits.join('');
	}

	function handleInput(index: number, e: Event) {
		const target = e.target as HTMLInputElement;
		const val = target.value.replace(/\D/g, '');

		if (val.length > 1) {
			// Handle paste across fields
			const chars = val.split('').slice(0, length);
			const arr = [...digits];
			chars.forEach((char, i) => {
				if (index + i < length) {
					arr[index + i] = char;
				}
			});
			setDigits(arr);
			const nextIndex = Math.min(index + chars.length, length - 1);
			inputs[nextIndex]?.focus();
		} else {
			setDigit(index, val);
			if (val && index < length - 1) {
				inputs[index + 1]?.focus();
			}
		}
	}

	function handleKeydown(index: number, e: KeyboardEvent) {
		if (e.key === 'Backspace' && !digits[index] && index > 0) {
			setDigit(index - 1, '');
			inputs[index - 1]?.focus();
		}
		if (e.key === 'ArrowLeft' && index > 0) {
			inputs[index - 1]?.focus();
		}
		if (e.key === 'ArrowRight' && index < length - 1) {
			inputs[index + 1]?.focus();
		}
	}

	function handlePaste(e: ClipboardEvent) {
		e.preventDefault();
		const pasted = (e.clipboardData?.getData('text') || '').replace(/\D/g, '').slice(0, length);
		const arr = Array(length).fill('');
		pasted.split('').forEach((char, i) => {
			arr[i] = char;
		});
		setDigits(arr);
		const nextIndex = Math.min(pasted.length, length - 1);
		inputs[nextIndex]?.focus();
	}

	function handleFocus(e: FocusEvent) {
		(e.target as HTMLInputElement).select();
	}
</script>

<div class="w-full">
	{#if label}
		<label class="af-label" for="otp-0">{label}</label>
	{/if}

	<div class="flex justify-center gap-2" role="group" aria-label="Verification code">
		{#each Array(length) as _, i}
			<input
				bind:this={inputs[i]}
				id="otp-{i}"
				type="text"
				inputmode="numeric"
				pattern="[0-9]*"
				maxlength={length}
				value={digits[i]}
				{disabled}
				class="af-input h-12 w-12 text-center text-lg font-semibold sm:h-14 sm:w-14 sm:text-xl"
				class:error={!!error}
				oninput={(e) => handleInput(i, e)}
				onkeydown={(e) => handleKeydown(i, e)}
				onpaste={handlePaste}
				onfocus={handleFocus}
				aria-label="Digit {i + 1}"
				autocomplete="one-time-code"
			/>
		{/each}
	</div>

	{#if error}
		<p class="mt-2 text-center text-sm" style="color: var(--af-color-error)" role="alert">
			{$t(error)}
		</p>
	{/if}
</div>
