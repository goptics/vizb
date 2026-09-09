import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';
import { buildLlmsTxt, docPageFromEntry, type DocPage } from '../lib/agent-md';

export const GET: APIRoute = async ({ site }) => {
	const docs = await getCollection('docs');
	const pages: DocPage[] = docs.map((entry) =>
		docPageFromEntry({
			id: entry.id,
			data: entry.data,
			body: entry.body,
			filePath: (entry as { filePath?: string }).filePath,
		}),
	);
	const origin = site?.href ?? 'https://vizb.goptics.org/';
	return new Response(buildLlmsTxt(origin, pages), {
		headers: { 'Content-Type': 'text/plain; charset=utf-8' },
	});
};
