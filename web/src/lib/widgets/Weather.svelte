<script lang="ts">
	import { onMount } from 'svelte';

	let { settings }: { settings: { name?: string; lat?: number; lon?: number } } = $props();

	type Weather = {
		temp: number;
		feels: number;
		code: number;
		isDay: boolean;
		max: number;
		min: number;
		rain: number; // today's max chance of rain, %
	};

	let weather = $state<Weather | null>(null);
	let error = $state('');

	let place = $derived(settings.name ?? 'Paços de Ferreira ');

	async function load() {
		const lat = settings.lat ?? 41.17;
		const lon = settings.lon ?? -8.23;
		const url =
			'https://api.open-meteo.com/v1/forecast' +
			`?latitude=${lat}&longitude=${lon}` +
			'&current=temperature_2m,apparent_temperature,weather_code,is_day' +
			'&daily=temperature_2m_max,temperature_2m_min,precipitation_probability_max' +
			'&timezone=Europe%2FLisbon&forecast_days=1';
		try {
			const res = await fetch(url);
			if (!res.ok) throw new Error(`HTTP ${res.status}`);
			const d = await res.json();
			weather = {
				temp: Math.round(d.current.temperature_2m),
				feels: Math.round(d.current.apparent_temperature),
				code: d.current.weather_code,
				isDay: d.current.is_day === 1,
				max: Math.round(d.daily.temperature_2m_max[0]),
				min: Math.round(d.daily.temperature_2m_min[0]),
				rain: d.daily.precipitation_probability_max[0] ?? 0
			};
			error = '';
		} catch {
			error = 'Weather unavailable';
		}
	}

	// Load now, then every 15 minutes; stop when the widget is removed.
	onMount(() => {
		load();
		const timer = setInterval(load, 15 * 60 * 1000);
		return () => clearInterval(timer);
	});

	// WMO weather codes (what Open-Meteo returns) → icon + label
	function describe(code: number, isDay: boolean): { icon: string; label: string } {
		if (code === 0) return { icon: isDay ? '☀️' : '🌙', label: 'Clear' };
		if (code <= 2) return { icon: isDay ? '🌤️' : '☁️', label: 'Partly cloudy' };
		if (code === 3) return { icon: '☁️', label: 'Cloudy' };
		if (code <= 48) return { icon: '🌫️', label: 'Fog' };
		if (code <= 57) return { icon: '🌦️', label: 'Drizzle' };
		if (code <= 67) return { icon: '🌧️', label: 'Rain' };
		if (code <= 77) return { icon: '🌨️', label: 'Snow' };
		if (code <= 82) return { icon: '🌧️', label: 'Showers' };
		if (code <= 86) return { icon: '🌨️', label: 'Snow showers' };
		return { icon: '⛈️', label: 'Thunderstorm' };
	}

	let look = $derived(weather ? describe(weather.code, weather.isDay) : null);
</script>

{#if error}
	<p class="muted">{error}</p>
{:else if !weather || !look}
	<p class="muted">Loading weather…</p>
{:else}
	<div class="weather">
		<span class="icon">{look.icon}</span>
		<div class="main">
			<span class="temp">{weather.temp}°</span>
			<span class="muted">{look.label} · feels {weather.feels}°</span>
		</div>
		<div class="side">
			<b>{place}</b>
			<span class="muted">↑ {weather.max}° ↓ {weather.min}°</span>
			{#if weather.rain >= 40}
				<span class="umbrella">☂ Take an umbrella ({weather.rain}%)</span>
			{:else}
				<span class="muted">Rain {weather.rain}%</span>
			{/if}
		</div>
	</div>
{/if}

<style>
	.weather {
		height: 100%;
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}
	.icon {
		font-size: 3.5rem;
		line-height: 1;
	}
	.main {
		display: grid;
	}
	.temp {
		font-size: 3rem;
		font-weight: 700;
		line-height: 1;
	}
	.side {
		margin-left: auto;
		display: grid;
		gap: 0.2rem;
		text-align: right;
	}
	.umbrella {
		color: var(--accent);
		font-weight: 600;
	}
	.muted {
		color: var(--muted);
	}
</style>