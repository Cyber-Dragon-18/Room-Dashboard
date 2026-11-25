// api.ts — the only place that talks to the Go server.
// Components call these functions; they never use fetch() directly.

import { getToken, clearToken } from './token';
import type { PlacedWidget } from './widgets/types.ts';

// ---- Types (mirror the JSON your Go server sends) ----

export type Device = {
	id: number;
	name: string;
	type: string; // "light" | "socket"
	driver: string; // "tuya" | "mock" | ...
};

export type LightState = {
	on: boolean;
	mode: 'white' | 'colour'; // which LEDs are on: real white, or RGB
	brightness: number; // 0-100
	temp: number; // 0 warm - 100 cool (white mode)
	r: number; // r, g, b: colour mode
	g: number;
	b: number;
};

export type DeviceWithState = {
	device: Device;
	state: LightState;
};

export type SceneAction = {
	device_id: number;
	power: boolean | null;
	temp?: number | null;
	brightness: number | null;
	color: { r: number; g: number; b: number } | null;
};

export type Scene = {
	id: number;
	name: string;
	actions: SceneAction[];
};

// Thrown when the server says 401, so the page can show the token screen again.
export class UnauthorizedError extends Error {}

// ---- The core: every request goes through here ----

async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
	const res = await fetch(path, {
		...options,
		headers: {
			'Content-Type': 'application/json',
			Authorization: `Bearer ${getToken()}`,
			...options.headers
		}
	});

	if (res.status === 401) {
		clearToken();
		throw new UnauthorizedError('token invalid or revoked');
	}
	if (!res.ok) {
		// Your Go handlers put the reason in the body ("brightness must be 0-100", ...)
		throw new Error((await res.text()).trim() || `HTTP ${res.status}`);
	}
	return res;
}

// ---- Reading ----

export async function getDevices(): Promise<Device[]> {
	const res = await apiFetch('/api/devices');
	return res.json();
}

export async function getDevice(id: number): Promise<DeviceWithState> {
	const res = await apiFetch(`/api/devices/${id}`);
	return res.json();
}

export async function getScenes(): Promise<Scene[]> {
	const res = await apiFetch('/api/scenes');
	const data: Record<string, Scene> | Scene[] = await res.json();
	// The server returns a map keyed by name; turn it into a list sorted by id.
	const list = Array.isArray(data) ? data : Object.values(data);
	return list.sort((a, b) => a.id - b.id);
}

// ---- Commands ----

export async function setPower(id: number, on: boolean): Promise<void> {
	await apiFetch(`/api/devices/${id}/power`, {
		method: 'POST',
		body: JSON.stringify({ on })
	});
}

export async function setBrightness(id: number, pct: number): Promise<void> {
	await apiFetch(`/api/devices/${id}/brightness`, {
		method: 'POST',
		body: JSON.stringify({ pct })
	});
}

export async function setColor(id: number, r: number, g: number, b: number): Promise<void> {
	await apiFetch(`/api/devices/${id}/color`, {
		method: 'POST',
		body: JSON.stringify({ r, g, b })
	});
}

// White mode: the lamp's real white LEDs. 0 = warm, 100 = cool.
export async function setWhite(id: number, temp: number): Promise<void> {
	await apiFetch(`/api/devices/${id}/white`, {
		method: 'POST',
		body: JSON.stringify({ temp })
	});
}

export async function runScene(name: string): Promise<void> {
	await apiFetch(`/api/scenes/${encodeURIComponent(name)}/run`, { method: 'POST' });
}

// ---- Layouts ----

// Returns null if this page was never saved (the app then uses its default).
export async function getLayout(page: string): Promise<PlacedWidget[] | null> {
	try {
		const res = await apiFetch(`/api/layouts/${page}`);
		const data: { widgets: PlacedWidget[] } = await res.json();
		return data.widgets;
	} catch (e) {
		if (e instanceof Error && e.message === 'no saved layout') return null;
		throw e;
	}
}

export async function saveLayout(page: string, widgets: PlacedWidget[]): Promise<void> {
	await apiFetch(`/api/layouts/${page}`, {
		method: 'PUT',
		body: JSON.stringify({ widgets })
	});
}