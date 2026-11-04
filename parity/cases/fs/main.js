const fs = require('fs');
const path = require('path');

fs.rmSync('tmp', {recursive: true, force: true});
fs.mkdirSync('tmp/a/b', {recursive: true});
fs.writeFileSync('tmp/a/b/x.txt', 'hello');
fs.appendFileSync('tmp/a/b/x.txt', ' world');
fs.copyFileSync('tmp/a/b/x.txt', 'tmp/a/b/y.txt');
fs.renameSync('tmp/a/b/y.txt', 'tmp/a/b/z.txt');

console.log(fs.readFileSync('tmp/a/b/x.txt', 'utf8'));
console.log(fs.statSync('tmp/a/b/x.txt').size);
console.log(fs.readdirSync('tmp/a/b').sort().join(','));
console.log(fs.existsSync('tmp/a/b/z.txt'), fs.existsSync('tmp/a/nope.txt'));
console.log(fs.readFileSync('data.json', 'base64'));
console.log(JSON.parse(fs.readFileSync('data.json', 'utf8')).name);

fs.rmSync('tmp', {recursive: true, force: true});
console.log(fs.existsSync('tmp'));
