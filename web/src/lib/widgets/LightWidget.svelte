<script lang="ts">

	import { setPower, setBrightness } from '#lib/api.ts';
	import { room, handleError, rgbToHex, tempToHex } from '#lib/room.svelte.ts';

	let { settings }: { settings: { deviceId?: number } } = $props();

	// $derived: finds the light again whenever the data loads or changes
	let light = $derived(room.lights.find((l) => l.device.id === settings.deviceId));

	let glow = $derived(
		!light
			? 'transparent'
			: light.state.mode === 'white'
				? tempToHex(light.state.temp)
				: rgbToHex(light.state.r, light.state.g, light.state.b)
	);

	let busy = $state(false);
	let dragBrightness = $state<number | null>(null);
	let shownBrightness = $derived(dragBrightness ?? light?.state.brightness ?? 0);

	async function run(action: () => Promise<void>) {
		busy = true;
		try {
			await action();
		} catch (e) {
			handleError(e);
		} finally {
			busy = false;
		}
	}

	function toggle() {
		if (!light) return;
		const { id } = light.device;
		const on = !light.state.on;
		run(() => setPower(id, on));
	}

	async function commitBrightness(e: Event) {
		if (!light) return;
		const { id } = light.device;
		const pct = Number((e.target as HTMLInputElement).value);
		await run(() => setBrightness(id, pct));
		dragBrightness = null;
	}
</script>

{#if !light}
	<p class="muted">
		{room.loading ? 'Loading…' : `Light ${settings.deviceId ?? '?'} not found`}
	</p>
{:else}
	<div class="light" class:on={light.state.on} style="--glow: {glow}">
		<button class="power" onclick={toggle} disabled={busy} aria-pressed={light.state.on}>
			<span class="bulb"></span>
			<span class="text">
				<b>{light.device.name}</b>
				<small>{light.state.on ? `On · ${shownBrightness}%` : 'Off'}</small>
			</span>
		</button>

		<input
			type="range"
			min="0"
			max="100"
			aria-label="Brightness"
			value={shownBrightness}
			oninput={(e) => (dragBrightness = Number((e.target as HTMLInputElement).value))}
			onchange={commitBrightness}
			disabled={busy || !light.state.on}
		/>
	</div>
{/if}

<style>
	.light {
		height: 100%;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		gap: 0.75rem;
	}
	.power {
		display: flex;
		align-items: center;
		gap: 1rem;
		width: 100%;
		min-height: 72px;
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 16px;
		background: var(--surface-2);
		color: var(--text);
		text-align: left;
		cursor: pointer;
	}
	.bulb {
		width: 40px;
		height: 40px;
		flex-shrink: 0;
		border-radius: 50%;
		background: var(--surface);
		border: 2px solid var(--muted);
		transition: background 0.3s, box-shadow 0.3s;
	}
	.light.on .bulb {
		background: var(--glow);
		border-color: var(--glow);
		box-shadow: 0 0 24px var(--glow);
	}
	.text {
		display: grid;
	}
	.text b {
		font-size: 1.15rem;
	}
	.text small {
		color: var(--muted);
		font-size: 0.95rem;
	}
	input[type='range'] {
		width: 100%;
		height: 40px;
		accent-color: var(--accent);
	}
	input:disabled,
	button:disabled {
		opacity: 0.5;
	}
	.muted {
		color: var(--muted);
	}
</style>