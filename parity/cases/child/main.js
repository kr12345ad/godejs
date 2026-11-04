const {execSync, execFileSync, spawnSync} = require('child_process');

console.log(execSync('printf hi', {encoding: 'utf8'}));
console.log(execFileSync('/bin/echo', ['ok'], {encoding: 'utf8'}).trim());

const r = spawnSync('/bin/sh', ['-c', 'printf 42']);
console.log(r.status, r.signal, r.stdout.toString());
console.log(spawnSync('/bin/sh', ['-c', 'exit 7']).status);

try {
  execSync('exit 3', {encoding: 'utf8'});
} catch (err) {
  console.log(err.status, String(err.signal), err.message);
}
