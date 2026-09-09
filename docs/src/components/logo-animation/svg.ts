import {
	CHORD_EXTRA_ARCS,
	CHORD_RIBBONS,
	GRID_GRAY,
	HEAT,
	HEATMAP_LOW,
	LOGO_COLORS,
	RADAR,
	SANKEY_EXTRA_NODES,
	SANKEY_RIBBONS,
	type OverlayPath,
} from './constants';

const SVG_NS = 'http://www.w3.org/2000/svg';

export function svgEl(
	name: string,
	attrs: Record<string, string | number> = {},
): SVGElement {
	const node = document.createElementNS(SVG_NS, name);
	for (const [key, value] of Object.entries(attrs)) {
		node.setAttribute(key, String(value));
	}
	return node;
}

function heatPadY(): number {
	return (
		(300 - HEAT.rows * HEAT.cell - (HEAT.rows - 1) * HEAT.gap) / 2
	);
}

function heatCellRect(row: number, col: number) {
	return {
		x: HEAT.padX + col * (HEAT.cell + HEAT.gap),
		y: heatPadY() + row * (HEAT.cell + HEAT.gap),
		size: HEAT.cell,
	};
}

function radarPoint(index: number, radius: number) {
	const angle = ((-90 + index * 72) * Math.PI) / 180;
	return {
		x: RADAR.cx + radius * Math.cos(angle),
		y: RADAR.cy + radius * Math.sin(angle),
	};
}

function pointsAttr(
	count: number,
	radiusAt: (i: number) => number,
): string {
	return Array.from({ length: count }, (_, i) => {
		const p = radarPoint(i, radiusAt(i));
		return `${p.x},${p.y}`;
	}).join(' ');
}

/** Low-intensity heatmap cells (highlights come from morph targets). */
export function buildHeatmapGrid(): SVGElement {
	const group = svgEl('g', { class: 'heatmap-grid', opacity: 0 });
	const red = new Set(HEAT.redCells.map(([r, c]) => `${r},${c}`));

	for (let row = 0; row < HEAT.rows; row++) {
		for (let col = 0; col < HEAT.cols; col++) {
			if (red.has(`${row},${col}`)) continue;
			const { x, y, size } = heatCellRect(row, col);
			group.appendChild(
				svgEl('rect', {
					x,
					y,
					width: size,
					height: size,
					fill: HEATMAP_LOW,
					stroke: HEATMAP_LOW,
					'stroke-width': 1,
				}),
			);
		}
	}
	return group;
}

export function buildRadarGrid(): SVGElement {
	const group = svgEl('g', { class: 'radar-grid', opacity: 0 });

	for (const scale of [0.25, 0.5, 0.75, 1]) {
		group.appendChild(
			svgEl('polygon', {
				points: pointsAttr(5, () => RADAR.maxR * scale),
				fill: 'none',
				stroke: GRID_GRAY,
				'stroke-width': 1,
				opacity: 0.55,
			}),
		);
	}

	for (let i = 0; i < 5; i++) {
		const tip = radarPoint(i, RADAR.maxR);
		group.appendChild(
			svgEl('line', {
				x1: RADAR.cx,
				y1: RADAR.cy,
				x2: tip.x,
				y2: tip.y,
				stroke: GRID_GRAY,
				'stroke-width': 1,
				opacity: 0.55,
			}),
		);
	}
	return group;
}

export function buildRadarSeries(): SVGElement {
	const group = svgEl('g', { class: 'radar-series', opacity: 0 });

	RADAR.series.forEach((radii, seriesIdx) => {
		group.appendChild(
			svgEl('polygon', {
				points: pointsAttr(5, (i) => radii[i]!),
				fill: LOGO_COLORS[seriesIdx],
				'fill-opacity': 0.12,
				stroke: LOGO_COLORS[seriesIdx],
				'stroke-width': 2,
				'stroke-linejoin': 'round',
			}),
		);
	});
	return group;
}

function overlayPaths(className: string, items: readonly OverlayPath[]): SVGElement {
	const group = svgEl('g', { class: className, opacity: 0 });
	for (const item of items) {
		const attrs: Record<string, string | number> = {
			d: item.d,
			fill: item.fill,
		};
		if (item.opacity !== undefined) attrs.opacity = item.opacity;
		group.appendChild(svgEl('path', attrs));
	}
	return group;
}

export function buildSankeyOverlay(): SVGElement {
	const group = overlayPaths('sankey-overlay', SANKEY_RIBBONS);
	for (const node of SANKEY_EXTRA_NODES) {
		group.appendChild(svgEl('path', { d: node.d, fill: node.fill }));
	}
	return group;
}

export function buildChordOverlay(): SVGElement {
	const group = overlayPaths('chord-overlay', CHORD_RIBBONS);
	for (const arc of CHORD_EXTRA_ARCS) {
		group.appendChild(svgEl('path', { d: arc.d, fill: arc.fill }));
	}
	return group;
}

/** Ensure overlay layers exist once under the hero SVG. */
export function ensureOverlays(hero: SVGElement): {
	heatmap: Element | null;
	radar: Element | null;
	series: Element | null;
	sankey: Element | null;
	chord: Element | null;
} {
	if (!hero.querySelector('.heatmap-grid')) {
		hero.insertBefore(buildHeatmapGrid(), hero.firstChild);
		hero.insertBefore(buildRadarGrid(), hero.firstChild);
		hero.insertBefore(buildRadarSeries(), hero.firstChild);
	}
	if (!hero.querySelector('.sankey-overlay')) {
		hero.insertBefore(buildSankeyOverlay(), hero.firstChild);
	}
	if (!hero.querySelector('.chord-overlay')) {
		hero.insertBefore(buildChordOverlay(), hero.firstChild);
	}
	return {
		heatmap: hero.querySelector('.heatmap-grid'),
		radar: hero.querySelector('.radar-grid'),
		series: hero.querySelector('.radar-series'),
		sankey: hero.querySelector('.sankey-overlay'),
		chord: hero.querySelector('.chord-overlay'),
	};
}
