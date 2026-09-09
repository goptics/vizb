import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { describe, it } from 'node:test';
import { fileURLToPath } from 'node:url';
import {
	CORE_MD_PATHS,
	buildLlmsTxt,
	flattenMdx,
	mdPathForId,
	pageMarkdown,
	sourceBody,
} from './agent-md.ts';

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = join(here, '../../..');

describe('mdPathForId', () => {
	it('maps index pages to /dir.md not /dir/index.md', () => {
		assert.equal(mdPathForId('getting-started/index'), '/getting-started.md');
		assert.equal(mdPathForId('getting-started'), '/getting-started.md');
		assert.equal(mdPathForId('charts/bar'), '/charts/bar.md');
		assert.equal(mdPathForId('index'), '/index.md');
	});
});

describe('flattenMdx', () => {
	it('drops import/export and does not leave import {', () => {
		const src = `import { Aside } from '@astrojs/starlight/components';\n\nHello\n`;
		const out = flattenMdx(src);
		assert.equal(out.includes('import {'), false);
		assert.match(out, /Hello/);
	});

	it('keeps import/export lines inside fenced code', () => {
		const src = '```bash\nexport PATH="$PATH:/usr/local/bin"\nimport { x } from "pkg"\n```\n';
		const out = flattenMdx(src);
		assert.match(out, /export PATH=/);
		assert.match(out, /import \{ x \} from "pkg"/);
	});

	it('turns Aside into a blockquote', () => {
		const src = `<Aside type="note">\nBe kind.\n</Aside>\n`;
		assert.match(flattenMdx(src), /> Be kind/);
	});

	it('turns TabItem into a heading and keeps fenced commands', () => {
		const src = `<TabItem label="Linux / macOS">\n\n\`\`\`bash\ncurl -fsSL https://vizb.goptics.org/install.sh | bash\n\`\`\`\n\n</TabItem>\n`;
		const out = flattenMdx(src);
		assert.match(out, /### Linux \/ macOS/);
		assert.match(out, /install\.sh/);
		assert.equal(out.includes('<TabItem'), false);
	});

	it('inlines relative .mdx imports so QuickInstall survives', () => {
		const installPath = join(
			repoRoot,
			'docs/src/content/docs/getting-started/install.mdx',
		);
		const src = readFileSync(installPath, 'utf8').replace(/^---[\s\S]*?---\n/, '');
		const out = flattenMdx(src, { fromFile: installPath });
		assert.match(out, /https:\/\/vizb\.goptics\.org\/install\.sh/);
		assert.equal(out.includes('<QuickInstall'), false);
		assert.equal(out.includes('import {'), false);
	});

	it('lifts InvokeTabs cli with > in the command into a bash fence', () => {
		const src = `<InvokeTabs
  cli={\`vizb bar data.csv --group-regex "(?<n>.*)/(?<x>.*)/(?<y>.*)/(?<z>.*)" -o out.html\`}
/>
`;
		const out = flattenMdx(src);
		assert.match(out, /```bash/);
		assert.match(out, /vizb bar data.csv --group-regex/);
		assert.match(out, /\(\?<n>\.\*\)/);
		assert.equal(out.includes('<InvokeTabs'), false);
	});

	it('unwraps Steps and keeps inner fenced bash', () => {
		const src = `<Steps>\n\n\`\`\`bash\nvizb merge v1.json v2.json --tag v1\n\`\`\`\n\n</Steps>\n`;
		const out = flattenMdx(src);
		assert.match(out, /```bash/);
		assert.match(out, /vizb merge v1\.json v2\.json --tag v1/);
		assert.equal(out.includes('<Steps'), false);
	});

	it('lifts a simple InvokeTabs cli into a bash fence', () => {
		const src = `<InvokeTabs cli={\`vizb bar data.csv -o out.html\`} />\n`;
		const out = flattenMdx(src);
		assert.match(out, /```bash/);
		assert.match(out, /vizb bar data.csv -o out.html/);
		assert.equal(out.includes('<InvokeTabs'), false);
	});

	it('lifts CopyableCsv csv={IDENT} from samples.ts into a csv fence', () => {
		const src = `import { DIMENSIONS_SALES_SAMPLE } from '../../../data/samples';

<CopyableCsv filename="sales.csv" csv={DIMENSIONS_SALES_SAMPLE} />
`;
		const out = flattenMdx(src);
		assert.match(out, /```csv/);
		assert.match(out, /region,product,sales/);
		assert.equal(out.includes('<CopyableCsv'), false);
	});

	it('lifts SalesSampleCsv into a csv fence of SALES_SAMPLE', () => {
		const src = `import SalesSampleCsv from '../../../components/SalesSampleCsv.astro';

<SalesSampleCsv />
`;
		const out = flattenMdx(src);
		assert.match(out, /```csv/);
		assert.match(out, /order_date,region,category/);
		assert.equal(out.includes('<SalesSampleCsv'), false);
	});

	it('lifts Card title into a heading and keeps inner text', () => {
		const src = `<Card title="Bar Chart" icon="bars" href="/charts/bar">
  Compare values across categories.
</Card>
`;
		const out = flattenMdx(src);
		assert.match(out, /### Bar Chart/);
		assert.match(out, /Compare values across categories/);
		assert.equal(out.includes('<Card'), false);
	});
});

describe('flattenMdx corpus', () => {
	function contentPathForMd(mdPath: string): string {
		const slug = mdPath.replace(/^\//, '').replace(/\.md$/, '');
		const docsRoot = join(repoRoot, 'docs/src/content/docs');
		const candidates = [
			join(docsRoot, `${slug}.mdx`),
			join(docsRoot, slug, 'index.mdx'),
		];
		const found = candidates.find((p) => existsSync(p));
		assert.ok(found, `missing content for ${mdPath}`);
		return found;
	}

	it('flattens every Core page without leftover JSX and keeps key commands', () => {
		const byPath = new Map<string, string>();
		for (const mdPath of CORE_MD_PATHS) {
			const filePath = contentPathForMd(mdPath);
			const out = flattenMdx(readFileSync(filePath, 'utf8'), { fromFile: filePath });
			byPath.set(mdPath, out);
			assert.equal(out.includes('import {'), false, mdPath);
			assert.equal(/<[A-Z]/.test(out), false, `${mdPath} leftover JSX`);
		}

		const merging = byPath.get('/guides/merging.md') ?? '';
		assert.match(merging, /vizb merge/);
		assert.match(merging, /--tag/);

		const dimensions = byPath.get('/getting-started/dimensions.md') ?? '';
		assert.match(dimensions, /vizb/);
		assert.match(dimensions, /```csv/);
		assert.match(dimensions, /region,product/);

		const charts = byPath.get('/charts.md') ?? '';
		assert.match(charts, /Bar Chart/);
		assert.match(charts, /Line Chart/);

		const charts3d = byPath.get('/charts/3d.md') ?? '';
		assert.equal(charts3d.includes('<InvokeTabs'), false);
	});
});

describe('CORE_MD_PATHS', () => {
	it('lists every Core page from the spec', () => {
		const expected = [
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
		];
		assert.deepEqual(CORE_MD_PATHS, expected);
	});
});

describe('buildLlmsTxt', () => {
	it('links only .md URLs and llms-full.txt, never HTML or GitHub raw', () => {
		const txt = buildLlmsTxt('https://vizb.goptics.org/', [
			{
				id: 'getting-started/install',
				title: 'Install',
				description: 'Install vizb',
				body: '# Install\n',
			},
		]);
		assert.match(txt, /https:\/\/vizb\.goptics\.org\/getting-started\/install\.md/);
		assert.match(txt, /llms-full\.txt/);
		const hrefs = [...txt.matchAll(/\((https?:\/\/[^)\s]+)\)/g)].map((m) => m[1]);
		assert.ok(hrefs.length > 0);
		for (const href of hrefs) {
			assert.equal(new URL(href).hostname, 'vizb.goptics.org');
		}
		assert.equal(/https:\/\/vizb\.goptics\.org\/getting-started\/install[^.m]/.test(txt), false);
	});
});

describe('pageMarkdown', () => {
	it('keeps title frontmatter and flattened body', () => {
		const out = pageMarkdown({
			id: 'guides/group',
			title: 'Group',
			description: 'Group columns',
			body: 'import { Aside } from "x";\n\n# Group\n',
		});
		assert.match(out, /^---\ntitle: "Group"\n/);
		assert.equal(out.includes('import {'), false);
	});
});

describe('sourceBody', () => {
	it('throws when a Core page has no body and no file', () => {
		assert.throws(
			() =>
				sourceBody({
					id: 'getting-started/install/index',
					body: '',
					filePath: join(here, 'missing-core.mdx'),
				}),
			/empty markdown body for core docs page getting-started\/install\/index/,
		);
	});

	it('resolves id/index.mdx when body is empty', () => {
		const out = sourceBody({ id: 'getting-started', body: '' });
		assert.match(out.filePath ?? '', /getting-started\/index\.mdx$/);
		assert.match(out.body, /Need the binary first/);
	});
});

