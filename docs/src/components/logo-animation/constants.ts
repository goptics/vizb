/** Shared palette (matches hero logo fills). */
export const LOGO_COLORS = [
	'#5470C6',
	'#1A202C',
	'#3BA272',
	'#FC8452',
	'#EE6666',
] as const;

export const HEATMAP_LOW = LOGO_COLORS[2];
export const HEATMAP_HIGH = LOGO_COLORS[4];
export const GRID_GRAY = '#4a5568';

export const DUR = 400;
export const OVERLAY_DUR = 200;
export const STAGGER = 30;

export const HEAT = {
	rows: 3,
	cols: 5,
	cell: 54,
	gap: 1,
	padX: 12,
	/** Highlight cells rendered as morph targets (red); grid fills the rest. */
	redCells: [
		[0, 2],
		[0, 4],
		[1, 1],
		[2, 0],
		[2, 3],
	] as const,
} as const;

export const RADAR = {
	cx: 150,
	cy: 150,
	maxR: 108,
	series: [
		[95, 72, 108, 65, 88],
		[88, 95, 60, 100, 75],
		[70, 55, 90, 80, 105],
		[100, 78, 72, 92, 58],
		[62, 102, 85, 68, 95],
	] as const,
} as const;

export const SANKEY_COLORS = [
	'#5470C6',
	'#3BA272',
	'#FC8452',
	'#EE6666',
	'#73C0DE',
] as const;

export const CHORD_COLORS = [
	'#5470C6',
	'#3BA272',
	'#FC8452',
	'#73C0DE',
	'#EE6666',
] as const;

export type OverlayPath = {
	d: string;
	fill: string;
	opacity?: number;
};

export const SANKEY_RIBBONS: readonly OverlayPath[] = [
	{
		d: 'M36 48 C53.6 48, 60.4 48, 78 48 L78 252 C60.4 252, 53.6 252, 36 252 Z',
		fill: '#5470C6',
		opacity: 0.45,
	},
	{
		d: 'M92 48 C167.6 48, 196.4 48, 272 48 L272 203 C196.4 203, 167.6 203, 92 203 Z',
		fill: '#3BA272',
		opacity: 0.32,
	},
	{
		d: 'M92 203 C115.5 203, 124.5 198, 148 198 L148 252 C124.5 252, 115.5 252, 92 252 Z',
		fill: '#3BA272',
		opacity: 0.4,
	},
	{
		d: 'M160 198 C181 198, 189 198, 210 198 L210 234 C189 234, 181 234, 160 234 Z',
		fill: '#FC8452',
		opacity: 0.45,
	},
	{
		d: 'M160 234 C207 234, 225 210, 272 210 L272 220 C225 220, 207 252, 160 252 Z',
		fill: '#FC8452',
		opacity: 0.32,
	},
	{
		d: 'M220 198 C241.8 198, 250.2 224, 272 224 L272 236 C250.2 236, 241.8 216, 220 216 Z',
		fill: '#EE6666',
		opacity: 0.38,
	},
	{
		d: 'M220 216 C241.8 216, 250.2 240, 272 240 L272 252 C250.2 252, 241.8 234, 220 216 Z',
		fill: '#EE6666',
		opacity: 0.38,
	},
];

export const SANKEY_EXTRA_NODES: readonly OverlayPath[] = [
	{ d: 'M272 210H284V220H272Z', fill: '#FAC858' },
	{ d: 'M272 224H284V236H272Z', fill: '#E87EA1' },
	{ d: 'M272 240H284V252H272Z', fill: '#9A60B4' },
];

export const CHORD_RIBBONS: readonly OverlayPath[] = [
	{
		d: 'M150.00 38.00A112.0 112.0 0 0 1 186.46 44.10Q150.0 150.0 259.22 125.21A112.0 112.0 0 0 1 261.34 162.11Q150.0 150.0 150.00 38.00Z',
		fill: '#5470C6',
		opacity: 0.38,
	},
	{
		d: 'M186.46 44.10A112.0 112.0 0 0 1 205.75 52.86Q150.0 150.0 165.47 260.93A112.0 112.0 0 0 1 144.30 261.85Q150.0 150.0 186.46 44.10Z',
		fill: '#5470C6',
		opacity: 0.38,
	},
	{
		d: 'M261.34 162.11A112.0 112.0 0 0 1 243.56 211.57Q150.0 150.0 144.30 261.85A112.0 112.0 0 0 1 93.90 246.94Q150.0 150.0 261.34 162.11Z',
		fill: '#3BA272',
		opacity: 0.38,
	},
	{
		d: 'M230.29 228.09A112.0 112.0 0 0 1 198.12 251.14Q150.0 150.0 39.23 166.58A112.0 112.0 0 0 1 40.38 127.02Q150.0 150.0 230.29 228.09Z',
		fill: '#3BA272',
		opacity: 0.38,
	},
	{
		d: 'M93.90 246.94A112.0 112.0 0 0 1 76.65 234.64Q150.0 150.0 243.56 211.57A112.0 112.0 0 0 1 230.29 228.09Q150.0 150.0 93.90 246.94Z',
		fill: '#FC8452',
		opacity: 0.38,
	},
	{
		d: 'M76.65 234.64A112.0 112.0 0 0 1 70.83 229.22Q150.0 150.0 40.38 127.02A112.0 112.0 0 0 1 42.29 119.30Q150.0 150.0 76.65 234.64Z',
		fill: '#FC8452',
		opacity: 0.38,
	},
	{
		d: 'M48.72 102.18A112.0 112.0 0 0 1 66.19 75.70Q150.0 150.0 236.06 78.32A112.0 112.0 0 0 1 252.70 105.32Q150.0 150.0 48.72 102.18Z',
		fill: '#EE6666',
		opacity: 0.38,
	},
	{
		d: 'M66.19 75.70A112.0 112.0 0 0 1 71.68 69.94Q150.0 150.0 198.12 251.14A112.0 112.0 0 0 1 190.82 254.30Q150.0 150.0 66.19 75.70Z',
		fill: '#EE6666',
		opacity: 0.38,
	},
	{
		d: 'M51.92 204.08A112.0 112.0 0 0 1 39.23 166.58Q150.0 150.0 205.75 52.86A112.0 112.0 0 0 1 236.06 78.32Q150.0 150.0 51.92 204.08Z',
		fill: '#73C0DE',
		opacity: 0.38,
	},
	{
		d: 'M42.29 119.30A112.0 112.0 0 0 1 46.67 106.79Q150.0 150.0 70.83 229.22A112.0 112.0 0 0 1 62.02 219.31Q150.0 150.0 42.29 119.30Z',
		fill: '#73C0DE',
		opacity: 0.38,
	},
	{
		d: 'M83.59 59.81A112.0 112.0 0 0 1 94.71 52.60Q150.0 150.0 62.02 219.31A112.0 112.0 0 0 1 54.45 208.43Q150.0 150.0 83.59 59.81Z',
		fill: '#FAC858',
		opacity: 0.38,
	},
	{
		d: 'M99.15 50.21A112.0 112.0 0 0 1 113.79 44.01Q150.0 150.0 252.70 105.32A112.0 112.0 0 0 1 258.00 120.32Q150.0 150.0 99.15 50.21Z',
		fill: '#9A60B4',
		opacity: 0.38,
	},
	{
		d: 'M113.79 44.01A112.0 112.0 0 0 1 134.40 39.09Q150.0 150.0 190.82 254.30A112.0 112.0 0 0 1 170.44 260.12Q150.0 150.0 113.79 44.01Z',
		fill: '#9A60B4',
		opacity: 0.38,
	},
	{
		d: 'M134.40 39.09A112.0 112.0 0 0 1 144.96 38.11Q150.0 150.0 71.68 69.94A112.0 112.0 0 0 1 79.60 62.89Q150.0 150.0 134.40 39.09Z',
		fill: '#9A60B4',
		opacity: 0.38,
	},
];

export const CHORD_EXTRA_ARCS: readonly OverlayPath[] = [
	{
		d: 'M74.10 46.93A128 128 0 0 1 86.81 38.68L94.71 52.60A112 112 0 0 0 83.59 59.81Z',
		fill: '#FAC858',
	},
	{
		d: 'M91.88 35.95A128 128 0 0 1 144.24 22.13L144.96 38.11A112 112 0 0 0 99.15 50.21Z',
		fill: '#9A60B4',
	},
];

export type Ease = 'outCubic' | 'outBack' | 'inOutCubic';
export type OverlayMode = 'hidden' | 'heatmap' | 'radar' | 'sankey' | 'chord';
export type ColorMode = 'logo' | 'heat-high' | 'sankey' | 'chord';

export type OverlayOpacities = {
	heatmap: number;
	radar: number;
	series: number;
	sankey: number;
	chord: number;
};

/** Phase metadata only — DOM refs are resolved at runtime by selector. */
export type PhaseDef = {
	selector: string;
	label: string;
	pause: number;
	ease: Ease;
	overlay: OverlayMode;
	colors?: ColorMode;
};

/**
 * `pause` = hold the *previous* shape before morphing into this one.
 * First phase pause is the initial logo hold after boot.
 * Logo rest between cycles lives in LOGO_HOLD_MS.
 */
export const PHASES: readonly PhaseDef[] = [
	{ selector: '.bar-ref', label: 'bar', pause: 1500, ease: 'outCubic', overlay: 'hidden' },
	{
		selector: '.sankey-ref',
		label: 'sankey',
		pause: 2000,
		ease: 'outCubic',
		overlay: 'sankey',
		colors: 'sankey',
	},
	{
		selector: '.line-ref',
		label: 'line',
		pause: 2000,
		ease: 'outCubic',
		overlay: 'hidden',
		colors: 'logo',
	},
	{ selector: '.scatter-ref', label: 'scatter', pause: 1500, ease: 'outBack', overlay: 'hidden' },
	{ selector: '.pie-ref', label: 'pie', pause: 1500, ease: 'inOutCubic', overlay: 'hidden' },
	{
		selector: '.chord-ref',
		label: 'chord',
		pause: 2000,
		ease: 'inOutCubic',
		overlay: 'chord',
		colors: 'chord',
	},
	{
		selector: '.heatmap-ref',
		label: 'heatmap',
		pause: 2000,
		ease: 'inOutCubic',
		overlay: 'heatmap',
		colors: 'heat-high',
	},
	{
		selector: '.radar-ref',
		label: 'radar',
		pause: 2500,
		ease: 'inOutCubic',
		overlay: 'radar',
		colors: 'logo',
	},
	{
		selector: '.logo-ref',
		label: 'logo',
		pause: 2500,
		ease: 'inOutCubic',
		overlay: 'hidden',
		colors: 'logo',
	},
];

/** Hold on the logo after the full cycle before looping back to bar. */
export const LOGO_HOLD_MS = 2000;

const OVERLAY_OFF: OverlayOpacities = {
	heatmap: 0,
	radar: 0,
	series: 0,
	sankey: 0,
	chord: 0,
};

export function overlayOpacity(mode: OverlayMode): OverlayOpacities {
	if (mode === 'heatmap') return { ...OVERLAY_OFF, heatmap: 1 };
	if (mode === 'radar') return { ...OVERLAY_OFF, radar: 1, series: 1 };
	if (mode === 'sankey') return { ...OVERLAY_OFF, sankey: 1 };
	if (mode === 'chord') return { ...OVERLAY_OFF, chord: 1 };
	return OVERLAY_OFF;
}
