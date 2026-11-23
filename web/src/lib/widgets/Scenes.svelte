<script lang="ts">
	import { runScene } from '#lib/api.ts';
	import { room, handleError } from '#lib/room.svelte.ts';

	async function run(name: string) {
		try {
			await runScene(name);
		} catch (e) {
			handleError(e);
		}
	}
</script>

<div class="scenes">
	<h3>Scenes</h3>
	<div class="buttons">
		{#each room.scenes as s (s.id)}
			<button onclick={() => run(s.name)}>{s.name}</button>
		{/each}
	</div>
</div>

<style>
	.scenes {
		height: 100%;
		display: flex;
		flex-direction: column;
	}
	h3 {
		margin: 0 0 0.75rem;
		color: var(--muted);
		font-size: 1rem;
		font-weight: 600;
	}
	.buttons {
		flex: 1;
		display: grid;
		grid-auto-flow: column;
		gap: 0.75rem;
	}
	button {
		min-height: 64px;
		border: none;
		border-radius: 16px;
		background: var(--surface-2);
		color: var(--text);
		font-size: 1.15rem;
		font-weight: 600;
		text-transform: capitalize;
		cursor: pointer;
	}
	button:active {
		background: var(--accent);
		color: #1a1410;
	}
</style>