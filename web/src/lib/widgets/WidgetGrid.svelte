<script lang="ts">
	import { widgets } from './registry.ts';
	import type { PlacedWidget } from './types.ts';

	let { layout, columns = 12 }: { layout: PlacedWidget[]; columns?: number } = $props();
</script>

<div class="grid" style="--cols: {columns}">
	{#each layout as item (item.key)}
		{@const def = widgets[item.widget]}
		<section
			class="cell"
			style="grid-column: {item.x} / span {item.w}; grid-row: {item.y} / span {item.h};"
		>
			{#if def}
				<def.component settings={item.settings ?? {}} />
			{:else}
				<p class="missing">Unknown widget "{item.widget}"</p>
			{/if}
		</section>
	{/each}
</div>

<style>
	.grid {
		display: grid;
		grid-template-columns: repeat(var(--cols), 1fr);
		grid-auto-rows: 80px; 
		gap: 1rem;
	}
	.cell {
		background: var(--surface);
		border-radius: 20px;
		padding: 1.25rem;
		overflow: hidden;
	}
	.missing {
		color: var(--danger);
	}
</style>