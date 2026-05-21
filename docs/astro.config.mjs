// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://goptics.github.io/vizb',
	base: '/vizb/',
	integrations: [
		starlight({
			title: 'Vizb',
			description: 'Transform Go benchmark output into interactive 4D visualizations',
			logo: { src: './src/assets/logo.png', replacesTitle: false },
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/goptics/vizb' }],
			sidebar: [
				{ label: 'Getting Started', slug: 'getting-started' },
				{ label: 'Installation', slug: 'installation' },
				{ label: 'Features', slug: 'features' },
				{
					label: 'Commands',
					items: [
						{ label: 'vizb', slug: 'commands/root' },
						{ label: 'vizb merge', slug: 'commands/merge' },
						{ label: 'vizb html', slug: 'commands/html' },
					],
				},
				{
					label: 'UI',
					items: [
						{ label: 'Overview', slug: 'ui' },
						{ label: 'Settings', slug: 'ui/settings' },
						{ label: 'Axis Swapping', slug: 'ui/swapping' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Grouping', slug: 'guides/grouping' },
						{ label: 'Merging', slug: 'guides/merging' },
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
			credits: false,
		}),
	],
});
