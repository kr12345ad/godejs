const {EventEmitter, once} = require('events');

const e = new EventEmitter();
const seen = [];
e.on('x', (v) => seen.push('on:' + v));
e.prependListener('x', (v) => seen.push('pre:' + v));
e.once('x', (v) => seen.push('once:' + v));

console.log(e.listenerCount('x'), e.listeners('x').length, e.eventNames().join(','));
console.log(e.emit('x', 1), e.emit('x', 2));
console.log(seen.join('|'), e.listenerCount('x'));

e.removeAllListeners('x');
console.log(e.listenerCount('x'), e.emit('x', 3));

const err = new Error('boom');
e.once('error', (got) => console.log('caught', got.message));
e.emit('error', err);

(async () => {
  setTimeout(() => e.emit('later', 42), 1);
  const [value] = await once(e, 'later');
  console.log('await', value);
})();
