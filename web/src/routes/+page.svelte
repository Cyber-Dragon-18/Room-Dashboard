<script lang="ts">
	import { saveToken } from '#lib/token.ts';
	import { runScene } from '#lib/api.ts';
	import { room, loadAll, handleError } from '#lib/room.svelte.ts';
	import LightCard from '#lib/LightCard.svelte';

	let input = $state('');

	function save() {
		if (!input.trim()) return;
		saveToken(input);
		input = '';
		room.needToken = false;
		loadAll();
	}

	async function scene(name: string) {
		try {
			await runScene(name);
		} catch (e) {
			handleError(e);
		}
	}

	// If we already have a token, load everything when the page opens
	if (!room.needToken) loadAll();
</script>

{#if room.needToken}
	<main class="narrow">
		<h1>Room Dashboard</h1>
		<p class="muted">Paste a device token to connect.</p>
		<p class="muted small">Create one with: <code>go run . token create browser</code></p>

		<form onsubmit={(e) => { e.preventDefault(); save(); }}>
			<input bind:value={input} placeholder="rd_..." autocomplete="off" />
			<button type="submit">Save</button>
		</form>
		{#if room.error}<p class="error">{room.error}</p>{/if}
	</main>
{:else}
	<main>
		<header class="top">
			<h1>My Room</h1>
			<span class="status" class:live={room.connected}>
				{room.connected ? 'Live' : 'Offline'}
			</span>
		</header>

		{#if room.error}<p class="error">{room.error}</p>{/if}
		{#if room.loading}<p class="muted">Loading…</p>{/if}

		<section class="scenes">
			{#each room.scenes as s (s.id)}
				<button onclick={() => scene(s.name)}>{s.name}</button>
			{/each}
		</section>

		<section class="lights">
			{#each room.lights as light (light.device.id)}
				<LightCard {light} />
			{/each}
		</section>
	</main>
{/if}

<style>
	/* Cozy dark theme — all colors in one place */
	:global(:root) {
		--bg: #17130f;
		--surface: #221c17;
		--surface-2: #2e2620;
		--text: #f3e9dc;
		--muted: #a8998a;
		--accent: #f2a65a;
		--danger: #e5736b;
		--ok: #8fcf8a;
	}
	:global(body) {
		margin: 0;
		background: var(--bg);
		color: var(--text);
		font-family: 'Nunito', system-ui, sans-serif;
	}
	main {
		max-width: 900px;
		margin: 0 auto;
		padding: 1.5rem;
	}
	.narrow {
		max-width: 480px;
		margin-top: 15vh;
	}
	.top {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	h1 {
		margin: 0 0 1rem;
	}
	.muted {
		color: var(--muted);
	}
	.small {
		font-size: 0.9rem;
	}
	.error {
		color: var(--danger);
	}
	.status {
		padding: 0.35rem 0.9rem;
		border-radius: 999px;
		background: var(--surface);
		color: var(--muted);
		font-size: 0.9rem;
	}
	.status.live {
		color: var(--ok);
	}
	.status.live::before {
		content: '● ';
	}
	form {
		display: flex;
		gap: 0.5rem;
	}
	input {
		flex: 1;
		padding: 0.8rem;
		font-size: 1rem;
		border-radius: 12px;
		border: 1px solid var(--surface-2);
		background: var(--surface);
		color: var(--text);
	}
	button {
		min-height: 56px;
		padding: 0 1.4rem;
		border: none;
		border-radius: 16px;
		background: var(--surface-2);
		color: var(--text);
		font-size: 1.05rem;
		font-weight: 600;
		cursor: pointer;
		text-transform: capitalize;
	}
	.scenes {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
		margin-bottom: 1.5rem;
	}
	.lights {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
	}
</style>