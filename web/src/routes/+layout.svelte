<script lang="ts">
	import favicon from '#lib/assets/favicon.svg';
	import { page } from '$app/state';
	import type { LayoutProps } from './$types';
	import { saveToken } from '#lib/token.ts';
	import { room, loadAll } from '#lib/room.svelte.ts';

	let { children }: LayoutProps = $props();

	// --- moved here from +page.svelte: token screen + loading ---
	let input = $state('');

	function save() {
		if (!input.trim()) return;
		saveToken(input);
		input = '';
		room.needToken = false;
		loadAll();
	}

	// Load once, whichever page you open first
	if (!room.needToken) loadAll();

	const links = [
		{ href: '/', label: 'Home', icon: '⌂' },
		{ href: '/room', label: 'Room', icon: '💡' },
		{ href: '/planner', label: 'Planner', icon: '▦' }
	];
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if room.needToken}
	<div class="login">
		<h1>Room Dashboard</h1>
		<p class="muted">Paste a device token to connect.</p>
		<form onsubmit={(e) => { e.preventDefault(); save(); }}>
			<input bind:value={input} placeholder="rd_..." autocomplete="off" />
			<button type="submit">Save</button>
		</form>
		{#if room.error}<p class="error">{room.error}</p>{/if}
	</div>
{:else}
<div class="app">
	<nav>
		{#each links as link (link.href)}
			<a href={link.href} class:active={page.url.pathname === link.href}>
				<span class="icon">{link.icon}</span>
				<span class="label">{link.label}</span>
			</a>
		{/each}
	</nav>

	<main>
		{#if room.error}<p class="error">{room.error}</p>{/if}
		{@render children()}
	</main>
</div>
{/if}

<style>
	:global(:root) {
		--bg: #17130f;
		--surface: #221c17;
		--text: #f3e9dc;
		--muted: #a8998a;
		--accent: #f2a65a;
		--surface-2: #2e2620;
		--danger: #e5736b;
		--ok: #8fcf8a;
	}
	:global(body) {
		margin: 0;
		background: var(--bg);
		color: var(--text);
		font-family: 'Nunito', system-ui, sans-serif;
	}
	.app {
		display: grid;
		grid-template-columns: 96px 1fr;
		height: 100vh;
	}
	nav {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 1rem 0.5rem;
		background: var(--surface);
	}
	a {
		display: grid;
		place-items: center;
		gap: 0.25rem;
		min-height: 72px;
		border-radius: 16px;
		color: var(--muted);
		text-decoration: none;
	}
	a.active {
		background: color-mix(in srgb, var(--accent) 18%, transparent);
		color: var(--accent);
	}
	.icon {
		font-size: 1.6rem;
	}
	.label {
		font-size: 0.8rem;
	}
	.login {
		max-width: 480px;
		margin: 15vh auto 0;
		padding: 0 1rem;
	}
	.login form {
		display: flex;
		gap: 0.5rem;
	}
	.login input {
		flex: 1;
		padding: 0.8rem;
		font-size: 1rem;
		border-radius: 12px;
		border: 1px solid var(--surface-2);
		background: var(--surface);
		color: var(--text);
	}
	.login button {
		min-height: 56px;
		padding: 0 1.4rem;
		border: none;
		border-radius: 16px;
		background: var(--accent);
		color: #1a1410;
		font-weight: 600;
		font-size: 1rem;
	}
	.muted {
		color: var(--muted);
	}
	.error {
		color: var(--danger);
	}
	main {
		overflow-y: auto;
		padding: 1.5rem;
	}
</style>