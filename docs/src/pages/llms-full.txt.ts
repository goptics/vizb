import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';
import { buildLlmsFull, docPageFromEntry, type DocPage } from '../lib/agent-md';

export const GET: APIRoute = async () => {
	const docs = await getCollection('docs');
	const pages: DocPage[] = docs.map((entry) =>
		docPageFromEntry({
			id: entry.id,
			data: entry.data,
			body: entry.body,
			filePath: (entry as { filePath?: string }).filePath,
		}),
	);
	return new Response(buildLlmsFull(pages), {
		headers: { 'Content-Type': 'text/plain; charset=utf-8' },
	});
};
