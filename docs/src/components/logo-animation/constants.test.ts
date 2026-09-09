import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import {
	PHASES,
	overlayOpacity,
	type OverlayOpacities,
} from './constants.ts';

const ZERO: OverlayOpacities = {
	heatmap: 0,
	radar: 0,
	series: 0,
	sankey: 0,
	chord: 0,
};

describe('PHASES', () => {
	it('loops bar → sankey → line → scatter → heatmap → pie → chord → radar → logo', () => {
		assert.deepEqual(
			PHASES.map((p) => p.label),
			[
				'bar',
				'sankey',
				'line',
				'scatter',
				'heatmap',
				'pie',
				'chord',
				'radar',
				'logo',
			],
		);
	});

	it('gives sankey and chord overlay + color modes and restores logo colors on line and pie', () => {
		const byLabel = Object.fromEntries(PHASES.map((p) => [p.label, p]));
		assert.equal(byLabel.sankey.overlay, 'sankey');
		assert.equal(byLabel.sankey.colors, 'sankey');
		assert.equal(byLabel.sankey.pause, 2000);
		assert.equal(byLabel.sankey.ease, 'outCubic');
		assert.equal(byLabel.heatmap.overlay, 'heatmap');
		assert.equal(byLabel.heatmap.colors, 'heat-high');
		assert.equal(byLabel.heatmap.pause, 1500);
		assert.equal(byLabel.chord.overlay, 'chord');
		assert.equal(byLabel.chord.colors, 'chord');
		assert.equal(byLabel.chord.pause, 2000);
		assert.equal(byLabel.chord.ease, 'inOutCubic');
		assert.equal(byLabel.line.colors, 'logo');
		assert.equal(byLabel.line.pause, 2000);
		assert.equal(byLabel.line.overlay, 'hidden');
		assert.equal(byLabel.pie.colors, 'logo');
		assert.equal(byLabel.pie.pause, 2500);
		assert.equal(byLabel.pie.overlay, 'hidden');
	});
});

describe('overlayOpacity', () => {
	it('raises only the active overlay group', () => {
		assert.deepEqual(overlayOpacity('hidden'), ZERO);
		assert.deepEqual(overlayOpacity('heatmap'), { ...ZERO, heatmap: 1 });
		assert.deepEqual(overlayOpacity('radar'), {
			...ZERO,
			radar: 1,
			series: 1,
		});
		assert.deepEqual(overlayOpacity('sankey'), { ...ZERO, sankey: 1 });
		assert.deepEqual(overlayOpacity('chord'), { ...ZERO, chord: 1 });
	});
});
