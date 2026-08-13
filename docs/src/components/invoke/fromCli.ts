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

const VALUE_FLAGS: Record<string, string> = {
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
	'-s': 'sort',
	'--sort': 'sort',
	'-S': 'scale',
	'--scale': 'scale',
	'--swap': 'swap',
	'--symbol': 'symbol',
	'--symbol-size': 'symbolSize',
	'-P': 'parser',
	'--parser': 'parser',
	'-n': 'name',
	'--name': 'name',
	'-d': 'description',
	'--description': 'description',
	'-t': 'tag',
	'--tag': 'tag',
	'--json-path': 'jsonPath',
	'--col-axis': 'colAxis',
	'--chart': 'chart',
	'--stat': 'stat',
};

/** Value flags that belong on charts.configs / Action `chart:`, not the request root. */
const CONFIG_VALUE_KEYS = new Set(['sort', 'scale', 'swap', 'symbol', 'symbolSize', 'stat']);

const BOOL_FLAGS: Record<string, string> = {
	'-l': 'showLabels',
	'--show-labels': 'showLabels',
	'--stack': 'stack',
	'--horizontal': 'horizontal',
	'--3d': 'threeD',
	'--3d-visualmap': 'threeDVisualMap',
	'--3d-rotate': 'threeDRotate',
	'--visualmap': 'visualMap',
	'--smooth': 'smooth',
	'--round': 'round',
};

const HTTP_PARSER: Record<string, string> = {
	csv: 'csv',
	json: 'json',
	go: 'go',
	javascript: 'javascript',
	'js:tinybench': 'javascript',
	'js:vitest': 'javascript',
	js: 'javascript',
	rust: 'rust',
	'rs:criterion': 'rust',
	'rs:divan': 'rust',
	rs: 'rust',
	auto: 'auto',
};

const ACTION_CONFIG_KEY: Record<string, string> = {
	showLabels: 'labels',
	sort: 'sort',
	scale: 'scale',
	stack: 'stack',
	horizontal: 'horizontal',
	threeD: '3d',
	threeDVisualMap: '3d-visualmap',
	threeDRotate: '3d-rotate',
	swap: 'swap',
	visualMap: 'visualmap',
	smooth: 'smooth',
	symbol: 'symbol',
	symbolSize: 'symbol-size',
	stat: 'stat',
};

export interface InvokeSnippets {
	cli: string;
	http: string;
	action: string;
}

interface ParsedCli {
	cmd?: string;
	file?: string;
	chart?: string;
	group?: string;
	pattern?: string;
	regex?: string;
	select: string[];
	charts?: string;
	output?: string;
	parser?: string;
	name?: string;
	description?: string;
	tag?: string;
	jsonPath?: string;
	colAxis?: string;
	round?: boolean;
	chartOverrides: string[];
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
	const tokens = tokenize(stripContinuations(cli));
	const parsed: ParsedCli = { select: [], chartOverrides: [], config: {} };

	const pipe = tokens.indexOf('|');
	let i = 0;
	if (pipe !== -1) {
		parsed.cmd = tokens.slice(0, pipe).join(' ');
		i = pipe + 1;
	}

	while (i < tokens.length && !isVizbBin(tokens[i] ?? '')) i += 1;
	if (i < tokens.length) i += 1;

	if (i < tokens.length && CHART_TYPES.has(tokens[i] ?? '')) {
		parsed.chart = tokens[i];
		i += 1;
	}

	while (i < tokens.length) {
		const tok = tokens[i] ?? '';
		if (tok === '|') break;

		const eq = splitEquals(tok);
		if (eq) {
			applyFlag(parsed, eq.name, eq.value);
			i += 1;
			continue;
		}

		if (VALUE_FLAGS[tok]) {
			const value = tokens[i + 1];
			if (value !== undefined && !value.startsWith('-')) {
				applyFlag(parsed, tok, value);
				i += 2;
				continue;
			}
		}

		if (BOOL_FLAGS[tok]) {
			const next = tokens[i + 1];
			if (next === 'true' || next === 'false') {
				applyFlag(parsed, tok, next);
				i += 2;
				continue;
			}
			applyFlag(parsed, tok, 'true');
			i += 1;
			continue;
		}

		if (!tok.startsWith('-') && !parsed.file) {
			parsed.file = tok;
		}
		i += 1;
	}

	return parsed;
}

function applyFlag(parsed: ParsedCli, rawName: string, rawValue: string): void {
	const name = rawName;
	if (VALUE_FLAGS[name] === 'select') {
		parsed.select.push(rawValue);
		return;
	}
	if (VALUE_FLAGS[name] === 'chart') {
		parsed.chartOverrides.push(rawValue);
		return;
	}
	if (name === '--round') {
		parsed.round = rawValue !== 'false';
		return;
	}
	const valueKey = VALUE_FLAGS[name];
	if (valueKey) {
		if (CONFIG_VALUE_KEYS.has(valueKey)) {
			parsed.config[valueKey] = valueKey === 'symbolSize' ? Number(rawValue) : rawValue;
			return;
		}
		(parsed as unknown as Record<string, unknown>)[valueKey] = rawValue;
		return;
	}
	const boolKey = BOOL_FLAGS[name];
	if (boolKey) {
		if (rawValue === 'false') parsed.config[boolKey] = false;
		else parsed.config[boolKey] = rawValue === 'true' ? true : rawValue;
	}
}

function buildHttp(parsed: ParsedCli, input?: string): Record<string, unknown> {
	const types = chartTypes(parsed);
	const body: Record<string, unknown> = {
		input: httpInput(parsed, input),
	};

	const parser = httpParser(parsed);
	if (parser) body.parser = parser;

	const grouping: Record<string, unknown> = {};
	if (parsed.group) grouping.columns = splitGroupColumns(parsed.group);
	if (parsed.pattern) grouping.pattern = parsed.pattern;
	if (parsed.regex) grouping.regex = parsed.regex;
	if (parsed.colAxis) grouping.colAxis = parsed.colAxis;
	if (Object.keys(grouping).length > 0) body.grouping = grouping;

	if (parsed.select.length > 0) body.select = parsed.select;
	if (parsed.jsonPath) body.jsonPath = parsed.jsonPath;
	if (parsed.name) body.name = parsed.name;
	if (parsed.description) body.description = parsed.description;
	if (parsed.tag) body.tag = parsed.tag;
	if (parsed.round) body.round = true;

	if (types.length > 0) {
		const charts: Record<string, unknown> = { types };
		const configs = httpConfigs(parsed, types);
		if (configs.length > 0) charts.configs = configs;
		body.charts = charts;
	}

	body.output = { format: outputFormat(parsed) };
	return body;
}

function buildAction(parsed: ParsedCli): string {
	const withBlock: [string, string][] = [];
	if (parsed.cmd) withBlock.push(['cmd', parsed.cmd]);
	if (parsed.file) withBlock.push(['file', parsed.file]);
	if (parsed.parser && parsed.parser !== 'auto') withBlock.push(['parser', parsed.parser]);
	if (parsed.group) withBlock.push(['group', parsed.group]);
	if (parsed.pattern) withBlock.push(['group-pattern', parsed.pattern]);
	if (parsed.regex) withBlock.push(['group-regex', parsed.regex]);
	if (parsed.select.length === 1) withBlock.push(['select', parsed.select[0] ?? '']);
	else if (parsed.select.length > 1) withBlock.push(['select', parsed.select.join('\n')]);
	if (parsed.jsonPath) withBlock.push(['json-path', parsed.jsonPath]);
	if (parsed.colAxis) withBlock.push(['col-axis', parsed.colAxis]);
	if (parsed.name) withBlock.push(['name', parsed.name]);
	if (parsed.description) withBlock.push(['description', parsed.description]);
	if (parsed.tag) withBlock.push(['tag', parsed.tag]);
	if (parsed.round) withBlock.push(['round', 'true']);

	const types = chartTypes(parsed);
	if (types.length > 0) withBlock.push(['charts', types.join(',')]);

	const chart = actionChart(parsed, types);
	if (chart) withBlock.push(['chart', chart]);

	if (outputFormat(parsed) === 'dataset' && parsed.output) {
		withBlock.push(['output-json', parsed.output]);
	} else if (parsed.output) {
		withBlock.push(['output-html', parsed.output]);
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

function httpInput(parsed: ParsedCli, input?: string): string {
	if (input !== undefined) return input.trimEnd() + '\n';
	if (parsed.file) return `<contents of ${parsed.file}>`;
	if (parsed.cmd) return `<output of ${parsed.cmd}>`;
	return '<input>';
}

function httpParser(parsed: ParsedCli): string | undefined {
	if (parsed.parser) return HTTP_PARSER[parsed.parser] ?? parsed.parser;
	if (parsed.cmd) {
		if (/\bgo\b/.test(parsed.cmd)) return 'go';
		if (/\bcargo\b/.test(parsed.cmd)) return 'rust';
		if (/\b(npx|node|vitest|tinybench)\b/.test(parsed.cmd)) return 'javascript';
	}
	const file = parsed.file ?? '';
	if (file.endsWith('.csv')) return 'csv';
	if (file.endsWith('.json')) return 'json';
	if (file.endsWith('.txt')) return parsed.pattern?.includes('/') ? 'go' : undefined;
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

function httpConfigs(parsed: ParsedCli, types: string[]): Record<string, unknown>[] {
	const byType = new Map<string, Record<string, unknown>>();
	const ensure = (type: string) => {
		let cfg = byType.get(type);
		if (!cfg) {
			cfg = { type };
			byType.set(type, cfg);
		}
		return cfg;
	};

	if (Object.keys(parsed.config).length > 0) {
		const target = parsed.chart ?? types[0];
		if (target) Object.assign(ensure(target), parsed.config);
	}

	for (const override of parsed.chartOverrides) {
		const split = override.indexOf(':');
		if (split === -1) continue;
		const type = override.slice(0, split);
		const props = override.slice(split + 1);
		if (!CHART_TYPES.has(type)) continue;
		const cfg = ensure(type);
		for (const part of splitProps(props)) {
			const eq = part.indexOf('=');
			if (eq === -1) {
				const key = httpConfigKey(part);
				if (key) cfg[key] = true;
			} else {
				const key = httpConfigKey(part.slice(0, eq));
				if (key) cfg[key] = coerceConfigValue(part.slice(eq + 1));
			}
		}
	}

	return [...byType.values()].filter((cfg) => Object.keys(cfg).length > 1);
}

function actionChart(parsed: ParsedCli, types: string[]): string | undefined {
	const lines: string[] = [...parsed.chartOverrides];
	const props: string[] = [];
	for (const [key, value] of Object.entries(parsed.config)) {
		const actionKey = ACTION_CONFIG_KEY[key];
		if (!actionKey) continue;
		if (value === true) props.push(actionKey);
		else props.push(`${actionKey}=${value}`);
	}
	if (props.length > 0) {
		const type = parsed.chart ?? types[0];
		if (type) lines.unshift(`${type}:${props.join(';')}`);
	}
	if (lines.length === 0) return undefined;
	return lines.join('\n');
}

function httpConfigKey(actionKey: string): string | undefined {
	switch (actionKey) {
		case 'labels':
		case 'showLabels':
			return 'showLabels';
		case '3d':
		case 'threeD':
			return 'threeD';
		case '3d-visualmap':
		case 'threeDVisualMap':
			return 'threeDVisualMap';
		case '3d-rotate':
		case 'threeDRotate':
			return 'threeDRotate';
		case 'visualmap':
		case 'visualMap':
			return 'visualMap';
		case 'symbol-size':
		case 'symbolSize':
			return 'symbolSize';
		case 'sort':
		case 'scale':
		case 'stack':
		case 'horizontal':
		case 'swap':
		case 'smooth':
		case 'symbol':
		case 'stat':
			return actionKey;
		default:
			return undefined;
	}
}

function coerceConfigValue(value: string): string | boolean | number {
	if (value === 'true') return true;
	if (value === 'false') return false;
	if (/^\d+(\.\d+)?$/.test(value)) return Number(value);
	return value;
}

function splitProps(props: string): string[] {
	return props
		.split(/[;,]/)
		.map((p) => p.trim())
		.filter(Boolean);
}

function splitGroupColumns(group: string): string[] {
	if (group.includes(',')) return group.split(',').map((c) => c.trim()).filter(Boolean);
	return group.split(/\s+/).filter(Boolean);
}

function outputFormat(parsed: ParsedCli): 'html' | 'dataset' {
	return parsed.output?.toLowerCase().endsWith('.json') ? 'dataset' : 'html';
}

function isVizbBin(token: string): boolean {
	return /(^|[\\/])vizb$/.test(token) || token === 'vizb';
}

function yamlScalar(value: string): string {
	if (value === '') return '""';
	if (/^[A-Za-z0-9_.,/@+-]+$/.test(value)) return value;
	return JSON.stringify(value);
}

function splitEquals(token: string): { name: string; value: string } | undefined {
	if (!token.startsWith('-')) return undefined;
	const eq = token.indexOf('=');
	if (eq <= 1) return undefined;
	return { name: token.slice(0, eq), value: token.slice(eq + 1) };
}

function stripContinuations(src: string): string {
	return src.replace(/\\\r?\n/g, ' ');
}

function tokenize(src: string): string[] {
	const tokens: string[] = [];
	let i = 0;
	const s = src.trim();
	while (i < s.length) {
		const ch = s[i] ?? '';
		if (/\s/.test(ch)) {
			i += 1;
			continue;
		}
		if (ch === '#') {
			while (i < s.length && s[i] !== '\n') i += 1;
			continue;
		}
		if (ch === '"' || ch === "'") {
			const q = ch;
			i += 1;
			let tok = '';
			while (i < s.length && s[i] !== q) {
				tok += s[i];
				i += 1;
			}
			if (s[i] === q) i += 1;
			tokens.push(tok);
			continue;
		}
		let tok = '';
		while (i < s.length && !/\s/.test(s[i] ?? '')) {
			tok += s[i];
			i += 1;
		}
		tokens.push(tok);
	}
	return tokens;
}
