import { existsSync, readFileSync, statSync } from 'node:fs';
import { CORE_MD_PATHS } from '../src/lib/agent-md.ts';

const dist = new URL('../dist/', import.meta.url);
const llms = readFileSync(new URL('./llms.txt', dist), 'utf8');
if (!llms.includes('text/plain') && llms.includes('<html')) throw new Error('llms.txt looks like HTML');

const full = new URL('./llms-full.txt', dist);
if (!existsSync(full)) throw new Error('missing llms-full.txt');
if (readFileSync(full, 'utf8').includes('<html')) throw new Error('llms-full.txt looks like HTML');

for (const p of CORE_MD_PATHS) {
	if (!llms.includes(p)) throw new Error(`llms.txt missing ${p}`);
	const file = new URL('.' + p, dist);
	if (!existsSync(file) || statSync(file).isDirectory()) {
		throw new Error(`expected real file at dist${p}, not a directory`);
	}
	const body = readFileSync(file, 'utf8');
	if (body.includes('import {')) throw new Error(`${p} still has import {`);
}
