import type { APIRoute, GetStaticPaths } from 'astro';
import { getCollection } from 'astro:content';
import { docPageFromEntry, mdPathForId, pageMarkdown } from '../lib/agent-md';

export const getStaticPaths = (async () => {
	const docs = await getCollection('docs');
	return docs.map((entry) => {
		const page = docPageFromEntry({
			id: entry.id,
			data: entry.data,
			body: entry.body,
			filePath: (entry as { filePath?: string }).filePath,
		});
		const path = mdPathForId(page.id).replace(/^\//, '').replace(/\.md$/, '');
		return {
			params: { slug: path },
			props: page,
		};
	});
}) satisfies GetStaticPaths;

export const GET: APIRoute = ({ props }) => {
	const body = pageMarkdown({
		id: props.id,
		title: props.title,
		description: props.description,
		body: props.body,
		filePath: props.filePath,
	});
	if (body.includes('import {')) {
		throw new Error(`flattened ${props.id} still contains import {`);
	}
	return new Response(body, {
		headers: { 'Content-Type': 'text/plain; charset=utf-8' },
	});
};
