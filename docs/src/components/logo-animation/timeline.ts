import {
	DUR,
	HEATMAP_HIGH,
	LOGO_COLORS,
	LOGO_HOLD_MS,
	OVERLAY_DUR,
	PHASES,
	STAGGER,
	overlayOpacity,
	type ColorMode,
} from './constants';
import { ensureOverlays } from './svg';

/** Minimal animejs surface we use — avoids pulling the full package types into this module. */
type AnimeSvg = { morphTo: (target: Element) => unknown };
type Anime = {
	createTimeline: (opts: { loop: boolean }) => Timeline;
	svg: AnimeSvg;
};

type Timeline = {
	set: (target: Element, props: object, at: number | string) => Timeline;
	add: (target: Element, props: object, at: number | string) => Timeline;
	label: (name: string, at: string) => Timeline;
	play: () => void;
	pause: () => void;
};

function pathFill(
	mode: ColorMode | undefined,
	index: number,
): { fill: string; stroke: string } | null {
	if (mode === 'heat-high') {
		return { fill: HEATMAP_HIGH, stroke: HEATMAP_HIGH };
	}
	if (mode === 'logo') {
		return { fill: LOGO_COLORS[index]!, stroke: LOGO_COLORS[index]! };
	}
	return null;
}

function fade(
	tl: Timeline,
	node: Element | null,
	opacity: number,
	at: string,
) {
	if (!node) return;
	tl.add(node, { opacity, duration: OVERLAY_DUR, ease: 'inOutCubic' }, at);
}

/**
 * Build the looping morph timeline across chart shapes.
 * Call only after animejs is loaded.
 */
export function startMorphLoop(
	anime: Anime,
	hero: SVGElement,
	vizPaths: Element[],
): Timeline {
	const { createTimeline, svg } = anime;
	const logoRefs = [...document.querySelectorAll('.logo-ref')];
	const overlays = ensureOverlays(hero);
	const tl = createTimeline({ loop: true });

	vizPaths.forEach((p, i) => {
		tl.set(p, { d: svg.morphTo(logoRefs[i]!) }, 0);
	});
	for (const node of [overlays.heatmap, overlays.radar, overlays.series]) {
		if (node) tl.set(node, { opacity: 0 }, 0);
	}

	for (const phase of PHASES) {
		const refs = [...document.querySelectorAll(phase.selector)];
		tl.label(phase.label, `+=${phase.pause}`);

		const [hm, rd, sr] = overlayOpacity(phase.overlay);
		fade(tl, overlays.heatmap, hm, phase.label);
		fade(tl, overlays.radar, rd, phase.label);
		fade(tl, overlays.series, sr, phase.label);

		vizPaths.forEach((p, i) => {
			const at = `${phase.label}+=${i * STAGGER}`;
			tl.add(
				p,
				{ d: svg.morphTo(refs[i]!), duration: DUR, ease: phase.ease },
				at,
			);
			const color = pathFill(phase.colors, i);
			if (color) {
				tl.add(p, { ...color, duration: DUR, ease: phase.ease }, at);
			}
		});
	}

	// Rest on the logo at end of cycle (replaces the old 2s delay before first morph).
	tl.label('logo-hold', `+=${LOGO_HOLD_MS}`);

	return tl;
}

/** Pause when the tab is hidden or the hero leaves the viewport. */
export function bindPlayback(tl: Timeline, hero: Element) {
	let heroVisible = true;

	const sync = () => {
		if (heroVisible && document.visibilityState === 'visible') tl.play();
		else tl.pause();
	};

	document.addEventListener('visibilitychange', sync);

	if (typeof IntersectionObserver === 'undefined') return;
	const io = new IntersectionObserver(
		(entries) => {
			heroVisible = entries.some((e) => e.isIntersecting);
			sync();
		},
		{ threshold: 0 },
	);
	io.observe(hero);
}
