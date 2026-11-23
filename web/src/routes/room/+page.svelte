<script lang="ts">
    import { room } from '#lib/room.svelte.ts';
    import LightCard from '#lib/LightCard.svelte';
</script>

<header>
    <h1>Room</h1>
    <span class="status" class:live={room.connected}>{room.connected ? 'Live' : 'Offline'}</span>
</header>

{#if room.loading}<p class="muted">Loading...</p>{/if}

<section class="lights">
    {#each room.lights as light (light.device.id)}
        <LightCard {light}/>
    {/each}
</section>

<style>
	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}
	h1 {
		margin: 0;
	}
	.muted {
		color: var(--muted);
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
	.lights {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
	}
</style>