// types.ts — the "contract" every widget follows.

import type { Component } from 'svelte';

// What the registry knows about a widget type.
export type WidgetDef = {
	id: string; // unique key, e.g. "clock"
	name: string; // shown in the "Add widget" menu
	component: Component<any>; // the Svelte component that draws it
	defaultSize: { w: number; h: number }; // in grid cells
	defaultSettings?: () => Record<string, unknown>; // settings for a newly added one
};


export type PlacedWidget = {
	key: string; // unique on the page (you could place two clocks)
	widget: string; // id from the registry
	x: number; // column, starting at 1
	y: number; // row, starting at 1
	w: number; // width in columns
	h: number; // height in rows
	settings?: Record<string, unknown>; // per-widget options (e.g. which light)
};

export const GRID_COLUMNS = 12;