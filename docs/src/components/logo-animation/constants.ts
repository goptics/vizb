/** Shared palette (matches hero logo fills). */
export const LOGO_COLORS = [
	'#5470C6',
	'#1A202C',
	'#3BA272',
	'#FC8452',
	'#EE6666',
] as const;

export const HEATMAP_LOW = LOGO_COLORS[0];
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

export type Ease = 'outCubic' | 'outBack' | 'inOutCubic';
export type OverlayMode = 'hidden' | 'heatmap' | 'radar';
export type ColorMode = 'logo' | 'heat-high';

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
	{ selector: '.line-ref', label: 'line', pause: 2000, ease: 'outCubic', overlay: 'hidden' },
	{ selector: '.scatter-ref', label: 'scatter', pause: 1500, ease: 'outBack', overlay: 'hidden' },
	{ selector: '.pie-ref', label: 'pie', pause: 1500, ease: 'inOutCubic', overlay: 'hidden' },
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

/** Overlay opacity triples: [heatmapGrid, radarGrid, radarSeries]. */
export function overlayOpacity(mode: OverlayMode): [number, number, number] {
	if (mode === 'heatmap') return [1, 0, 0];
	if (mode === 'radar') return [0, 1, 1];
	return [0, 0, 0];
}
