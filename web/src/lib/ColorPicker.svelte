<script lang="ts">
	// ColorPicker — a touch-friendly colour picker for the tablet:
	//   • big preset swatches (one tap)
	//   • a rainbow slider for any hue
	//   • an intensity slider (pastel → full colour)
	// Like the other sliders, it sends ONE command when you let go.

	type Props = {
		r: number;
		g: number;
		b: number;
		disabled?: boolean;
		onpick: (r: number, g: number, b: number) => void;
	};
	let { r, g, b, disabled = false, onpick }: Props = $props();

	// --- colour maths (HSV with h 0-360, s and v 0-1) ---
	function rgbToHsv(r: number, g: number, b: number) {
		const [rf, gf, bf] = [r / 255, g / 255, b / 255];
		const max = Math.max(rf, gf, bf);
		const min = Math.min(rf, gf, bf);
		const d = max - min;
		let h = 0;
		if (d !== 0) {
			if (max === rf) h = 60 * (((gf - bf) / d) % 6);
			else if (max === gf) h = 60 * ((bf - rf) / d + 2);
			else h = 60 * ((rf - gf) / d + 4);
		}
		if (h < 0) h += 360;
		return { h: Math.round(h), s: max === 0 ? 0 : d / max };
	}

	function hsvToRgb(h: number, s: number): [number, number, number] {
		const c = s; // v = 1: brightness is its own slider
		const x = c * (1 - Math.abs(((h / 60) % 2) - 1));
		const m = 1 - c;
		const [rf, gf, bf] =
			h < 60 ? [c, x, 0] : h < 120 ? [x, c, 0] : h < 180 ? [0, c, x]
			: h < 240 ? [0, x, c] : h < 300 ? [x, 0, c] : [c, 0, x];
		return [rf, gf, bf].map((v) => Math.round((v + m) * 255)) as [number, number, number];
	}

	const css = (h: number, s: number) => {
		const [r, g, b] = hsvToRgb(h, s);
		return `rgb(${r}, ${g}, ${b})`;
	};

	// The lamp's current hue/intensity, unless you're dragging a slider
	let current = $derived(rgbToHsv(r, g, b));
	let dragHue = $state<number | null>(null);
	let dragSat = $state<number | null>(null);
	let hue = $derived(dragHue ?? current.h);
	let sat = $derived(dragSat ?? Math.round(current.s * 100));

	const presets = [
		{ name: 'Amber', h: 30, s: 100 },
		{ name: 'Red', h: 0, s: 100 },
		{ name: 'Pink', h: 330, s: 70 },
		{ name: 'Purple', h: 275, s: 85 },
		{ name: 'Blue', h: 225, s: 100 },
		{ name: 'Cyan', h: 185, s: 100 },
		{ name: 'Green', h: 120, s: 90 },
		{ name: 'Yellow', h: 55, s: 100 }
	];

	function pick(h: number, sPct: number) {
		const [r, g, b] = hsvToRgb(h, sPct / 100);
		onpick(r, g, b);
	}

	// Is this swatch (roughly) the lamp's current colour?
	function isCurrent(p: { h: number; s: number }) {
		const dh = Math.min(Math.abs(p.h - current.h), 360 - Math.abs(p.h - current.h));
		return dh < 8 && Math.abs(p.s - current.s * 100) < 12;
	}

	function value(e: Event) {
		return Number((e.target as HTMLInputElement).value);
	}
</script>

<div class="picker" class:disabled>
	<div class="swatches">
		{#each presets as p (p.name)}
			<button
				class="swatch"
				class:selected={isCurrent(p)}
				style="background: {css(p.h, p.s / 100)}"
				aria-label={p.name}
				title={p.name}
				{disabled}
				onclick={() => pick(p.h, p.s)}
			></button>
		{/each}
	</div>

	<label>
		<span>Colour</span>
		<input
			class="hue"
			type="range"
			min="0"
			max="359"
			value={hue}
			style="--thumb: {css(hue, sat / 100)}"
			{disabled}
			oninput={(e) => (dragHue = value(e))}
			onchange={(e) => {
				pick(value(e), sat);
				dragHue = null;
			}}
		/>
	</label>

	<label>
		<span>Intensity <b>{sat}%</b></span>
		<input
			class="sat"
			type="range"
			min="10"
			max="100"
			value={sat}
			style="--full: {css(hue, 1)}; --thumb: {css(hue, sat / 100)}"
			{disabled}
			oninput={(e) => (dragSat = value(e))}
			onchange={(e) => {
				pick(hue, value(e));
				dragSat = null;
			}}
		/>
	</label>
</div>

<style>
	.picker {
		display: grid;
		gap: 1.25rem;
	}
	.picker.disabled {
		opacity: 0.5;
	}
	.swatches {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 0.75rem;
		justify-items: center;
	}
	.swatch {
		width: 52px;
		height: 52px; /* big enough for a finger */
		border-radius: 50%;
		border: 3px solid transparent;
		cursor: pointer;
		transition: transform 0.15s;
	}
	.swatch:active {
		transform: scale(0.92);
	}
	.swatch.selected {
		border-color: var(--text);
		box-shadow: 0 0 0 3px var(--surface);
	}
	label {
		display: grid;
		gap: 0.5rem;
		color: var(--muted);
	}
	label b {
		color: var(--text);
	}

	/* Big gradient sliders with a fat thumb, easy to grab on a tablet */
	input[type='range'] {
		appearance: none;
		width: 100%;
		height: 18px;
		margin: 11px 0;
		border-radius: 999px;
		cursor: pointer;
	}
	.hue {
		background: linear-gradient(
			to right,
			#f00 0%, #ff0 17%, #0f0 33%, #0ff 50%, #00f 67%, #f0f 83%, #f00 100%
		);
	}
	.sat {
		background: linear-gradient(to right, #fff, var(--full));
	}
	input[type='range']::-webkit-slider-thumb {
		appearance: none;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: var(--thumb);
		border: 4px solid var(--text);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
	}
	input[type='range']::-moz-range-thumb {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		background: var(--thumb);
		border: 4px solid var(--text);
	}
</style>