import type { APIRoute } from 'astro';

export const GET: APIRoute = ({ site }) => {
	const sitemapURL = new URL('sitemap-index.xml', site);
	const body = `
User-agent: *
Allow: /

Sitemap: ${sitemapURL.href}
`.trim();

	return new Response(body, {
		headers: { 'Content-Type': 'text/plain; charset=utf-8' },
	});
};