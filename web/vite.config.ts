import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// Static adapter: builds plain files the Go server can serve.
			// fallback: every unknown URL loads index.html (needed for a single-page app).
			adapter: adapter({ fallback: 'index.html' })
		})
	],

	// Dev only: forward API and WebSocket calls to the Go server on :8080,
	// so the browser sees one site and there's no CORS problem.
	server: {
		proxy: {
			'/api': 'http://localhost:8080',
			'/health': 'http://localhost:8080',
			'/ws': { target: 'ws://localhost:8080', ws: true }
		}
	}
});