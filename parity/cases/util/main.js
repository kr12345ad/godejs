const util = require('util');

console.log(util.format('%s:%d:%i:%f:%j:%%', 'a', 1.9, '2.9', '3.5', {x: 1}));
console.log(util.inspect([1, 'a'], {colors: false, compact: true}));
console.log(util.inspect({a: {b: {c: 1}}}, {depth: 1}));
console.log(util.types.isDate(new Date()), util.types.isPromise(Promise.resolve()));
console.log(util.isDeepStrictEqual({a: [1]}, {a: [1]}), util.isDeepStrictEqual({a: 1}, {a: '1'}));

const twice = util.promisify((x, cb) => cb(null, x * 2));
twice(21).then((v) => console.log('promisified', v));

console.log(Buffer.from('hé').toString('hex'), Buffer.from('aGk=', 'base64').toString());
console.log(Buffer.alloc(4).fill(7).join(','), Buffer.concat([Buffer.from('a'), Buffer.from('b')]).toString());
console.log(Buffer.byteLength('hé'), Buffer.isBuffer(Buffer.from('x')));
console.log(Buffer.from([0x61, 0x62]).compare(Buffer.from('ab')));
console.log(new util.TextDecoder().decode(new util.TextEncoder().encode('hé')));
