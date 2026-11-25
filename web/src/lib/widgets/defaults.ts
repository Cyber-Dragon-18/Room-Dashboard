// defaults.ts — the layout each page starts with, before you save your own.
// Once you press Save in edit mode, the saved layout (in the database) wins.

import type { PlacedWidget } from './types.ts';

export const defaultLayouts: Record<string, PlacedWidget[]> = {
	home: [
		// left column: time and weather
		{ key: 'clock-1', widget: 'clock', x: 1, y: 1, w: 7, h: 3 },
		{ key: 'weather-1', widget: 'weather', x: 1, y: 4, w: 7, h: 2 },
		// right column: room controls
		{ key: 'scenes-1', widget: 'scenes', x: 8, y: 1, w: 5, h: 2 },
		{ key: 'desk-1', widget: 'light', x: 8, y: 3, w: 5, h: 3, settings: { deviceId: 1 } }
	]
};