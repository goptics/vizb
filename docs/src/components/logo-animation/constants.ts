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
	{
		selector: '.heatmap-ref',
		label: 'heatmap',
		pause: 1500,
		ease: 'inOutCubic',
		overlay: 'heatmap',
		colors: 'heat-high',
	},
	{
		selector: '.pie-ref',
		label: 'pie',
		pause: 2500,
		ease: 'inOutCubic',
		overlay: 'hidden',
		colors: 'logo',
	},
	{
		selector: '.chord-ref',
		label: 'chord',
		pause: 2000,
		ease: 'inOutCubic',
		overlay: 'chord',
		colors: 'chord',
	},
	{
		selector: '.radar-ref',
		label: 'radar',
		pause: 2000,
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
