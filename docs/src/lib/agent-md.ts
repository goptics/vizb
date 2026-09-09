import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';

export const CORE_MD_PATHS = [
	'/getting-started.md',
	'/getting-started/install.md',
	'/getting-started/dimensions.md',
	'/guides/group-vs-select.md',
	'/guides/group.md',
	'/guides/select.md',
	'/guides/data.md',
	'/guides/parsers.md',
	'/guides/merging.md',
	'/charts.md',
	'/charts/bar.md',
	'/charts/line.md',
	'/charts/scatter.md',
	'/charts/pie.md',
	'/charts/heatmap.md',
	'/charts/radar.md',
	'/charts/sankey.md',
	'/charts/chord.md',
	'/charts/3d.md',
	'/commands/root.md',
	'/commands/charts.md',
	'/commands/merge.md',
	'/troubleshooting.md',
] as const;

export function mdPathForId(id: string): string {
	const trimmed = id.replace(/\/index$/, '').replace(/^index$/, 'index');
	const slug = trimmed === 'index' ? 'index' : trimmed.replace(/\/index$/, '');
	return `/${slug}.md`;
}

/** Quoted/backtick spans may contain `>` (named-capture regex in cli). */
const ATTR_CHUNK = String.raw`(?:[^>\`'"]|\`[^\`]*\`|'[^']*'|"[^"]*")*`;

function quotedProp(tag: string, prop: string): string | undefined {
	const fromTicks = tag.match(new RegExp(`${prop}=\\{\`([\\s\\S]*?)\`\\}`));
	if (fromTicks) return fromTicks[1];
	const fromDq = tag.match(new RegExp(`${prop}="([^"]*)"`));
	if (fromDq) return fromDq[1];
	const fromSq = tag.match(new RegExp(`${prop}='([^']*)'`));
	if (fromSq) return fromSq[1];
}

function liftFence(
	src: string,
	name: string,
	prop: string,
	lang: string,
): string {
	return src.replace(new RegExp(`<${name}\\b${ATTR_CHUNK}\\/>`, 'g'), (tag) => {
		const value = quotedProp(tag, prop);
		return value ? `\n\`\`\`${lang}\n${value.trim()}\n\`\`\`\n` : '';
	});
}

export function flattenMdx(
	source: string,
	opts: { fromFile?: string } = {},
): string {
	let body = source.replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, '');
	const locals = new Map<string, string>();

	body = body.replace(
		/^import\s+(\w+)\s+from\s+['"](.+\.mdx)['"]\s*;?[ \t]*$/gm,
		(_m, name: string, spec: string) => {
			if (!opts.fromFile) return '';
			const abs = join(dirname(opts.fromFile), spec);
			const imported = readFileSync(abs, 'utf8');
			locals.set(name, flattenMdx(imported, { fromFile: abs }));
			return '';
		},
	);

	body = body.replace(/^import\s+[\s\S]*?from\s+['"][^'"]+['"]\s*;?[ \t]*$/gm, '');
	body = body.replace(/^export\s+[\s\S]*?;?[ \t]*$/gm, '');

	for (const [name, md] of locals) {
		body = body.replace(new RegExp(`<${name}\\s*/>`, 'g'), md);
		body = body.replace(
			new RegExp(`<${name}\\b[^>]*>[\\s\\S]*?</${name}>`, 'g'),
			md,
		);
	}

	body = liftFence(body, 'InvokeTabs', 'cli', 'bash');
	body = liftFence(body, 'CopyableCsv', 'csv', 'csv');

	body = body.replace(
		/<Aside\b[^>]*>\s*([\s\S]*?)\s*<\/Aside>/g,
		(_m, inner: string) =>
			inner
				.trim()
				.split(/\n/)
				.map((line: string) => `> ${line}`)
				.join('\n'),
	);

	body = body.replace(
		/<TabItem\b[^>]*\blabel="([^"]+)"[^>]*>\s*([\s\S]*?)\s*<\/TabItem>/g,
		(_m, label: string, inner: string) => `### ${label}\n\n${inner.trim()}\n`,
	);

	body = body.replace(
		/<\/?(?:Tabs|TabItem|CardGrid|Card|LinkCard|Steps|FileTree)\b[^>]*>/g,
		'',
	);
	body = body.replace(new RegExp(`<[A-Z][\\w.]*\\b${ATTR_CHUNK}\\/>`, 'g'), '');
	body = body.replace(
		new RegExp(`<([A-Z][\\w.]*)\\b${ATTR_CHUNK}>([\\s\\S]*?)<\\/\\1>`, 'g'),
		(_m, _name: string, inner: string) => inner,
	);

	return body.replace(/\n{3,}/g, '\n\n').trim() + '\n';
}

export type DocPage = {
	id: string;
	title: string;
	description: string;
	body: string;
	filePath?: string;
};

export function pageMarkdown(page: DocPage): string {
	const md = flattenMdx(page.body, { fromFile: page.filePath });
	return `---\ntitle: ${page.title}\ndescription: ${page.description}\n---\n\n${md}`;
}

export function buildLlmsTxt(
	site: string,
	pages: DocPage[],
): string {
	const origin = site.replace(/\/$/, '');
	const byPath = new Map(
		pages.map((p) => [mdPathForId(p.id), p] as const),
	);

	const line = (path: string) => {
		const page = byPath.get(path);
		const title = page?.title ?? path;
		const desc = page?.description ? `: ${page.description}` : '';
		return `- [${title}](${origin}${path})${desc}`;
	};

	const coreSet = new Set<string>(CORE_MD_PATHS);
	const more = pages
		.map((p) => mdPathForId(p.id))
		.filter((p) => !coreSet.has(p) && p !== '/index.md')
		.sort();

	return [
		'# Vizb',
		'',
		'> Turn CSV, JSON, and Go/Rust/JavaScript benchmark output into a self-contained interactive HTML chart without writing chart code.',
		'',
		'Vizb is a CLI pipeline (auto-group, n/x/y/z dimensions, HTML or JSON), not a charting library to embed. Prefer minimal flags. Fetch one Core page rather than llms-full.txt.',
		'',
		'## Core',
		'',
		...CORE_MD_PATHS.map(line),
		'',
		'## More',
		'',
		...more.map(line),
		'',
		'## Optional',
		'',
		`- [Complete documentation](${origin}/llms-full.txt): concatenated markdown dump of every docs page`,
		'',
	].join('\n');
}

export function buildLlmsFull(pages: DocPage[]): string {
	return pages
		.map((p) => pageMarkdown(p))
		.join('\n\n----------\n\n');
}
