// Shared decoder for Champ sources (UTF-16LE or ANSI bytes holding TIS-620/cp874 Thai)
const fs = require('fs');
const ROOT = 'D:/project-champ/champ/champ';
function decode(buf) {
  const s = (buf[0] === 0xff && buf[1] === 0xfe) ? buf.toString('utf16le') : buf.toString('latin1');
  return s.replace(/[\u00a1-\u00fb]/g, c => String.fromCharCode(c.charCodeAt(0) - 0xa0 + 0x0e00));
}
const read = f => decode(fs.readFileSync(f));
module.exports = { ROOT, decode, read };
