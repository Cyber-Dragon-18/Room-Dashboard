<script lang="ts">
	// LightCard — controls for one light, with two tabs like the Hama app:
	//   White  → the lamp's real white LEDs (temperature + brightness)
	//   Colour → the RGB LEDs (color picker + brightness)
	// It doesn't change its own state: it calls the API, and the WebSocket
	// event updates `room`, which updates this card (and every other screen).
	import { setPower, setBrightness, setColor, setWhite } from './api.ts';
	import {
		handleError,
		refreshLight,
		rgbToHex,
		hexToRgb,
		tempToHex,
		type Light
	} from './room.svelte.ts';

	let { light }: { light: Light } = $props();

	// The tab always shows the lamp's real mode (changed by you, a scene, or another screen)
	let tab = $derived(light.state.mode);

	let colorHex = $derived(rgbToHex(light.state.r, light.state.g, light.state.b));
	// The glow shows what the lamp really looks like right now
	let glow = $derived(light.state.mode === 'white' ? tempToHex(light.state.temp) : colorHex);
	let busy = $state(false);

	// While you drag a slider, show your value; when you let go, go back to
	// showing the real state. This keeps the number and the slider in sync,
	// and if a command fails the slider jumps back to the real value.
	let dragBrightness = $state<number | null>(null);
	let dragTemp = $state<number | null>(null);
	let shownBrightness = $derived(dragBrightness ?? light.state.brightness);
	let shownTemp = $derived(dragTemp ?? light.state.temp);

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

	const toggle = () => run(() => setPower(light.device.id, !light.state.on));

	// Pressing a tab switches the lamp right away (like the Hama app),
	// then re-reads its state, because each mode has its own brightness.
	function switchMode(mode: 'white' | 'colour') {
		if (mode === light.state.mode) return;
		const { id } = light.device;
		run(async () => {
			if (mode === 'white') {
				await setWhite(id, light.state.temp);
			} else {
				await setColor(id, light.state.r, light.state.g, light.state.b);
			}
			refreshLight(id);
		});
	}

	function value(e: Event): number {
		return Number((e.target as HTMLInputElement).value);
	}

	// oninput: update the number live while dragging (no command yet)
	// onchange: ONE command when you let go
	async function commitBrightness(e: Event) {
		const pct = value(e);
		await run(() => setBrightness(light.device.id, pct));
		dragBrightness = null;
	}
	async function commitTemp(e: Event) {
		const temp = value(e);
		await run(() => setWhite(light.device.id, temp));
		dragTemp = null;
	}

	function onColor(e: Event) {
		const { r, g, b } = hexToRgb((e.target as HTMLInputElement).value);
		run(() => setColor(light.device.id, r, g, b));
	}
</script>

<article class="card" class:on={light.state.on} style="--glow: {glow}">
	<header>
		<div>
			<h3>{light.device.name}</h3>
			<span class="sub">{light.device.driver}</span>
		</div>
		<button class="power" onclick={toggle} disabled={busy} aria-pressed={light.state.on}>
			{light.state.on ? 'On' : 'Off'}
		</button>
	</header>

	<div class="tabs" role="tablist">
		<button
			role="tab"
			aria-selected={tab === 'white'}
			class:active={tab === 'white'}
			onclick={() => switchMode('white')}
			disabled={busy || !light.state.on}
		>
			White
		</button>
		<button
			role="tab"
			aria-selected={tab === 'colour'}
			class:active={tab === 'colour'}
			onclick={() => switchMode('colour')}
			disabled={busy || !light.state.on}
		>
			Colour
		</button>
	</div>

	{#if tab === 'white'}
		<label>
			<span>Temperature <b>{shownTemp < 34 ? 'Warm' : shownTemp > 66 ? 'Cool' : 'Neutral'}</b></span>
			<input
				class="temp"
				type="range"
				min="0"
				max="100"
				value={shownTemp}
				oninput={(e) => (dragTemp = value(e))}
				onchange={commitTemp}
				disabled={busy || !light.state.on}
			/>
		</label>
	{:else}
		<label class="color">
			<span>Colour</span>
			<input type="color" value={colorHex} onchange={onColor} disabled={busy || !light.state.on} />
		</label>
	{/if}

	<label>
		<span>Brightness <b>{shownBrightness}%</b></span>
		<input
			type="range"
			min="0"
			max="100"
			value={shownBrightness}
			oninput={(e) => (dragBrightness = value(e))}
			onchange={commitBrightness}
			disabled={busy || !light.state.on}
		/>
	</label>
</article>

<style>
	.card {
		background: var(--surface);
		border-radius: 20px;
		padding: 1.25rem;
		display: grid;
		gap: 1.25rem;
		border: 1px solid transparent;
		transition: border-color 0.3s, box-shadow 0.3s;
	}
	.card.on {
		border-color: color-mix(in srgb, var(--glow) 40%, transparent);
		box-shadow: 0 0 40px -12px var(--glow);
	}
	header {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	h3 {
		margin: 0;
		font-size: 1.2rem;
	}
	.sub {
		color: var(--muted);
		font-size: 0.85rem;
	}
	.power {
		min-width: 88px;
		min-height: 56px; /* big touch target */
		border-radius: 16px;
		border: none;
		font-size: 1.1rem;
		font-weight: 600;
		background: var(--surface-2);
		color: var(--text);
		cursor: pointer;
	}
	.card.on .power {
		background: var(--accent);
		color: #1a1410;
	}
	.tabs {
		display: grid;
		grid-template-columns: 1fr 1fr;
		background: var(--surface-2);
		border-radius: 14px;
		padding: 4px;
	}
	.tabs button {
		min-height: 44px;
		border: none;
		border-radius: 10px;
		background: none;
		color: var(--muted);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
	}
	.tabs button.active {
		background: var(--surface);
		color: var(--text);
	}
	label {
		display: grid;
		gap: 0.5rem;
		color: var(--muted);
	}
	label b {
		color: var(--text);
	}
	input[type='range'] {
		width: 100%;
		height: 40px;
		accent-color: var(--accent);
	}
	/* Temperature slider: warm on the left, cool on the right */
	input.temp {
		appearance: none;
		height: 14px;
		border-radius: 999px;
		background: linear-gradient(to right, #ffb35c, #fff4e0, #cfe8ff);
		margin: 13px 0;
	}
	input.temp::-webkit-slider-thumb {
		appearance: none;
		width: 30px;
		height: 30px;
		border-radius: 50%;
		background: var(--text);
		border: 3px solid var(--surface);
		cursor: pointer;
	}
	input.temp::-moz-range-thumb {
		width: 26px;
		height: 26px;
		border-radius: 50%;
		background: var(--text);
		border: 3px solid var(--surface);
		cursor: pointer;
	}
	.color {
		grid-template-columns: 1fr auto;
		align-items: center;
	}
	input[type='color'] {
		width: 64px;
		height: 48px;
		border: none;
		border-radius: 12px;
		background: none;
		cursor: pointer;
	}
	input:disabled,
	button:disabled {
		opacity: 0.5;
	}
</style>