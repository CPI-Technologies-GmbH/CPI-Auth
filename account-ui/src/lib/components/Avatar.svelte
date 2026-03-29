<script lang="ts">
	import { getInitials } from '$lib/utils/format';

	let {
		src,
		name,
		email,
		size = 40,
		editable = false,
		onupload
	}: {
		src?: string | null;
		name?: string;
		email?: string;
		size?: number;
		editable?: boolean;
		onupload?: (file: File) => void;
	} = $props();

	let fileInput = $state<HTMLInputElement>();
	let imageError = $state(false);

	function handleFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (file && onupload) {
			onupload(file);
		}
	}

	function handleClick() {
		if (editable) {
			fileInput?.click();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (editable && (event.key === 'Enter' || event.key === ' ')) {
			event.preventDefault();
			fileInput?.click();
		}
	}

	const initials = $derived(getInitials(name, email));
	const fontSize = $derived(Math.max(12, size * 0.35));
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
	class="avatar"
	class:editable
	style="width: {size}px; height: {size}px;"
	onclick={handleClick}
	onkeydown={handleKeydown}
	role={editable ? 'button' : 'img'}
	tabindex={editable ? 0 : -1}
	aria-label={editable ? 'Change avatar' : `Avatar for ${name || email || 'user'}`}
>
	{#if src && !imageError}
		<img
			{src}
			alt=""
			class="avatar-img"
			onerror={() => (imageError = true)}
		/>
	{:else}
		<span class="avatar-initials" style="font-size: {fontSize}px;">
			{initials}
		</span>
	{/if}

	{#if editable}
		<div class="avatar-overlay">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<path d="M23 19a2 2 0 01-2 2H3a2 2 0 01-2-2V8a2 2 0 012-2h4l2-3h6l2 3h4a2 2 0 012 2z" />
				<circle cx="12" cy="13" r="4" />
			</svg>
		</div>
		<input
			bind:this={fileInput}
			type="file"
			accept="image/*"
			class="sr-only"
			onchange={handleFileChange}
		/>
	{/if}
</div>

<style>
	.avatar {
		position: relative;
		border-radius: 50%;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
		flex-shrink: 0;
	}

	.avatar.editable {
		cursor: pointer;
	}

	.avatar-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.avatar-initials {
		color: white;
		font-weight: 600;
		user-select: none;
	}

	.avatar-overlay {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		background-color: rgba(0, 0, 0, 0.5);
		opacity: 0;
		transition: opacity 0.2s;
	}

	.avatar.editable:hover .avatar-overlay,
	.avatar.editable:focus-visible .avatar-overlay {
		opacity: 1;
	}

	.avatar-overlay svg {
		width: 40%;
		height: 40%;
		color: white;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border-width: 0;
	}
</style>
