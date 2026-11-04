import def, {add, label} from './math.mjs';
import * as path from 'node:path';
import {URL} from 'node:url';

console.log(add(2, 3), def(4, 5), label);
console.log(path.join('a', 'b'));
console.log(new URL('https://example.com/a b').href);
console.log(import.meta.url.startsWith('file://'));
console.log(import.meta.filename.endsWith('main.mjs'));

const {default: lazy} = await import('./lazy.mjs');
console.log(lazy());
