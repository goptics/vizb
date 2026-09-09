import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { describe, it } from 'node:test';
import { fileURLToPath } from 'node:url';
import { CORE_MD_PATHS, flattenMdx, mdPathForId } from './agent-md.ts';

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
