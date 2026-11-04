const {Readable, Writable, Transform, pipeline, finished} = require('stream');
const zlib = require('zlib');

const written = [];
const sink = new Writable({
  write(chunk, enc, cb) {
    written.push(chunk.toString());
    cb();
  },
});
sink.on('finish', () => console.log('finish', written.join('')));
sink.write('a');
sink.end('b');

const upper = new Transform({
  transform(chunk, enc, cb) {
    cb(null, chunk.toString().toUpperCase());
  },
});
const seen = [];
upper.on('data', (chunk) => seen.push(chunk.toString()));
upper.on('end', () => console.log('transform', seen.join('')));
upper.end('hello');

const piped = [];
const sink2 = new Writable({
  write(chunk, enc, cb) {
    piped.push(chunk.toString());
    cb();
  },
});
finished(sink2, () => console.log('piped', piped.join('')));
pipeline(Readable.from(['x', 'y']), sink2, (err) => {
  if (err) throw err;
});

const gz = zlib.gzipSync(Buffer.from('hello'));
console.log('gzip', gz[0], gz[1], zlib.gunzipSync(gz).toString());
console.log('crc', zlib.crc32('123456789'));
