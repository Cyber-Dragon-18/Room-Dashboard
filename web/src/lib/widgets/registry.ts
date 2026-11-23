import type { WidgetDef } from "./types";
import Clock from './Clock.svelte';
import Scenes from './Scenes.svelte';

export const widgets: Record<string, WidgetDef> = {
    clock: {
        id: 'clock',
        name: 'Clock',
        component: Clock,
        defaultSize: {w:6, h:3}
    },

    scenes: {
        id: 'scenes',
        name: 'Scenes',
        component: Scenes,
        defaultSize: {w:6,h:2}
    }
};