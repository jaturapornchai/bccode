// Build a navigable atlas of Champ: menu -> command -> handler -> form classes -> dialogs -> controls, plus SQL tables.
// Output: atlas.json and one markdown file per top-level menu module (atlas-<n>-<name>.md).
const fs = require('fs');
const path = require('path');
const { ROOT, read } = require('./lib');
const OUT = path.join(__dirname, '../../scratch/champ-atlas'); // gitignored output
fs.mkdirSync(OUT, { recursive: true });

// ---------- dialogs from BC5Account.rc ----------
const rc = read(ROOT + '/BC5Account.rc').split(/\r?\n/);
const KW = new Set(['LTEXT', 'RTEXT', 'CTEXT', 'EDITTEXT', 'COMBOBOX', 'LISTBOX', 'GROUPBOX', 'PUSHBUTTON', 'DEFPUSHBUTTON', 'CHECKBOX', 'AUTOCHECKBOX', 'RADIOBUTTON', 'AUTORADIOBUTTON', 'CONTROL', 'ICON', 'SCROLLBAR', 'STATE3', 'AUTO3STATE']);
const TEXT_FIRST = new Set(['LTEXT', 'RTEXT', 'CTEXT', 'GROUPBOX', 'PUSHBUTTON', 'DEFPUSHBUTTON', 'CHECKBOX', 'AUTOCHECKBOX', 'RADIOBUTTON', 'AUTORADIOBUTTON', 'ICON', 'STATE3', 'AUTO3STATE']);
function splitArgs(s) {
  const out = []; let cur = ''; let q = false;
  for (const ch of s) {
    if (ch === '"') { q = !q; cur += ch; continue; }
    if (ch === ',' && !q) { out.push(cur.trim()); cur = ''; continue; }
    cur += ch;
  }
  if (cur.trim()) out.push(cur.trim());
  return out;
}
const unq = s => (s || '').replace(/^"|"$/g, '');
const dialogs = {};
for (let i = 0; i < rc.length; i++) {
  const h = rc[i].match(/^(\w+)\s+DIALOG(?:EX)?\s+(-?\d+)\s*,\s*(-?\d+)\s*,\s*(\d+)\s*,\s*(\d+)/);
  if (!h) continue;
  const d = { id: h[1], w: +h[4], h: +h[5], caption: '', line: i + 1, controls: [] };
  let j = i + 1;
  for (; j < rc.length && !/^BEGIN/.test(rc[j]); j++) { const c = rc[j].match(/^CAPTION\s+"(.*)"/); if (c) d.caption = c[1]; }
  const stmts = [];
  for (j++; j < rc.length && !/^END/.test(rc[j]); j++) {
    const m = rc[j].match(/^\s+([A-Z0-9]+)\s+(.*)$/);
    if (m && KW.has(m[1])) stmts.push([m[1], m[2]]);
    else if (stmts.length) stmts[stmts.length - 1][1] += ' ' + rc[j].trim();
  }
  for (const [kw, rest] of stmts) {
    const a = splitArgs(rest); let c;
    if (kw === 'CONTROL') c = { type: kw, text: unq(a[0]), ctl: a[1], cls: unq(a[2]), style: a[3], x: +a[4], y: +a[5], w: +a[6], h: +a[7] };
    else if (TEXT_FIRST.has(kw)) c = { type: kw, text: unq(a[0]), ctl: a[1], x: +a[2], y: +a[3], w: +a[4], h: +a[5], style: a.slice(6).join(',') };
    else c = { type: kw, text: '', ctl: a[0], x: +a[1], y: +a[2], w: +a[3], h: +a[4], style: a.slice(5).join(',') };
    if (c.type === 'CONTROL' && /^button$/i.test(c.cls)) c.type = /CHECKBOX/.test(c.style) ? 'CHECKBOX' : /RADIO/.test(c.style) ? 'RADIO' : /GROUPBOX/.test(c.style) ? 'GROUPBOX' : 'BUTTON';
    d.controls.push(c);
  }
  dialogs[d.id] = d; i = j;
}
// key$text labels: resolve Thai for keys whose text was lost ('?') from any other occurrence of the same key
const keyThai = {};
for (const d of Object.values(dialogs)) for (const c of d.controls) {
  const m = c.text.match(/^([A-Z0-9_]+)\$(.*)$/s);
  if (m) { c.key = m[1]; c.label = m[2]; if (/[\u0e00-\u0e7f]/.test(m[2]) && !keyThai[m[1]]) keyThai[m[1]] = m[2]; }
  else c.label = c.text;
}
for (const d of Object.values(dialogs)) for (const c of d.controls) if (c.key && !/[\u0e00-\u0e7f]/.test(c.label) && keyThai[c.key]) c.label = keyThai[c.key] + ' ⟨key⟩';

// ---------- classes from headers ----------
const files = fs.readdirSync(ROOT);
const classes = {};
for (const f of files.filter(f => /\.h$/i.test(f))) {
  const s = read(path.join(ROOT, f));
  for (const m of s.matchAll(/class\s+(C\w+)\s*:\s*public\s+(\w+)([\s\S]*?)\n};/g)) {
    const idd = (m[3].match(/enum\s*\{\s*IDD\s*=\s*(\w+)/) || [])[1];
    classes[m[1]] = { name: m[1], base: m[2], idd, header: f };
  }
}
const cppOf = {};
for (const c of Object.values(classes)) {
  const base = c.header.replace(/\.h$/i, '');
  const f = files.find(x => x.toLowerCase() === (base + '.cpp').toLowerCase());
  if (f) c.cpp = f;
}
const refCache = {};
function refs(cls) {
  if (refCache[cls]) return refCache[cls];
  const c = classes[cls]; refCache[cls] = [];
  if (!c || !c.cpp) return [];
  const s = read(path.join(ROOT, c.cpp));
  const hs = c.header ? read(path.join(ROOT, c.header)) : '';
  const set = new Set();
  for (const m of (s + hs).matchAll(/\b(C[A-Z]\w+)\b/g)) if (m[1] !== cls && classes[m[1]] && (classes[m[1]].idd || /Page|Sheet|Dlg/i.test(m[1]))) set.add(m[1]);
  return (refCache[cls] = [...set]);
}

// ---------- commands ----------
const resH = read(ROOT + '/resource.h');
const byNum = {};
for (const m of resH.matchAll(/#define\s+(\w+)\s+(\d+)/g)) (byNum[m[2]] ||= []).push(m[1]);
const mf = read(ROOT + '/MainFrm.cpp');
const onCmd = {};
for (const m of mf.matchAll(/ON_COMMAND\(\s*(\w+)\s*,\s*(\w+)\s*\)/g)) onCmd[m[1]] = m[2];
function handlerBody(h) {
  const i = mf.search(new RegExp('CMainFrame::' + h + '\\s*\\('));
  if (i < 0) return '';
  const j = mf.indexOf('\n}', i);
  return mf.slice(i, j + 2);
}

// ---------- menu ----------
const mx = read(ROOT + '/menuconfigxml/menuconfig.xml').split(/\r?\n/);
const root = { text: 'ROOT', children: [] }; const stack = [root];
const attr = (l, a) => (l.match(new RegExp(a + '="([^"]*)"')) || [])[1];
for (const l of mx) {
  if (/<popup/.test(l)) { const p = { kind: 'popup', text: attr(l, 'text'), disable: attr(l, 'disable') === 'yes', children: [] }; stack.at(-1).children.push(p); stack.push(p); }
  else if (/<\/popup>/.test(l)) stack.pop();
  else if (/<menuitem/.test(l)) {
    const id = attr(l, 'id'); const text = attr(l, 'text');
    if (text === 'separate') continue;
    const it = { kind: 'item', id, text, disable: attr(l, 'disable') === 'yes' };
    const names = byNum[id] || [];
    it.define = names.find(n => onCmd[n]) || names[0];
    it.handler = it.define && onCmd[it.define];
    const body = it.handler ? handlerBody(it.handler) : '';
    it.forms = [...new Set([...body.matchAll(/RUNTIME_CLASS\((C\w+)\)/g)].map(m => m[1]))];
    it.dlgs = [...new Set([...body.matchAll(/\b(C\w+)\s+\w+\s*(?:\(|;)/g)].map(m => m[1]).filter(n => classes[n] && classes[n].idd))];
    it.calls = [...new Set([...body.matchAll(/\b([A-Z]\w+)\s*\(/g)].map(m => m[1]).filter(n => !/^(IsConnect|IsMenuEnable|RUNTIME_CLASS|CreateOrActivateFrame|InsertRepColCommand|AfxMessageBox|SendMessage)$/.test(n)))].slice(0, 6);
    stack.at(-1).children.push(it);
  }
}

// ---------- SQL tables ----------
const sql = read(ROOT + '/Script/SQLSERVER_Script.sql');
const tables = {};
for (const m of sql.matchAll(/create\s+table\s+(?:\[?dbo\]?\.)?\[?(\w+)\]?[^\n(]*\n?\s*\(([\s\S]*?)\n\s*\)\s*;?/gi)) {
  const cols = [];
  for (const l of m[2].split(/\n/)) {
    const c = l.trim().match(/^\[?(\w+)\]?\s+\[?([A-Za-z]+)\]?\s*(\([^)]*\))?/);
    if (c && !/^(PRIMARY|CONSTRAINT|UNIQUE|FOREIGN)$/i.test(c[1])) cols.push(c[1] + ' ' + c[2].toUpperCase() + (c[3] || ''));
  }
  tables[m[1].toUpperCase()] = cols;
}

// ---------- emit ----------
function expand(cls, depth, seen) {
  if (seen.has(cls) || depth > 2) return [];
  seen.add(cls);
  const c = classes[cls]; if (!c) return [{ cls, missing: true }];
  const node = { cls, base: c.base, idd: c.idd, cpp: c.cpp };
  const out = [node];
  for (const r of refs(cls)) out.push(...expand(r, depth + 1, seen));
  return out;
}
function fmtDialog(idd) {
  const d = dialogs[idd]; if (!d) return `    (dialog ${idd} not found in rc)\n`;
  const ctl = d.controls.filter(c => !(c.type === 'LTEXT' && !c.label) && c.ctl !== 'IDC_STATIC' || c.label);
  ctl.sort((a, b) => (Math.round(a.y / 6) - Math.round(b.y / 6)) || a.x - b.x);
  let s = `    dialog ${idd} ${d.w}x${d.h}${d.caption ? ' caption="' + d.caption + '"' : ''} (BC5Account.rc:${d.line})\n`;
  let row = null;
  for (const c of ctl) {
    const r = Math.round(c.y / 6);
    const lab = c.label ? `"${c.label.replace(/\\n/g, ' ').slice(0, 60)}"` : '';
    const t = c.type === 'CONTROL' ? `CONTROL<${c.cls}>` : c.type;
    s += (r !== row ? '\n      ' : ' | ') + `${t}${lab ? ' ' + lab : ''} ${c.ctl}@${c.x},${c.y}`;
    row = r;
  }
  return s + '\n';
}
const atlas = { menu: root, classes, dialogs, tables, keyThai };
fs.writeFileSync(path.join(OUT, 'atlas.json'), JSON.stringify(atlas));
let n = 0;
const index = [];
for (const mod of root.children) {
  n++;
  let md = `# Champ atlas — module ${n}: ${mod.text}\n\nGenerated by tools/champ-atlas/build.js from D:/project-champ/champ/champ (menuconfig.xml, resource.h, MainFrm.cpp, *.h, BC5Account.rc). Rows of controls are grouped by y (dialog units); "⟨key⟩" = Thai taken from another dialog with the same language key.\n\n`;
  let count = 0;
  const walk = (node, depth) => {
    for (const ch of node.children) {
      if (ch.kind === 'popup') { md += `\n${'#'.repeat(Math.min(depth + 2, 5))} ${ch.text}${ch.disable ? ' (disabled)' : ''}\n`; walk(ch, depth + 1); continue; }
      count++;
      md += `\n- **${ch.id} ${ch.text}**${ch.disable ? ' (disabled)' : ''} — ${ch.define || '?'} → ${ch.handler || '?'}`;
      if (ch.calls.length) md += ` calls: ${ch.calls.join(', ')}`;
      md += '\n';
      const seen = new Set();
      for (const f of [...ch.forms, ...ch.dlgs]) for (const node2 of expand(f, 0, seen)) {
        if (node2.missing) { md += `  - class ${node2.cls} (no header found)\n`; continue; }
        md += `  - class ${node2.cls} : ${node2.base} ${node2.cpp ? '(' + node2.cpp + ')' : ''}${node2.idd ? ' IDD=' + node2.idd : ''}\n`;
        if (node2.idd) md += fmtDialog(node2.idd);
      }
    }
  };
  walk(mod, 0);
  const file = `atlas-${String(n).padStart(2, '0')}-${(mod.text || 'module').replace(/[^\w\u0e00-\u0e7f]+/g, '_')}.md`;
  fs.writeFileSync(path.join(OUT, file), md);
  index.push(`${file}\t${mod.text}\titems=${count}\tbytes=${md.length}`);
}
fs.writeFileSync(path.join(OUT, 'index.tsv'), index.join('\n') + '\n');
console.log(index.join('\n'));
console.log('dialogs', Object.keys(dialogs).length, 'classes', Object.keys(classes).length, 'tables', Object.keys(tables).length, 'keys', Object.keys(keyThai).length);
