<script lang="ts">
	import { onMount } from 'svelte';

	let now = $state(new Date());

	onMount(() => {
		const timer = setInterval(() => (now = new Date()), 1000);
		return () => clearInterval(timer); // cleanup
	});

	let time = $derived(now.toLocaleTimeString('pt-PT', { hour: '2-digit', minute: '2-digit' }));
	let date = $derived(
		now.toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long' })
	);
</script>

<div class="clock">
	<span class="time">{time}</span>
	<span class="date">{date}</span>
</div>

<style>
	.clock {
		height: 100%;
		display: grid;
		place-content: center;
		text-align: center;
	}
	.time {
		font-size: clamp(3rem, 10vw, 7rem); 
		font-weight: 700;
		line-height: 1;
		letter-spacing: -0.02em;
	}
	.date {
		margin-top: 0.5rem;
		color: var(--muted);
		font-size: 1.2rem;
	}
</style>