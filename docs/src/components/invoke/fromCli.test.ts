import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { snippetsFromCli } from './fromCli.ts';

describe('snippetsFromCli agent', () => {
	it('turns grouped labels into a /vizb prompt and always names the chart', () => {
		const { agent } = snippetsFromCli(
			'vizb bar sales.csv -g region,product -p x,y -l -o sales.html',
		);
		assert.equal(agent, '/vizb sales.csv as bar by region & product, show labels');
	});

	it('names bar, pie, and stacked extras', () => {
		const { agent } = snippetsFromCli(
			'vizb pie data.csv -g impl -o pie.html',
		);
		assert.equal(agent, '/vizb data.csv as pie by impl');

		assert.equal(
			snippetsFromCli('vizb bar sales.csv -g region,category -p x,y --stack -o stacked.html')
				.agent,
			'/vizb sales.csv as bar by region & category, stacked',
		);
	});

	it('defaults to as bar when the CLI has no chart type', () => {
		assert.equal(
			snippetsFromCli('vizb sales.csv -o sales.html').agent,
			'/vizb sales.csv as bar',
		);
	});

	it('joins --charts types and keeps trivial -p out of the prompt', () => {
		assert.equal(
			snippetsFromCli('vizb data.csv -g category,metric -p x,y -c bar,radar -o output.html')
				.agent,
			'/vizb data.csv as bar & radar by category & metric',
		);
	});

	it('verbalizes bracket patterns, labels, and regex instead of {} [] syntax', () => {
		assert.equal(
			snippetsFromCli(
				'vizb bar sales.csv -g order_date,category -p "[-y{Month}-x{Date}],z{Category}" -o out.html',
			).agent,
			'/vizb sales.csv as bar, split order_date into month & date (skip the year), category as depth',
		);
		assert.equal(
			snippetsFromCli(
				'vizb bar data.csv -g category,metric,group,series --group-regex "(?<n>.*)/(?<x>.*)/(?<y>.*)/(?<z>.*)" -o output.html',
			).agent,
			'/vizb data.csv as bar by category, metric, group & series, split by / into name, x, y & depth',
		);
	});

	it('does not leave [] {} or regex captures in derived agent prompts', () => {
		const clis = [
			'vizb bar sales.csv -g order_date,category -p "[-y{Month}-x{Date}],z{Category}" -o out.html',
			'vizb pie sales.csv -g order_date,category -p "[-y{Month}-x{Date}],z{Category}" -o out.html',
			'vizb bar data.csv -g category,metric,group,series --group-regex "(?<n>.*)/(?<x>.*)/(?<y>.*)/(?<z>.*)" -o output.html',
			'vizb line worker-pools.txt -p z/y/x --scale log -o out.html',
		];
		for (const cli of clis) {
			const { agent } = snippetsFromCli(cli);
			assert.equal(/[\[\]{}]/.test(agent), false, agent);
			assert.equal(agent.includes('?<'), false, agent);
			assert.match(agent, / as (?:bar|line|scatter|pie|heatmap|radar|sankey|chord)\b/);
		}
	});

	it('joins select columns and repeats with a semicolon', () => {
		assert.equal(
			snippetsFromCli(
				'vizb bar sales.csv -g region,product -p x,y --select amount,quantity -o bars.html',
			).agent,
			'/vizb sales.csv as bar by region & product, amount & quantity',
		);
		assert.equal(
			snippetsFromCli(
				`vizb sankey sankey-flows.csv \\
  --select source,target,value \\
  --select source,target,cost \\
  -o out.html`,
			).agent,
			'/vizb sankey-flows.csv as sankey, source, target & value; source, target & cost',
		);
	});

	it('maps 3d, scale, and symbol-size config flags', () => {
		assert.equal(
			snippetsFromCli('./bin/vizb bar noise-surface.csv -o surface.html --3d-visualmap')
				.agent,
			'/vizb noise-surface.csv as bar, 3d visual map',
		);
		assert.equal(
			snippetsFromCli('vizb line worker-pools.txt -p z/y/x --scale log -o out.html').agent,
			'/vizb worker-pools.txt as line, split into depth, y & x, log scale',
		);
		assert.equal(
			snippetsFromCli(
				'vizb scatter clusters.csv --visualmap --symbol-size 10 -o scatter.html',
			).agent,
			'/vizb clusters.csv as scatter, visual map, symbol size 10',
		);
	});
});
