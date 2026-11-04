const crypto = require('crypto');

console.log(crypto.createHash('sha256').update('abc').digest('hex'));
console.log(crypto.createHash('md5').update('abc').digest('base64'));
console.log(crypto.createHash('sha1').update('abc').digest('hex'));
console.log(crypto.createHmac('sha256', 'key').update('abc').digest('hex'));
console.log(crypto.createHmac('sha1', 'key').update('abc').digest('base64'));

// Random values have stable shapes, not stable bytes.
console.log(crypto.randomUUID().length, crypto.randomUUID().split('-').length);
console.log(crypto.randomBytes(8).length, crypto.randomFillSync(Buffer.alloc(4)).length);
console.log(crypto.randomInt(1, 10) >= 1);

console.log(crypto.timingSafeEqual(Buffer.from('abc'), Buffer.from('abc')));
console.log(crypto.timingSafeEqual(Buffer.from('abc'), Buffer.from('abd')));
console.log(crypto.getHashes().includes('sha256'), crypto.getCiphers().includes('aes-256-cbc'));
