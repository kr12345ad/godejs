const path = require('path');

const widget = require('widget');
console.log(widget.tag, widget.format(2));
console.log(require('widget/feature').tag);
console.log(require('widget/lib/extra').tag);
console.log(require('#private'));
console.log(path.basename(require.resolve('widget')));
