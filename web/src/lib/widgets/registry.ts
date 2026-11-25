// registry.ts — every widget the app knows about.
// Adding a widget = write its component + add ONE entry here.

import type { WidgetDef } from './types.ts';
import Clock from './Clock.svelte';
import Scenes from './Scenes.svelte';
import LightWidget from './LightWidget.svelte';
import Weather from './Weather.svelte';
import { room } from '#lib/room.svelte.ts';

export const widgets: Record<string, WidgetDef> = {
	clock: {
		id: 'clock',
		name: 'Clock',
		component: Clock,
		defaultSize: { w: 7, h: 3 }
	},
	scenes: {
		id: 'scenes',
		name: 'Scenes',
		component: Scenes,
		defaultSize: { w: 5, h: 2 }
	},
	light: {
		id: 'light',
		name: 'Light', // settings: { deviceId }
		component: LightWidget,
		defaultSize: { w: 5, h: 3 },
		// a new light widget controls the first light (change it in the layout later)
		defaultSettings: () => ({ deviceId: room.lights[0]?.device.id })
	},
	weather: {
		id: 'weather',
		name: 'Weather', // settings (optional): { name, lat, lon }
		component: Weather,
		defaultSize: { w: 7, h: 2 }
	}
};