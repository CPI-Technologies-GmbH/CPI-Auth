<script lang="ts">
	import AccountLayout from '$components/AccountLayout.svelte';
	import type { Snippet } from 'svelte';

	let { data, children }: { data: { user: import('$lib/types').User }; children: Snippet } = $props();
</script>

<AccountLayout user={data.user}>
	{@render children()}
</AccountLayout>
