// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightThemeRapide from 'starlight-theme-rapide'
import starlightOpenAPI, { openAPISidebarGroups } from 'starlight-openapi';

// https://astro.build/config
export default defineConfig({
	site: 'https://vizb.goptics.org',
	// Preserve old top-level URLs after the Getting Started restructure.
	// Fragment anchors (e.g. #docker) stay on the client across the redirect.
	redirects: {
		'/installation': '/getting-started/install',
		'/installation/': '/getting-started/install',
	},
	integrations: [
	starlight({
		title: 'One command to visualize your data',
		description: 'Turn CSV, JSON, and benchmarks into interactive charts and stats without writing chart code. One HTML report you can open in any browser.',
		logo: {
			dark: './src/assets/logo-dark.svg',
			light: './src/assets/logo-light.svg',
			replacesTitle: true,
		},
		social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/goptics/vizb' }],
		sidebar: [
			{
				label: 'Getting Started',
				items: [
					{ label: 'Introduction', slug: 'getting-started' },
					{ label: 'Install', slug: 'getting-started/install' },
					{ label: 'Dimensions', slug: 'getting-started/dimensions' },
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
					{ label: 'Sankey Chart', slug: 'charts/sankey' },
					{ label: 'Chord Chart', slug: 'charts/chord' },
					{ label: '3D Charts (WebGL)', slug: 'charts/3d' },
				],
			},
			{
				label: 'Guides',
				items: [
					{ label: 'Group vs Select', slug: 'guides/group-vs-select' },
					{ label: 'Tabular Data (CSV & JSON)', slug: 'guides/data' },
					{ label: 'Group', slug: 'guides/group' },
					{ label: 'Select', slug: 'guides/select' },
					{ label: 'Merging', slug: 'guides/merging' },
					{ label: 'Parser Guide', slug: 'guides/parsers' },
				],
			},
			{
				label: 'Commands',
				items: [
					{ label: 'vizb', slug: 'commands/root' },
					{ label: 'vizb <chart>', slug: 'commands/charts' },
					{ label: 'vizb merge', slug: 'commands/merge' },
					{ label: 'vizb ui', slug: 'commands/ui' },
					{ label: 'vizb serve', slug: 'commands/serve' },
					{ label: 'vizb update', slug: 'commands/update' },
				],
			},
			...openAPISidebarGroups,
			{
				label: 'UI',
				items: [
					{ label: 'Overview', slug: 'ui' },
					{ label: 'Color Themes', slug: 'ui/themes' },
					{ label: 'Settings', slug: 'ui/settings' },
					{ label: 'Axis Swapping', slug: 'ui/swapping' },
					{ label: 'Statistics', slug: 'ui/stats' },
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
			{
				label: 'Examples',
				items: [
					{ label: 'Overview', slug: 'examples' },
					{ label: 'Tabular data', slug: 'examples/tabular-data' },
					{ label: 'Math & 3D', slug: 'examples/math-and-3d' },
					{ label: 'Comparisons', slug: 'examples/comparisons' },
					{ label: 'GitHub Legends', slug: 'examples/github-legends' },
					{ label: 'Benchmarks', slug: 'examples/benchmarks' },
					{ label: 'Showcase', slug: 'examples/showcase' },
				],
			},
			{ label: 'Features', slug: 'features' },
			{ label: 'How It Works', slug: 'internals/how-it-works' },
			{ label: 'Troubleshooting', slug: 'troubleshooting' },
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
			starlightOpenAPI([
				{
					base: 'api',
					schema: '../api/openapi.yaml',
				},
			]),
			starlightThemeRapide()
		],
		components: {
			// Defer Pagefind until first search open (Lighthouse TBT / network).
			Search: './src/components/Search.astro',
		},
		head: [
			{
				tag: 'meta',
				attrs: {
					property: 'og:image',
					content: 'https://vizb.goptics.org/og-image.png',
				},
			},
			{
				tag: 'meta',
				attrs: { property: 'og:image:width', content: '1200' },
			},
			{
				tag: 'meta',
				attrs: { property: 'og:image:height', content: '600' },
			},
			{
				tag: 'meta',
				attrs: { name: 'twitter:card', content: 'summary_large_image' },
			},
			{
				tag: 'meta',
				attrs: {
					name: 'twitter:image',
					content: 'https://vizb.goptics.org/og-image.png',
				},
			},
		],
	}),
	],
});
