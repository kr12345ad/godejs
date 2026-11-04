// End-to-end check for the Node-API Go ABI. Run with:
//   ./gode internal/napi/testdata/test.js
const addon = require('/tmp/gode-napi-addon.node');

let failures = 0;
function eq(name, got, want) {
  const ok = Object.is(got, want);
  if (!ok) {
    failures++;
    console.log('FAIL ' + name + ': got ' + String(got) + ' want ' + String(want));
  } else {
    console.log('ok   ' + name);
  }
}

eq('add', addon.add(2, 3), 5);
eq('utf8', addon.greet('world'), 'hello world');

const info = addon.info();
eq('object prop', info.name, 'gode');
eq('object number', info.n, 42);
eq('accessor', info.computed, 7);

const counter = new addon.Counter(5);
eq('ctor new.target', counter.isNew, true);
eq('ctor state', counter.value, 5);
eq('method', counter.inc(), 6);
eq('method state', counter.value, 6);
eq('getter', counter.valueStr, 'value=6');
eq('instanceof', counter instanceof addon.Counter, true);
eq('instanceof call', addon.instanceCheck(counter, addon.Counter), true);

const buf = Buffer.from([1, 2, 3]);
eq('buffer length', addon.bump(buf), 3);
eq('buffer write', buf[0], 2);
eq('buffer write tail', buf[2], 4);

eq('typed array', addon.sumF64(new Float64Array([1.5, 2.5])), 4);
eq('bigint', addon.bigAdd(2n, 3n), 5n);
eq('date roundtrip', addon.dateMs(1234), 1234);

const wrapped = addon.wrapMe();
eq('unwrap', addon.unwrapMe(wrapped), 99);

let threw = null;
try {
  addon.boom();
} catch (e) {
  threw = e;
}
eq('throw message', threw && threw.message, 'boom!');
eq('throw code', threw && threw.code, 'EBOOM');
eq('throw is Error', threw instanceof Error, true);

addon.promiseValue().then((v) => eq('promise', v, 'resolved'));
addon.asyncDouble(21).then((v) => eq('async work', v, 42));
addon.tsfnEmit((v) => eq('threadsafe function', v, 7), 7);

setTimeout(() => {
  console.log(failures === 0 ? 'ALL PASS' : 'FAILURES=' + failures);
  process.exit(failures === 0 ? 0 : 1);
}, 400);
