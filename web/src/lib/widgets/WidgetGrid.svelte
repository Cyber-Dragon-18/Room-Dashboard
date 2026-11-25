<script lang="ts">
	// WidgetGrid — draws a page's layout: each widget in its grid cell.
	// In edit mode, every widget gets big buttons to move, resize and remove it.
	// It never saves anything itself: it hands the changed layout to `onchange`.
	import { widgets } from './registry.ts';
	import { GRID_COLUMNS, type PlacedWidget } from './types.ts';

	type Props = {
		layout: PlacedWidget[];
		editing?: boolean;
		onchange?: (layout: PlacedWidget[]) => void;
	};
	let { layout, editing = false, onchange }: Props = $props();

	// Return a NEW layout with one widget changed (never mutate the prop).
	function update(key: string, change: (w: PlacedWidget) => Partial<PlacedWidget>) {
		onchange?.(layout.map((w) => (w.key === key ? { ...w, ...change(w) } : w)));
	}

	const clamp = (v: number, min: number, max: number) => Math.min(Math.max(v, min), max);

	const move = (key: string, dx: number, dy: number) =>
		update(key, (w) => ({
			x: clamp(w.x + dx, 1, GRID_COLUMNS - w.w + 1),
			y: Math.max(1, w.y + dy)
		}));

	const resize = (key: string, dw: number, dh: number) =>
		update(key, (w) => ({
			w: clamp(w.w + dw, 1, GRID_COLUMNS - w.x + 1),
			h: clamp(w.h + dh, 1, 12)
		}));

	const remove = (key: string) => onchange?.(layout.filter((w) => w.key !== key));
</script>

<div class="grid" class:editing style="--cols: {GRID_COLUMNS}">
	{#each layout as item (item.key)}
		{@const def = widgets[item.widget]}
		<section
			class="cell"
			style="grid-column: {item.x} / span {item.w}; grid-row: {item.y} / span {item.h};"
		>
			{#if def}
				<!-- Svelte 5: a component stored in a variable is drawn like a tag -->
				<def.component settings={item.settings ?? {}} />
			{:else}
				<p class="missing">Unknown widget "{item.widget}"</p>
			{/if}

			{#if editing}
				<!-- Edit overlay: covers the widget so taps don't reach it -->
				<div class="overlay">
					<span class="name">{def?.name ?? item.widget} · {item.w}×{item.h}</span>
					<div class="controls">
						<div class="pad">
							<button class="up" aria-label="Move up" onclick={() => move(item.key, 0, -1)}>↑</button>
							<button class="left" aria-label="Move left" onclick={() => move(item.key, -1, 0)}>←</button>
							<button class="right" aria-label="Move right" onclick={() => move(item.key, 1, 0)}>→</button>
							<button class="down" aria-label="Move down" onclick={() => move(item.key, 0, 1)}>↓</button>
						</div>
						<div class="sizes">
							<button aria-label="Narrower" onclick={() => resize(item.key, -1, 0)}>W−</button>
							<button aria-label="Wider" onclick={() => resize(item.key, 1, 0)}>W+</button>
							<button aria-label="Shorter" onclick={() => resize(item.key, 0, -1)}>H−</button>
							<button aria-label="Taller" onclick={() => resize(item.key, 0, 1)}>H+</button>
						</div>
						<button class="remove" aria-label="Remove" onclick={() => remove(item.key)}>✕</button>
					</div>
				</div>
			{/if}
		</section>
	{/each}
</div>

<style>
	.grid {
		display: grid;
		grid-template-columns: repeat(var(--cols), 1fr);
		grid-auto-rows: 80px; /* one row = 80px; h: 3 = 240px tall */
		gap: 1rem;
	}
	.cell {
		position: relative;
		background: var(--surface);
		border-radius: 20px;
		padding: 1.25rem;
		overflow: hidden;
	}
	/* In edit mode, show the empty grid cells' outline so you can see where things go */
	.grid.editing {
		background-image: repeating-linear-gradient(
			to right,
			transparent 0 calc((100% - 11rem) / 12),
			color-mix(in srgb, var(--muted) 12%, transparent) 0 1rem
		);
	}
	.grid.editing .cell {
		outline: 2px dashed var(--accent);
		outline-offset: -2px;
	}
	.overlay {
		position: absolute;
		inset: 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.5rem;
		background: color-mix(in srgb, var(--bg) 80%, transparent);
		backdrop-filter: blur(2px);
	}
	.name {
		font-weight: 700;
	}
	.controls {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}
	.sizes {
		display: grid;
		grid-template-columns: repeat(2, 48px);
		gap: 0.35rem;
	}
	.pad {
		display: grid;
		grid-template-columns: repeat(3, 48px);
		grid-template-areas:
			'.    up    .'
			'left .     right'
			'.    down  .';
		gap: 0.35rem;
	}
	.pad .up { grid-area: up; }
	.pad .left { grid-area: left; }
	.pad .right { grid-area: right; }
	.pad .down { grid-area: down; }
	.overlay button {
		min-width: 48px;
		min-height: 44px; /* finger-sized */
		border: none;
		border-radius: 12px;
		background: var(--surface-2);
		color: var(--text);
		font-size: 1.05rem;
		font-weight: 700;
		cursor: pointer;
	}
	.overlay button:active {
		background: var(--accent);
		color: #1a1410;
	}
	.overlay .remove {
		background: color-mix(in srgb, var(--danger) 30%, var(--surface-2));
	}
	.missing {
		color: var(--danger);
	}
</style>