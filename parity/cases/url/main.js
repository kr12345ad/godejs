const {URL, URLSearchParams, pathToFileURL, fileURLToPath, format, resolve} = require('url');

const u = new URL('https://user:pass@example.com:8443/a/b?x=1&y=2#frag');
console.log(u.href);
console.log(u.protocol, u.username, u.password, u.host, u.hostname, u.port, u.pathname, u.search, u.hash);
u.searchParams.append('z', 'a b');
console.log(u.searchParams.getAll('x').join(','), u.searchParams.get('z'));
console.log(u.toString());

const p = new URLSearchParams('a=1&a=2&b=3');
p.delete('a');
p.set('c', '!');
console.log(p.toString(), p.size, p.has('b'));

console.log(pathToFileURL('/tmp/a b').href);
console.log(fileURLToPath('file:///tmp/a%20b'));
console.log(format({protocol: 'https:', host: 'example.com', pathname: '/x', query: {a: '1'}}));
console.log(resolve('https://a/b/c', '../d?x=1'));
