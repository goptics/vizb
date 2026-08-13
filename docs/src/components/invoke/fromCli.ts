/** Map a vizb CLI invocation to HTTP POST / JSON and a GitHub Action step. */

const CHART_TYPES = new Set([
	'bar',
	'line',
	'scatter',
	'pie',
	'heatmap',
	'radar',
	'sankey',
	'chord',
]);

const PIPELINE_FLAGS: Record<string, 'group' | 'pattern' | 'regex' | 'select' | 'charts' | 'output'> =
	{
		'-g': 'group',
		'--group': 'group',
		'-p': 'pattern',
		'--group-pattern': 'pattern',
		'-r': 'regex',
		'--group-regex': 'regex',
		'--select': 'select',
		'-c': 'charts',
		'--charts': 'charts',
		'-o': 'output',
		'--output': 'output',
	};

const CONFIG_VALUE_FLAGS: Record<string, 'scale' | 'symbolSize'> = {
	'-S': 'scale',
	'--scale': 'scale',
	'--symbol-size': 'symbolSize',
};

const BOOL_FLAGS: Record<string, string> = {
	'-l': 'showLabels',
	'--show-labels': 'showLabels',
	'--stack': 'stack',
	'--3d': 'threeD',
	'--3d-visualmap': 'threeDVisualMap',
	'--3d-rotate': 'threeDRotate',
	'--visualmap': 'visualMap',
};

const ACTION_CONFIG_KEY: Record<string, string> = {
	showLabels: 'labels',
	scale: 'scale',
	stack: 'stack',
	threeD: '3d',
	threeDVisualMap: '3d-visualmap',
	threeDRotate: '3d-rotate',
	visualMap: 'visualmap',
	symbolSize: 'symbol-size',
};

export interface InvokeSnippets {
	cli: string;
	http: string;
	action: string;
}

interface ParsedCli {
	file?: string;
	chart?: string;
	group?: string;
	pattern?: string;
	regex?: string;
	select: string[];
	charts?: string;
	output?: string;
	config: Record<string, string | boolean | number>;
}

export function snippetsFromCli(cli: string, input?: string): InvokeSnippets {
	const parsed = parseCli(cli);
	return {
		cli: cli.trim(),
		http: JSON.stringify(buildHttp(parsed, input), null, 2),
		action: buildAction(parsed),
	};
}

function parseCli(cli: string): ParsedCli {
	const tokens = tokenize(cli);
	const parsed: ParsedCli = { select: [], config: {} };

	let i = 0;
	while (i < tokens.length && !isVizbBin(tokens[i] ?? '')) i += 1;
	if (i < tokens.length) i += 1;

	if (i < tokens.length && CHART_TYPES.has(tokens[i] ?? '')) {
		parsed.chart = tokens[i];
		i += 1;
	}

	while (i < tokens.length) {
		const tok = tokens[i] ?? '';
		const valueKey = PIPELINE_FLAGS[tok];
		const configKey = CONFIG_VALUE_FLAGS[tok];
		if (valueKey || configKey) {
			const value = tokens[i + 1];
			if (value !== undefined && !value.startsWith('-')) {
				if (valueKey === 'select') parsed.select.push(value);
				else if (valueKey) parsed[valueKey] = value;
				else if (configKey === 'symbolSize') parsed.config.symbolSize = Number(value);
				else if (configKey) parsed.config[configKey] = value;
				i += 2;
				continue;
			}
		}

		if (BOOL_FLAGS[tok]) {
			parsed.config[BOOL_FLAGS[tok]] = true;
			i += 1;
			continue;
		}

		if (!tok.startsWith('-') && !parsed.file) parsed.file = tok;
		i += 1;
	}

	return parsed;
}

function buildHttp(parsed: ParsedCli, input?: string): Record<string, unknown> {
	const types = chartTypes(parsed);
	const body: Record<string, unknown> = {
		input: input !== undefined ? input.trimEnd() + '\n' : `<contents of ${parsed.file ?? 'input'}>`,
	};

	const parser = httpParser(parsed);
	if (parser) body.parser = parser;

	const grouping: Record<string, unknown> = {};
	if (parsed.group) grouping.columns = parsed.group.split(',').map((c) => c.trim()).filter(Boolean);
	if (parsed.pattern) grouping.pattern = parsed.pattern;
	if (parsed.regex) grouping.regex = parsed.regex;
	if (Object.keys(grouping).length > 0) body.grouping = grouping;

	if (parsed.select.length > 0) body.select = parsed.select;

	if (types.length > 0) {
		const charts: Record<string, unknown> = { types };
		const type = parsed.chart ?? types[0];
		if (type && Object.keys(parsed.config).length > 0) {
			charts.configs = [{ type, ...parsed.config }];
		}
		body.charts = charts;
	}

	body.output = { format: parsed.output?.toLowerCase().endsWith('.json') ? 'dataset' : 'html' };
	return body;
}

function buildAction(parsed: ParsedCli): string {
	const withBlock: [string, string][] = [];
	if (parsed.file) withBlock.push(['file', parsed.file]);
	if (parsed.group) withBlock.push(['group', parsed.group]);
	if (parsed.pattern) withBlock.push(['group-pattern', parsed.pattern]);
	if (parsed.regex) withBlock.push(['group-regex', parsed.regex]);
	if (parsed.select.length === 1) withBlock.push(['select', parsed.select[0] ?? '']);
	else if (parsed.select.length > 1) withBlock.push(['select', parsed.select.join('\n')]);

	const types = chartTypes(parsed);
	if (types.length > 0) withBlock.push(['charts', types.join(',')]);

	const props: string[] = [];
	for (const [key, value] of Object.entries(parsed.config)) {
		const actionKey = ACTION_CONFIG_KEY[key];
		if (!actionKey) continue;
		props.push(value === true ? actionKey : `${actionKey}=${value}`);
	}
	const chartType = parsed.chart ?? types[0];
	if (chartType && props.length > 0) withBlock.push(['chart', `${chartType}:${props.join(';')}`]);

	if (parsed.output) {
		withBlock.push([
			parsed.output.toLowerCase().endsWith('.json') ? 'output-json' : 'output-html',
			parsed.output,
		]);
	}

	const lines = ['- uses: goptics/vizb@v0', '  with:'];
	for (const [key, value] of withBlock) {
		if (value.includes('\n')) {
			lines.push(`    ${key}: |`);
			for (const row of value.split('\n')) lines.push(`      ${row}`);
		} else {
			lines.push(`    ${key}: ${yamlScalar(value)}`);
		}
	}
	return lines.join('\n');
}

function httpParser(parsed: ParsedCli): string | undefined {
	const file = parsed.file ?? '';
	if (file.endsWith('.csv')) return 'csv';
	if (file.endsWith('.json')) return 'json';
	if (file.endsWith('.txt') && parsed.pattern?.includes('/')) return 'go';
	return undefined;
}

function chartTypes(parsed: ParsedCli): string[] {
	if (parsed.charts) {
		return parsed.charts
			.split(',')
			.map((t) => t.trim())
			.filter((t) => CHART_TYPES.has(t));
	}
	if (parsed.chart) return [parsed.chart];
	return [];
}

function isVizbBin(token: string): boolean {
	return token === 'vizb' || token.endsWith('/vizb');
}

function yamlScalar(value: string): string {
	if (/^[A-Za-z0-9_.,/@+-]+$/.test(value)) return value;
	return JSON.stringify(value);
}

function tokenize(src: string): string[] {
	return (
		src
			.replace(/\\\r?\n/g, ' ')
			.match(/"[^"]*"|'[^']*'|\S+/g)
			?.map((tok) => tok.replace(/^['"]|['"]$/g, '')) ?? []
	);
}
