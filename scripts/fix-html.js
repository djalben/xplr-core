const fs = require('fs');
const path = './epn-killer-frontend/final_build/index.html';
let html = fs.readFileSync(path, 'utf8');
html = html.replace('<script src=', '<script type="module" src=');
fs.writeFileSync(path, html);
console.log('Fixed index.html: added type="module"');
