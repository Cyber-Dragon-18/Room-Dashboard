// room.svelte.ts — the shared, live state of the room.
//
// The ".svelte.ts" ending lets this file use runes ($state) outside a component.
// Every component that imports `room` sees the same data, and updates
// automatically when it changes (from a button OR from a WebSocket message).

import { getToken } from './token.ts';
import {
	getDevices,
	getDevice,
	getScenes,
	UnauthorizedError,
	type Device,
	type LightState,
	type Scene
} from './api.ts';

export type Light = { device: Device; state: LightState };

// One shared object. Mutate its fields (room.lights = ...), never reassign `room` itself.
export const room = $state({
	lights: [] as Light[],
	scenes: [] as Scene[],
	connected: false, // is the WebSocket open?
	needToken: !getToken(), // true → show the token screen
	error: '',
	loading: false,
	// Bumped when another screen saves a layout; pages watch it and reload.
	layoutUpdated: { page: '', count: 0 }
});

// ---------------------------------------------------------------------------
// Loading
// ---------------------------------------------------------------------------

export async function loadAll(): Promise<void> {
	room.loading = true;
	room.error = '';
	try {
		const devices = await getDevices();
		// Fetch every light's state in parallel (Promise.all waits for all of them)
		room.lights = await Promise.all(
			devices.filter((d) => d.type === 'light').map((d) => getDevice(d.id))
		);
		room.scenes = await getScenes();
		connect(); // start live updates once we know the token works
	} catch (e) {
		handleError(e);
	} finally {
		room.loading = false;
	}
}

// Reload only the states (used after a scene, since it changes several things at once)
async function refreshStates(): Promise<void> {
	try {
		room.lights = await Promise.all(room.lights.map((l) => getDevice(l.device.id)));
	} catch (e) {
		handleError(e);
	}
}

// Re-read ONE light's real state from the server after a delay.
// Used after switching White/Colour: each mode has its own brightness on the
// lamp, and Tuya needs a moment before it reports the new values.
export function refreshLight(id: number, delayMs = 2500): void {
	setTimeout(async () => {
		try {
			const fresh = await getDevice(id);
			const light = room.lights.find((l) => l.device.id === id);
			if (light) light.state = fresh.state;
		} catch (e) {
			handleError(e);
		}
	}, delayMs);
}

// Every component action goes through this, so errors are handled in one place.
export function handleError(e: unknown): void {
	if (e instanceof UnauthorizedError) {
		disconnect();
		room.needToken = true;
		room.error = 'Token invalid or revoked. Paste a new one.';
	} else {
		room.error = e instanceof Error ? e.message : String(e);
	}
}

// ---------------------------------------------------------------------------
// WebSocket: live updates
// ---------------------------------------------------------------------------

let ws: WebSocket | null = null;
let retryDelay = 1000; // grows 1s → 2s → 4s ... up to 30s
let retryTimer: ReturnType<typeof setTimeout> | undefined;
let stopped = false;

type ServerEvent = {
	type: string;
	device_id?: number;
	name?: string;
	changes?: {
		on?: boolean;
		mode?: 'white' | 'colour';
		temp?: number;
		brightness?: number;
		color?: { r: number; g: number; b: number };
	};
};

export function connect(): void {
	if (ws) return; // already connected or connecting
	stopped = false;

	// Same host as the page; the Vite proxy forwards /ws to Go in development.
	const proto = location.protocol === 'https:' ? 'wss' : 'ws';
	const token = encodeURIComponent(getToken() ?? '');
	ws = new WebSocket(`${proto}://${location.host}/ws?token=${token}`);

	ws.onopen = () => {
		room.connected = true;
		retryDelay = 1000; // reset the backoff after a good connection
	};

	ws.onmessage = (msg) => {
		const event: ServerEvent = JSON.parse(msg.data);
		applyEvent(event);
	};

	ws.onclose = () => {
		room.connected = false;
		ws = null;
		if (stopped) return;
		// Try again later, waiting a bit longer each time
		retryTimer = setTimeout(connect, retryDelay);
		retryDelay = Math.min(retryDelay * 2, 30_000);
	};
}

export function disconnect(): void {
	stopped = true;
	clearTimeout(retryTimer);
	ws?.close();
	ws = null;
	room.connected = false;
}

function applyEvent(event: ServerEvent): void {
	if (event.type === 'device.updated' && event.device_id && event.changes) {
		const light = room.lights.find((l) => l.device.id === event.device_id);
		if (!light) return;
		const c = event.changes;
		// Merge only what changed — the rest of the state stays as it was
		if (c.on !== undefined) light.state.on = c.on;
		if (c.mode !== undefined) light.state.mode = c.mode;
		if (c.temp !== undefined) light.state.temp = c.temp;
		if (c.brightness !== undefined) light.state.brightness = c.brightness;
		if (c.color) {
			light.state.r = c.color.r;
			light.state.g = c.color.g;
			light.state.b = c.color.b;
		}
	}

	if (event.type === 'layout.updated' && event.name) {
		room.layoutUpdated = { page: event.name, count: room.layoutUpdated.count + 1 };
	}

	if (event.type === 'scene.ran') {
		// A scene changed several things; Tuya needs a moment to report them.
		setTimeout(refreshStates, 3000);
	}
}

// ---------------------------------------------------------------------------
// Color helpers for <input type="color">, which works with "#rrggbb"
// ---------------------------------------------------------------------------

export function rgbToHex(r: number, g: number, b: number): string {
	return '#' + [r, g, b].map((v) => v.toString(16).padStart(2, '0')).join('');
}

// The color white light "looks like" at a temperature (for the card's glow):
// 0 = warm amber, 100 = cool blue-white.
export function tempToHex(temp: number): string {
	const warm = [255, 179, 92];
	const cool = [207, 232, 255];
	const t = temp / 100;
	const [r, g, b] = warm.map((w, i) => Math.round(w + (cool[i] - w) * t));
	return rgbToHex(r, g, b);
}

export function hexToRgb(hex: string): { r: number; g: number; b: number } {
	const n = parseInt(hex.slice(1), 16);
	return { r: (n >> 16) & 255, g: (n >> 8) & 255, b: n & 255 };
}