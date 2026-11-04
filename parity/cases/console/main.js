const util = require('util');

console.log({a: 1, b: 'two', c: [1, 2, 3], d: {e: true}, f: null, g: undefined});
console.log([1, 'a', null, undefined, [2, [3]]]);
console.log(new Map([['a', 1], ['b', 2]]));
console.log(new Set([1, 2, 3]));
console.log(Buffer.from('hi'), Buffer.from([0, 255]));
console.log(new Date(0));
console.log(/a+b/gi);
console.log(function foo() {});
console.log(class Bar {});
console.log({nested: {deep: {deeper: {deepest: {x: 1}}}}});
console.log({a: 1}, {depth: null});
console.log('fmt %s %d %j %o %%', 'x', 42, {a: 1}, [1]);
console.log(util.format('%s:%d', 'a', 1));
console.log(util.inspect({x: {y: {z: {w: 1}}}}, {depth: 1}));
console.table([{a: 1, b: 2}, {a: 3, b: 4}]);
