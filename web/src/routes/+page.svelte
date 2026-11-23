<script lang="ts">
	import { runScene } from "#lib/api.ts";
	import { room, handleError} from "#lib/room.svelte.ts";

	async function scene(name: string) {
		try {
			await runScene(name);
		} catch(e) {
			handleError(e);
		}
	}
</script>


<h1>Home</h1>
<p class="muted">Clock and weather will go here</p>

<section class="scenes">
	{#each room.scenes as s (s.id)}
		<button onclick={() => scene(s.name)}>{s.name}</button>
	{/each}
</section>

<style>
	h1 {
		margin: 0 0 0.5rem;
	}

	.muted {
		color: var(--muted);
	}

	.scenes {
		display: flex;
		gap: 0.75rem;
		flex-wrap: wrap;
		margin-top: 1.5rem;
	}

	button {
		min-height: 72px;
		min-width: 140px;
		padding: 0 1.5rem;
		border: none;
		border-radius: 20px;
		background: var(--surfer);
		color : var(--text);
		font-size: 1.15rem;
		font-weight: 600;
		text-transform: capitalize;
		cursor: pointer;
	}
</style>