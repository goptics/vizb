/**
 * Hero logo morph animation boot.
 * Starts as soon as the hero is visible (no artificial idle delay).
 * animejs still loads async so it does not block LCP.
 */

const PATH_COUNT = 5;

function waitVisible(el: Element): Promise<void> {
	return new Promise((resolve) => {
		if (typeof IntersectionObserver === 'undefined') {
			resolve();
			return;
		}
		// Already on screen (typical for homepage hero) → resolve immediately.
		const rect = el.getBoundingClientRect();
		if (rect.bottom > 0 && rect.top < window.innerHeight) {
			resolve();
			return;
		}
		const io = new IntersectionObserver(
			(entries) => {
				if (entries.some((e) => e.isIntersecting)) {
					io.disconnect();
					resolve();
				}
			},
			{ rootMargin: '50px' },
		);
		io.observe(el);
	});
}

export async function initLogoAnimation(): Promise<void> {
	const hero = document.querySelector('.vizb-hero');
	const vizPaths = [...document.querySelectorAll('.vizb-hero .viz-path')];

	if (
		!(hero instanceof SVGElement) ||
		vizPaths.length !== PATH_COUNT ||
		window.matchMedia('(prefers-reduced-motion: reduce)').matches
	) {
		return;
	}

	await waitVisible(hero);

	// Destructuring on the await expression enables animejs tree-shaking.
	const { createTimeline, svg } = await import('animejs');
	const { startMorphLoop, bindPlayback } = await import('./timeline');

	const tl = startMorphLoop({ createTimeline, svg }, hero, vizPaths);
	bindPlayback(tl, hero);
}

void initLogoAnimation();
