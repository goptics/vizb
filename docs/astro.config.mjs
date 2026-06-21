// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeRapide from 'starlight-theme-rapide'

// https://astro.build/config
export default defineConfig({
	site: 'https://vizb.goptics.org',
	redirects: {
		'/ui/heatmap': '/charts/heatmap',
		'/ui/3d-charts': '/charts/3d',
	},
	integrations: [
	starlight({
		title: 'Turn any table into interactive charts',
		description: 'Vizb renders any CSV or JSON table — and benchmark output from Go, Rust, and JavaScript — as interactive 4D charts in a single self-contained HTML file.',
		logo: {
			dark: './src/assets/logo-dark.svg',
			light: './src/assets/logo-light.svg',
			replacesTitle: true,
		},
		social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/goptics/vizb' }],
		sidebar: [
			{ label: 'Getting Started', slug: 'getting-started' },
			{ label: 'Installation', slug: 'installation' },
			{ label: 'Features', slug: 'features' },
			{
				label: 'Commands',
				items: [
					{ label: 'vizb', slug: 'commands/root' },
					{ label: 'vizb <chart>', slug: 'commands/charts' },
					{ label: 'vizb merge', slug: 'commands/merge' },
					{ label: 'vizb ui', slug: 'commands/ui' },
				],
			},
			{
				label: 'UI',
				items: [
					{ label: 'Overview', slug: 'ui' },
					{ label: 'Settings', slug: 'ui/settings' },
					{ label: 'Axis Swapping', slug: 'ui/swapping' },
					{ label: 'Statistics', slug: 'ui/stats' },
				],
			},
			{
				label: 'Charts',
				items: [
					{ label: 'Overview', slug: 'charts' },
					{ label: 'Bar Chart', slug: 'charts/bar' },
					{ label: 'Line Chart', slug: 'charts/line' },
					{ label: 'Scatter Chart', slug: 'charts/scatter' },
					{ label: 'Pie Chart', slug: 'charts/pie' },
					{ label: 'Radar Chart', slug: 'charts/radar' },
					{ label: 'Heatmap', slug: 'charts/heatmap' },
					{ label: '3D Charts (WebGL)', slug: 'charts/3d' },
				],
			},
			{
				label: 'Guides',
				items: [
					{ label: 'Grouping', slug: 'guides/grouping' },
					{ label: 'Merging', slug: 'guides/merging' },
					{ label: 'Parser Guide', slug: 'guides/parsers' },
					{ label: 'Tabular Data (CSV & JSON)', slug: 'guides/data' },
				],
			},
			{
				label: 'CI/CD',
				items: [
					{ label: 'GitHub Action', slug: 'ci-cd/github-action' },
					{ label: 'Stateless CI', slug: 'ci-cd/stateless' },
					{ label: 'Stateful CI', slug: 'ci-cd/stateful' },
					{ label: 'Deploying', slug: 'ci-cd/deploying' },
				],
			},
			{ label: 'How It Works', slug: 'internals/how-it-works' },
			{ label: 'Examples', slug: 'examples' },
			{
				label: 'Roadmap',
				slug: 'roadmap',
				badge: { text: 'Future', variant: 'caution' },
			},
		],
		editLink: {
			baseUrl: 'https://github.com/goptics/vizb/edit/main/docs/',
		},
		lastUpdated: true,
		tableOfContents: {
			minHeadingLevel: 2,
			maxHeadingLevel: 4,
		},
		plugins: [
			starlightThemeRapide()
		],
		head: [
			{
				tag: 'meta',
				attrs: { property: 'og:image', content: '/og-image.png' },
			},
			{
				tag: 'meta',
				attrs: { property: 'og:image:width', content: '1200' },
			},
			{
				tag: 'meta',
				attrs: { property: 'og:image:height', content: '600' },
			}
		],
	}),
	],
});
