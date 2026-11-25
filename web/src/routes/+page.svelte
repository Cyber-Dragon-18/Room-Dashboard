<script lang="ts">
	// Home — a grid of widgets you can rearrange.
	//   • Loads the saved layout from the server (or the default if never saved)
	//   • Edit → move / resize / remove / add widgets on a DRAFT copy
	//   • Save → sends it to the server; Cancel → throws the draft away
	//   • If another screen saves, this one reloads (live, via WebSocket)
	import { onMount } from 'svelte';
	import WidgetGrid from '#lib/widgets/WidgetGrid.svelte';
	import { widgets } from '#lib/widgets/registry.ts';
	import { defaultLayouts } from '#lib/widgets/defaults.ts';
	import type { PlacedWidget } from '#lib/widgets/types.ts';
	import { getLayout, saveLayout } from '#lib/api.ts';
	import { room, handleError } from '#lib/room.svelte.ts';

	const PAGE = 'home';

	let layout = $state<PlacedWidget[]>(defaultLayouts[PAGE]);
	let draft = $state<PlacedWidget[] | null>(null); // non-null = edit mode
	let saving = $state(false);
	let addOpen = $state(false);

	async function load() {
		try {
			const saved = await getLayout(PAGE);
			if (saved) layout = saved;
		} catch (e) {
			handleError(e);
		}
	}

	onMount(load);

	// Another screen saved this page → reload it (unless we're editing)
	$effect(() => {
		if (room.layoutUpdated.page === PAGE && !draft) load();
	});

	function startEdit() {
		draft = $state.snapshot(layout); // a plain copy: changes don't touch `layout`
	}

	function cancel() {
		draft = null;
		addOpen = false;
	}

	async function save() {
		if (!draft) return;
		saving = true;
		try {
			await saveLayout(PAGE, draft);
			layout = draft;
			draft = null;
			addOpen = false;
		} catch (e) {
			handleError(e);
		} finally {
			saving = false;
		}
	}

	function resetToDefault() {
		draft = structuredClone(defaultLayouts[PAGE]);
	}

	// Add a widget at the bottom of the page, with its default size and settings
	function add(id: string) {
		if (!draft) return;
		const def = widgets[id];
		const bottom = Math.max(0, ...draft.map((w) => w.y + w.h - 1));
		draft = [
			...draft,
			{
				key: `${id}-${Date.now()}`,
				widget: id,
				x: 1,
				y: bottom + 1,
				w: def.defaultSize.w,
				h: def.defaultSize.h,
				settings: def.defaultSettings?.() ?? {}
			}
		];
		addOpen = false;
	}
</script>

<div class="page">
	<WidgetGrid layout={draft ?? layout} editing={!!draft} onchange={(l) => (draft = l)} />
</div>

<!-- Floating toolbar, bottom-right -->
<div class="toolbar">
	{#if !draft}
		<button class="ghost" onclick={startEdit} aria-label="Edit layout">✎ Edit</button>
	{:else}
		{#if addOpen}
			<div class="add-menu">
				{#each Object.values(widgets) as def (def.id)}
					<button onclick={() => add(def.id)}>{def.name}</button>
				{/each}
			</div>
		{/if}
		<button onclick={() => (addOpen = !addOpen)}>＋ Add widget</button>
		<button class="ghost" onclick={resetToDefault}>Reset</button>
		<button class="ghost" onclick={cancel}>Cancel</button>
		<button class="primary" onclick={save} disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
	{/if}
</div>

<style>
	.page {
		padding-bottom: 5rem; /* room for the toolbar */
	}
	.toolbar {
		position: fixed;
		right: 1.5rem;
		bottom: 1.5rem;
		display: flex;
		align-items: flex-end;
		gap: 0.6rem;
	}
	.toolbar button {
		min-height: 56px;
		padding: 0 1.2rem;
		border: none;
		border-radius: 16px;
		background: var(--surface-2);
		color: var(--text);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
	}
	.toolbar .ghost {
		background: var(--surface);
		color: var(--muted);
	}
	.toolbar .primary {
		background: var(--accent);
		color: #1a1410;
	}
	.add-menu {
		position: absolute;
		right: 0;
		bottom: calc(100% + 0.6rem);
		display: grid;
		gap: 0.4rem;
		padding: 0.6rem;
		border-radius: 16px;
		background: var(--surface);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
	}
	.add-menu button {
		box-shadow: none;
		text-align: left;
	}
</style>