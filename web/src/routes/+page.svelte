<script lang="ts">
	import WidgetGrid from '#lib/widgets/WidgetGrid.svelte';
	import type { PlacedWidget } from '#lib/widgets/types.ts';

	const layout: PlacedWidget[] = [
		{ key: 'clock-1', widget: 'clock', x: 1, y: 1, w: 7, h: 4 },
		{ key: 'scenes-1', widget: 'scenes', x: 8, y: 1, w: 5, h: 2 }
	];
</script>

<WidgetGrid {layout} />