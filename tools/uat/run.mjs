// One-command UAT: `npm run uat` (local) — Demo login → API CRUD flows with a psql check after every
// operation → UI smoke of every implemented menu screen (console, API errors, screenshot) → cleanup by exact id.
// Prints ≤ ~30 lines; full detail in tmp/uat/summary.json. Exit code 1 when anything fails.
//
//   npm run uat                                   all flows + all screens (local)
//   npm run uat -- --flows=gl --screens=none      GL flow only
//   npm run uat -- --flows=erp --erp=sale-invoice,purchase
//   npm run uat -- --flows=none --screens=gl,ap --dark
//   npm run uat -- --flows=none --screens=all --routes=/branch,/gl/dimensions   only these screens, in this order
//   npm run uat -- --rebuild                     rebuild + restart local mainapi from the working tree first
//   npm run uat -- --target=prod                read-only: UI smoke on RUNGRUENG/01 only, no CRUD, no psql
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { makeApi, makePsql, sqlText } from './lib.mjs';

const args = Object.fromEntries(process.argv.slice(2).map((a) => a.replace(/^--/, '').split('=')).map(([k, v]) => [k, v ?? 'true']));
const PROD = args.target === 'prod';
const FE = (args.fe ?? (PROD ? 'https://account.bcaicloud.com' : 'http://localhost:3000')).replace(/\/$/, '');
const BE = (args.be ?? 'http://127.0.0.1:8888').replace(/\/$/, '');
const DB = args.db ?? 'rungrueng';
const OUT = path.resolve('tmp/uat');
const flows = PROD ? [] : (args.flows ?? 'gl,erp').split(',').filter((f) => f && f !== 'none');
const screens = (args.screens ?? 'all').split(',').filter((s) => s && s !== 'none');
const docdate = new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Bangkok' }).format(new Date());
const started = Date.now();

fs.mkdirSync(OUT, { recursive: true });
const steps = [];
const recorder = (flow) => ({ check: (name, ok, detail = '') => (steps.push({ flow, name, ok: !!ok, detail: String(detail).slice(0, 240) }), !!ok) });
const pre = recorder('preflight');

// ── Preflight + Demo login (same BFF calls as the login page)
const reach = async (url) => fetch(url, { signal: AbortSignal.timeout(15000) }).then((r) => r.status).catch((e) => e.message);
const feStatus = await reach(FE + '/');
pre.check('frontend', feStatus === 200, `${FE} → ${feStatus}`);
let psql = null;
if (!PROD && flows.length) {
  try {
    psql = makePsql({ db: DB });
    pre.check('postgres', psql('select 1') === '1', `container postgres db ${DB}`);
  } catch (e) {
    pre.check('postgres', false, e.message);
  }
  if (args.rebuild === 'true') {
    const compose = ['compose', '-f', 'docker-compose.yml', '-f', 'docker-compose.local.yml', '-f', 'docker-compose.dev.yml'];
    try {
      execFileSync('docker', [...compose, 'build', 'mainapi'], { cwd: 'backend', stdio: 'pipe', timeout: 600000 });
      execFileSync('docker', [...compose, 'up', '-d', '--no-deps', 'mainapi'], { cwd: 'backend', stdio: 'pipe', timeout: 120000 });
      for (let i = 0; i < 60 && (await reach(BE + '/healthz')) !== 200; i++) await new Promise((r) => setTimeout(r, 1000));
    } catch (e) {
      pre.check('rebuild', false, String(e.stderr ?? e.message).slice(-200));
    }
  }
  pre.check('backend', (await reach(BE + '/healthz')) === 200, BE);
  // A container older than the newest backend commit/edit tests old code — fail fast and say how to fix it.
  const built = Date.parse(execFileSync('docker', ['inspect', 'mainapi', '--format', '{{.Created}}'], { encoding: 'utf8' }).trim());
  const committed = Number(execFileSync('git', ['log', '-1', '--format=%ct', '--', 'backend'], { encoding: 'utf8' }).trim()) * 1000;
  const edited = execFileSync('git', ['status', '--porcelain', '--', 'backend'], { encoding: 'utf8' })
    .split('\n')
    .map((l) => l.slice(3).trim())
    .filter((f) => f && !f.endsWith('_test.go') && fs.existsSync(f)) // tests are not in the image
    .reduce((t, f) => Math.max(t, fs.statSync(f).mtimeMs), 0);
  pre.check('backend-fresh', built >= Math.max(committed, edited), `mainapi built ${new Date(built).toISOString()}, backend changed ${new Date(Math.max(committed, edited)).toISOString()} → rerun with --rebuild`);
}
let token = '';
if (flows.length) {
  const login = await fetch(FE + '/api/auth/demo-login', { method: 'POST' }).then(async (r) => ({ status: r.status, json: await r.json().catch(() => null) }));
  token = login.json?.token ?? '';
  pre.check('demo-login', login.status === 200 && token, `${login.status} (needs BCAI_DEMO_LOGIN_ENABLED=true)`);
  if (token) {
    const sel = await makeApi(FE, { Authorization: `Bearer ${token}` })('POST', '/api/workspace/select-holding', { holdingcode: DB, businesscode: '01', branchuid: '00000' });
    pre.check('select-holding', sel.status === 200 && sel.json?.success, `${DB}/01/00000 → ${sel.status}`);
  }
}

// ── API flows
const cleanup = [];
const ready = steps.every((s) => s.ok);
if (ready && flows.length) {
  const auth = { Authorization: `Bearer ${token}`, 'Accept-Language': 'th' };
  const ctx = { fe: makeApi(FE, auth), api: makeApi(BE, auth), psql, recorder, docdate, cleanup, only: args.erp ? new Set(args.erp.split(',')) : null };
  for (const flow of flows) {
    try {
      await (await import(`./flows/${flow === 'erp' ? 'erp-docs' : flow}.mjs`)).run(ctx);
    } catch (e) {
      recorder(flow).check('EXCEPTION', false, e.stack ?? e.message);
    }
  }
}

// ── UI smoke
let shots = 0;
if (screens.length && feStatus === 200) {
  const { runUiSmoke } = await import('./ui.mjs');
  shots = await runUiSmoke({ fe: FE, out: OUT, heads: screens, routes: args.routes?.split(','), dark: args.dark === 'true', recorder });
}

// ── Cleanup: only rows this run created, matched by exact id, only when nothing posted from them
const removed = { gl: 0, erp: 0, kept: 0 };
if (psql && args.keep !== 'true') {
  for (const c of cleanup) {
    try {
      if (c.kind.startsWith('gl-')) {
        const kind = { 'gl-journal': 'journals', 'gl-account': 'accounts', 'gl-journal-book': 'journal-books' }[c.kind];
        const out = psql(`delete from gl_records r where company='01' and kind='${kind}' and id=${sqlText(c.id)} and coalesce(payload->>'isdeleted','false')='true' and not exists (select 1 from gl_lines l where l.company='01' and l.journal_id=r.id) returning 1`);
        if (out) removed.gl++; else removed.kept++;
      } else if (c.kind === 'erp-doc') {
        const docno = sqlText(c.docno);
        const out = psql(`with d as (delete from doc where id=${Number(c.id)} and docno=${docno} and isdelete and not exists (select 1 from gl_lines where doc_no=${docno}) and not exists (select 1 from stock_ledger where docno=${docno}) returning docno), x as (delete from docdetail where docno in (select docno from d) returning 1) select count(*) from d`);
        if (out === '1') removed.erp++; else removed.kept++;
      }
    } catch {
      removed.kept++;
    }
  }
}

// ── Summary
const failed = steps.filter((s) => !s.ok);
const byFlow = {};
for (const s of steps) {
  const key = s.flow.split(':')[0];
  byFlow[key] ??= { pass: 0, fail: 0 };
  byFlow[key][s.ok ? 'pass' : 'fail']++;
}
fs.mkdirSync(OUT, { recursive: true }); // another tool may have wiped the folder mid-run
const seconds = Math.round((Date.now() - started) / 1000);
fs.writeFileSync(path.join(OUT, 'summary.json'), JSON.stringify({ target: FE, db: DB, docdate, seconds, byFlow, removed, failed, steps }, null, 1));
console.log(`UAT ${PROD ? 'PROD read-only' : 'local'} ${FE} · ${seconds}s · checks ${steps.length - failed.length}/${steps.length} pass · screenshots ${shots}`);
console.log(Object.entries(byFlow).map(([k, v]) => `${k} ${v.pass}/${v.pass + v.fail}`).join(' · '));
if (cleanup.length) console.log(`cleanup: removed gl=${removed.gl} erp=${removed.erp}, kept ${removed.kept} (posted/cancelled history stays)`);
for (const f of failed.slice(0, 20)) console.log(`FAIL ${f.flow} ${f.name}: ${f.detail.replace(/\s+/g, ' ').slice(0, 150)}`);
if (failed.length > 20) console.log(`… ${failed.length - 20} more failures in tmp/uat/summary.json`);
process.exitCode = failed.length ? 1 : 0;
