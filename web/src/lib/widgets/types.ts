import type { Component } from "svelte";

export type WidgetDef = {
    id: string;
    name: string;
    component: Component<any>;
    defaultSize: {w: number; h: number};
};

export type PlacedWidget = {
    key: string;
    widget: string;
    x: number;
    y: number;
    w: number;
    h: number;
    settings?: Record<string, unknown>;
};