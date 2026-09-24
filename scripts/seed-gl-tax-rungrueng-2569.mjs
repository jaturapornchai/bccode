/**
 * ข้อมูลบัญชี + ภาษี ก.ค.–ก.ย. 2569 ของ rungrueng / 01 / 00000 ผ่าน BFF เดียวกับหน้าจอ (คู่มือ: docs/kms/22-sample-data-gl-tax-2569.md)
 *   node scripts/seed-gl-tax-rungrueng-2569.mjs            # dry-run: อ่านอย่างเดียว บอกว่าจะสร้างอะไร
 *   node scripts/seed-gl-tax-rungrueng-2569.mjs --apply    # เขียนจริง (idempotent ตาม docno / แบบ+งวด)
 *   node scripts/seed-gl-tax-rungrueng-2569.mjs verify     # อ่านรายงาน VAT/WHT + ดาวน์โหลดไฟล์ → tmp/sample-data/out/<host>/
 * env: SEED_BASE (frontend origin, ค่าเริ่มต้น http://127.0.0.1:3000), SEED_MANIFEST (ทับ path manifest),
 *      SEED_MEDIA_REF_NO (เลขอ้างอิงการลงทะเบียนยื่นด้วยสื่อที่กรมสรรพากรออกให้ — ค่าเริ่มต้นว่าง)
 * token จาก demo-login อยู่ในหน่วยความจำเท่านั้น ห้ามพิมพ์/บันทึก
 */
import { randomUUID } from 'node:crypto';
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { ACCOUNTS, CERT_SCENARIOS, FILINGS, HEAD_OFFICE, JOURNALS, PARTNERS, RD_FILE_CODES, SCOPE, SIGNER, TAX_SECTION } from './seed-gl-tax-rungrueng-2569.data.mjs';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const BASE = (process.env.SEED_BASE || 'http://127.0.0.1:3000').replace(/\/+$/, '');
const HOST = new URL(BASE).hostname;
const args = process.argv.slice(2);
const MODE = args.find((a) => !a.startsWith('--')) ?? 'seed';
const APPLY = args.includes('--apply');
const unknown = args.filter((a) => a !== MODE && a !== '--apply');
if (!['seed', 'verify'].includes(MODE) || unknown.length || (MODE === 'verify' && APPLY)) {
  console.error('usage: node scripts/seed-gl-tax-rungrueng-2569.mjs [seed [--apply] | verify]');
  process.exit(2);
}
const MEDIA_REF_NO = (process.env.SEED_MEDIA_REF_NO ?? '').trim();
if (MEDIA_REF_NO && !/^\d{1,20}$/.test(MEDIA_REF_NO)) throw new Error('SEED_MEDIA_REF_NO must be digits only, at most 20 (Format กลาง USER_ID)');
const { holding: HOLDING, company: COMPANY, branch: BRANCH, fiscalYear: FY, taxYear: TAX_YEAR } = SCOPE;

// ---------- manifest: สิ่งที่สคริปต์สร้าง + สมุดที่เลือกไว้ (อยู่ใต้ tmp/ ซึ่ง gitignore) ----------
const MANIFEST = process.env.SEED_MANIFEST || path.join(ROOT, 'tmp', 'sample-data', `manifest-${HOST}.json`);
const manifest = existsSync(MANIFEST) ? JSON.parse(readFileSync(MANIFEST, 'utf8')) : { base: BASE };
manifest.books ??= {}; manifest.journals ??= {}; manifest.accounts ??= []; manifest.filings ??= [];
function save() {
  if (!APPLY) return;
  mkdirSync(path.dirname(MANIFEST), { recursive: true });
  writeFileSync(MANIFEST, `${JSON.stringify(manifest, null, 2)}\n`);
}

// ---------- เงิน: สตางค์ BigInt เท่านั้น ----------
function sat(value) {
  const s = String(value ?? '0').trim() || '0';
  if (!/^-?\d+(\.\d{1,2})?$/.test(s)) throw new Error(`not a 2-decimal amount: ${s}`);
  const neg = s.startsWith('-'); const [i, f = ''] = s.replace('-', '').split('.');
  const v = BigInt(i) * 100n + BigInt((f + '00').slice(0, 2));
  return neg ? -v : v;
}
const baht = (v) => `${v < 0n ? '-' : ''}${(v < 0n ? -v : v) / 100n}.${String((v < 0n ? -v : v) % 100n).padStart(2, '0')}`;
// เลขประจำตัวผู้เสียภาษี 13 หลัก: หลักที่ 13 = mod 11 แบบเดียวกับ whtcert.ValidThaiTaxID
const taxId = (twelve) => { let s = 0; for (let i = 0; i < 12; i++) s += Number(twelve[i]) * (13 - i); return twelve + ((11 - (s % 11)) % 10); };
const partner = (code) => { const { tax_id_base, ...p } = PARTNERS[code]; return { partner_code: code, ...p, tax_id: taxId(tax_id_base) }; };

// ---------- HTTP: demo-login → select-holding → call() ----------
let headers = null;
async function call(method, urlPath, body, { raw = false } = {}) {
  const res = await fetch(BASE + urlPath, { method, headers, body: body === undefined ? undefined : JSON.stringify(body) });
  if (raw) return res;
  const text = await res.text();
  let json; try { json = JSON.parse(text); } catch { json = { raw: text.slice(0, 300) }; }
  if (!res.ok || json.success === false) {
    const err = new Error(`${method} ${urlPath} → ${res.status} ${JSON.stringify(json).slice(0, 600)}`);
    err.status = res.status; throw err;
  }
  return json;
}
async function login() {
  const res = await fetch(`${BASE}/api/auth/demo-login`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: '{}' });
  const json = await res.json().catch(() => ({}));
  if (!res.ok || typeof json.token !== 'string' || !json.token) throw new Error(`demo-login failed (HTTP ${res.status}) — needs BCAI_DEMO_LOGIN_ENABLED=true`);
  headers = { Authorization: `Bearer ${json.token}`, 'x-bc-backend-url': `${BASE}/backend/goapi`, 'Content-Type': 'application/json', 'Accept-Language': 'th' };
  await call('POST', '/api/workspace/select-holding', { holdingcode: HOLDING, businesscode: COMPANY, branchuid: BRANCH });
}
async function listAll(resource) {
  const out = [];
  for (let page = 1; page <= 200; page++) {
    const data = (await call('GET', `/api/gl/${resource}?q=&page=${page}&limit=500`)).data ?? {};
    const items = Array.isArray(data.items) ? data.items : [];
    out.push(...items);
    if (!items.length || out.length >= (data.total ?? out.length)) return out;
  }
  throw new Error(`${resource}: pagination did not finish`);
}
const command = (cmd) => call('POST', '/api/gl/command', { requestid: randomUUID(), reason: '', date: '', docno: '', ...cmd });
const tax = (op, body) => call('POST', `/api/goapi/api/report/tax/form/${op}`, { holdingcode: HOLDING, businesscode: COMPANY, ...body });

// ---------- สมุดรายวัน: เลือกครั้งเดียวต่อ booktype แล้วจำไว้ใน manifest (ห้ามเลือกด้วยรหัสสมุด) ----------
// เปลี่ยนสมุดภายหลัง = docno ใหม่ = ใบซ้ำ จึงใช้ลำดับ: manifest → สมุดประเภทนั้นที่มีใบของชุดนี้อยู่แล้ว → สมุดที่ใช้มากที่สุด
async function resolveBooks(journals) {
  const books = (await listAll('journal-books')).filter((b) => b.isactive !== false);
  const have = new Set(journals.map((j) => `${j.bookcode ?? ''}|${j.docno}`));
  const usage = new Map(); for (const j of journals) usage.set(j.bookcode, (usage.get(j.bookcode) ?? 0) + 1);
  const chosen = {};
  for (const t of [...new Set(JOURNALS.map((s) => s.t))]) {
    const saved = manifest.books[t];
    if (saved) {
      if (!books.some((b) => b.code === saved && b.booktype === t)) throw new Error(`manifest book ${saved} (booktype ${t}) is missing/inactive/retyped — fix it by hand; a new book would duplicate journals`);
      chosen[t] = { code: saved, why: 'manifest' }; continue;
    }
    const cands = books.filter((b) => b.booktype === t);
    if (!cands.length) throw new Error(`no active journal book with booktype ${t}`);
    const nos = JOURNALS.filter((s) => s.t === t).map((s) => s.no);
    const seeded = cands.map((b) => [b.code, nos.filter((no) => have.has(`${b.code}|${b.code}${no}`)).length]).filter(([, n]) => n > 0).sort((a, b) => b[1] - a[1]);
    if (seeded.length > 1 && seeded[0][1] === seeded[1][1]) throw new Error(`booktype ${t}: books ${seeded.map(([c]) => c).join('/')} both hold this batch — set manifest.books[${t}] by hand`);
    if (seeded.length) { chosen[t] = { code: seeded[0][0], why: `already holds ${seeded[0][1]}/${nos.length} journals of this batch` }; continue; }
    cands.sort((a, b) => (usage.get(b.code) ?? 0) - (usage.get(a.code) ?? 0) || a.code.localeCompare(b.code));
    chosen[t] = { code: cands[0].code, why: `most-used active book (${usage.get(cands[0].code) ?? 0} journals)` };
  }
  return chosen;
}

// ---------- ใบรายวัน ----------
function buildJournal(s, bookcode) {
  let dr = 0n, cr = 0n;
  const lines = s.lines.map((l) => { dr += sat(l.dr); cr += sat(l.cr); return { accountcode: l.a, description: l.desc, debit: l.dr ?? '0', credit: l.cr ?? '0', departmentcode: '', projectcode: '', cashflow: '' }; });
  if (dr !== cr) throw new Error(`${s.key} unbalanced ${baht(dr)} vs ${baht(cr)}`);
  const codes = new Set([...(s.partners ?? []), ...(s.vats ?? []).map((v) => v.partner), ...(s.whts ?? []).map((w) => w.partner)].filter(Boolean));
  const details = {};
  if (codes.size) details.partners = [...codes].map(partner);
  if (s.vats) details.vats = s.vats.map(({ partner: code, ...v }) => {
    const p = code ? partner(code) : null;
    if (sat(v.vat_amount) !== (sat(v.base_amount) * 7n + 50n) / 100n) throw new Error(`${s.key} VAT is not 7% of the base`);
    return { id: randomUUID(), original_invoice_no: '', original_invoice_date: '', tax_period_year: Number(s.date.slice(0, 4)), tax_period_month: Number(s.date.slice(5, 7)), partner_code: p?.partner_code ?? '', partner_tax_id: p?.tax_id ?? '', partner_branch_no: p?.tax_branch_no ?? '', partner_name: p?.name_th ?? '', zero_rate_amount: '0', exempt_amount: '0', vat_rate: '7', claim_reason: '', remark: '', ...v };
  });
  if (s.whts) details.withholdings = s.whts.map(({ partner: code, ...w }) => {
    if (sat(w.tax_amount) * 100n !== sat(w.base_amount) * BigInt(w.wht_rate)) throw new Error(`${s.key} WHT amount ≠ ${w.wht_rate}% of base`);
    return { id: randomUUID(), partner_code: code, remark: '', ...w };
  });
  return { docno: `${bookcode}${s.no}`, date: s.date, bookcode, fiscalyear: FY, description: s.desc, reference: s.ref, branchcode: BRANCH, kind: 'manual', lines, ...(Object.keys(details).length ? { details } : {}) };
}

// ใบในระบบตรงกับใบที่ scenario จะสร้างไหม (หัวใบ + บรรทัด + VAT + ภาษีหัก) — ใช้รับใบร่างของชุดนี้ที่ manifest ไม่รู้จักกลับมาผ่านรายการ
// id ของ VAT/ภาษีหักเป็น UUID สุ่มจึงไม่เทียบ; เงินเทียบเป็นสตางค์ (ตัดศูนย์ท้ายเกิน 2 ตำแหน่ง, อ่านไม่ออก = ไม่ตรง)
function sameContent(have, want) {
  const amt = (v) => { const s = String(v ?? '0').trim().replace(/(\.\d{2})0+$/, '$1'); try { return String(sat(s)); } catch { return `?${s}`; } };
  const lines = (j) => (j.lines ?? []).map((l) => `${l.accountcode}|${amt(l.debit)}|${amt(l.credit)}`).join(';');
  const vats = (j) => (j.details?.vats ?? []).map((v) => `${v.tax_type}|${v.tax_invoice_no}|${amt(v.base_amount)}|${amt(v.vat_amount)}`).join(';');
  const whts = (j) => (j.details?.withholdings ?? []).map((w) => `${w.wht_direction}|${w.form_type}|${w.wht_cert_no}|${amt(w.base_amount)}|${amt(w.tax_amount)}`).join(';');
  return ['date', 'bookcode', 'branchcode', 'description', 'reference'].every((k) => String(have[k] ?? '') === String(want[k] ?? ''))
    && lines(have) === lines(want) && vats(have) === vats(want) && whts(have) === whts(want);
}

// ---------- ทะเบียนบริษัท: ที่อยู่สำนักงานใหญ่ (เติมเมื่อว่างเท่านั้น) ----------
async function ensureCompanyAddress() {
  const companyPath = `/backend/organization/company/${encodeURIComponent(COMPANY)}`;
  const company = (await call('GET', companyPath)).data;
  if (!company || typeof company !== 'object' || !('address' in company)) { console.log('registry has no address fields yet — filings keep the form header address'); return; }
  const current = company.address ?? {};
  if (Object.keys(HEAD_OFFICE.address).some((k) => String(current[k] ?? '').trim())) { console.log('registry address present — kept'); return; }
  const body = { ...company, address: { ...current, ...HEAD_OFFICE.address }, phone: String(company.phone ?? '').trim() || HEAD_OFFICE.phone };
  if (!APPLY) { console.log('would set registry address (blank now):', Object.values(HEAD_OFFICE.address).filter(Boolean).join(' '), body.phone); return; }
  await call('PUT', companyPath, body);
  console.log('registry address set');
}

// ---------- แบบยื่น: prefill + หัวแบบเฉพาะช่องที่ schema ของแบบนั้นมี ----------
const schemaCache = new Map();
async function formKeys(code) {
  if (!schemaCache.has(code)) schemaCache.set(code, new Set(((await tax('schema', { code })).data?.fields ?? []).map((f) => f.key)));
  return schemaCache.get(code);
}
function applyHeader(code, values, keys) {
  const set = [];
  // ที่อยู่/โทรศัพท์: prefill ดึงจากทะเบียนบริษัท/ฉบับก่อนมาทั้งชุด — เติม HEAD_OFFICE เฉพาะเมื่อช่อง addr_* ว่างทุกช่อง (backend เก่าที่ทะเบียนยังไม่มีที่อยู่)
  // ห้ามเติมทีละช่อง: ที่อยู่อื่นที่ไม่มีหมู่/เบอร์ + "หมู่ 7"/เบอร์ของที่อยู่นี้ = ที่อยู่ผสมสองแห่ง (กฎเดียวกับ applyRegistryAddress ใน tax_form_fill.go)
  const addrKeys = new Set([...Object.keys(values), ...Object.keys(HEAD_OFFICE.address)].filter((k) => k.startsWith('addr_')));
  if ([...addrKeys].every((k) => !String(values[k] ?? '').trim())) {
    for (const [k, v] of Object.entries(HEAD_OFFICE.address)) if (v && keys.has(k)) { values[k] = v; set.push(k); }
    if (keys.has('phone') && !String(values.phone ?? '').trim()) { values.phone = HEAD_OFFICE.phone; set.push('phone'); }
  }
  const header = { ...SIGNER, ...(TAX_SECTION[code] ? { tax_section: TAX_SECTION[code] } : {}), ...(MEDIA_REF_NO ? { media_ref_no: MEDIA_REF_NO } : {}) };
  for (const [k, v] of Object.entries(header)) if (keys.has(k)) { values[k] = v; set.push(k); }
  return set;
}

async function seed() {
  console.log(`SEED ${BASE} ${HOLDING}/${COMPANY}/${BRANCH} FY${FY} ${APPLY ? '(APPLY)' : '(dry-run — nothing is written)'} manifest=${path.relative(ROOT, MANIFEST)}`);
  const journals = await listAll('journals');
  const books = await resolveBooks(journals);
  for (const [t, b] of Object.entries(books)) { console.log(`book type ${t}: ${b.code} (${b.why})`); manifest.books[t] = b.code; }

  // บัญชี: ตรวจครบทั้งรายการ (อ่านอย่างเดียว) ก่อนสร้างบัญชีแรก — ผังที่ขาดบัญชี/บัญชีพ่อหยุดตรงนี้โดยยังไม่เขียนอะไรเลย
  const accounts = new Map((await listAll('accounts')).map((a) => [a.accountcode, a]));
  const planned = new Set();
  const toCreate = [];
  for (const a of ACCOUNTS) {
    const have = accounts.get(a.accountcode);
    if (have) { if (a.allowposting !== false && !have.allowposting) throw new Error(`account ${a.accountcode} exists but is not a posting account`); continue; }
    if (!a.name) throw new Error(`account ${a.accountcode} is missing and has no definition — this seed needs the rungrueng chart of accounts`);
    if (!accounts.has(a.parentaccountcode) && !planned.has(a.parentaccountcode)) throw new Error(`account ${a.accountcode}: parent ${a.parentaccountcode} is missing`);
    planned.add(a.accountcode);
    toCreate.push({ accountcode: a.accountcode, names: [{ code: 'th', name: a.name }], accounttype: a.accounttype, normalbalance: a.normalbalance, allowposting: a.allowposting !== false, isactive: true, level: a.level, parentaccountcode: a.parentaccountcode });
  }
  save(); // สมุดที่เลือก — บันทึกหลังตรวจบัญชีผ่าน
  for (const account of toCreate) {
    console.log(`${APPLY ? 'create' : 'would create'} account ${account.accountcode} ${account.names[0].name}`);
    if (APPLY) { const r = await command({ resource: 'accounts', action: 'create', id: '', version: 0, account }); manifest.accounts.push({ code: account.accountcode, id: r.data?.id ?? r.id }); save(); }
  }

  const byDocno = new Map(journals.map((j) => [j.docno, j]));
  const n = { created: 0, posted: 0, existing: 0, planned: 0 };
  const notPosted = []; // ใบของชุดที่จะยังไม่ผ่านรายการหลังรอบนี้ — แบบยื่นอ่านเฉพาะใบที่ผ่านรายการ จึงห้ามบันทึกแบบยื่นถ้ามี
  for (const s of JOURNALS) {
    const journal = buildJournal(s, books[s.t].code);
    const existing = byDocno.get(journal.docno);
    let mine = manifest.journals[journal.docno];
    if (existing) {
      n.existing++;
      // ใบร่างที่ไม่อยู่ใน manifest (คำตอบ create หาย / สร้างจากเครื่องอื่น) = ของชุดนี้เมื่อเนื้อหาตรง scenario ทุกช่อง
      if (existing.status === 'draft' && !mine && sameContent((await call('GET', `/api/gl/journals/${encodeURIComponent(existing.id)}`)).data ?? {}, journal)) {
        mine = { id: existing.id, version: existing.version, status: 'draft', scenario: s.key, recovered: true };
        if (APPLY) { manifest.journals[journal.docno] = mine; save(); }
      }
      if (existing.status === 'draft' && mine) { // ใบของเราที่สร้างแล้วแต่ยังไม่ผ่านรายการ (รอบก่อนหยุดกลางทาง)
        console.log(`${APPLY ? 'post' : 'would post'} existing draft ${journal.docno}${mine.recovered ? ' (not in manifest, content matches this batch)' : ''}`);
        if (APPLY) { const p = await command({ resource: 'journals', action: 'post', id: existing.id, version: existing.version }); Object.assign(mine, { status: 'posted', version: p.data?.version }); save(); n.posted++; }
      } else {
        console.log(`existing ${journal.docno} ${existing.status} — skip`);
        if (existing.status !== 'posted') notPosted.push(`${journal.docno} (${existing.status}${existing.status === 'draft' ? ', content differs from this batch' : ''})`);
      }
      continue;
    }
    if (!APPLY) { n.planned++; console.log(`would create+post ${journal.docno} [${s.key}] ${journal.lines.length} lines ${journal.details ? Object.keys(journal.details).join('+') : ''}`); continue; }
    const r = await command({ resource: 'journals', action: 'create', id: '', version: 0, journal });
    manifest.journals[journal.docno] = { id: r.data?.id, version: r.data?.version, status: 'draft', scenario: s.key, seededat: new Date().toISOString() }; save(); n.created++;
    const p = await command({ resource: 'journals', action: 'post', id: r.data?.id, version: r.data?.version });
    Object.assign(manifest.journals[journal.docno], { status: 'posted', version: p.data?.version }); save(); n.posted++;
    console.log(`created+posted ${journal.docno} [${s.key}]`);
  }
  console.log('journals', n);

  const filings = (await tax('list', { year: TAX_YEAR })).data ?? [];
  const filed = (code, month) => filings.find((f) => f.code === code && f.year === TAX_YEAR && f.month === month && (f.filingseq ?? 0) === 0);
  const missing = FILINGS.filter(([code, month]) => !filed(code, month)).map(([code, month]) => `${code} ${TAX_YEAR}-${month}`);
  if (notPosted.length) {
    // prefill อ่านเฉพาะใบที่ผ่านรายการ และแบบที่บันทึกแล้วไม่ถูกเติมใหม่ — บันทึกแบบยื่นตอนนี้ = แบบยื่นขาดใบเหล่านี้ถาวร จึงหยุดก่อนเขียนอะไรต่อ
    const msg = `batch journals not posted: ${notPosted.join(', ')} — post or fix them on the journal screen first`;
    if (missing.length) {
      if (APPLY) throw new Error(`${msg}; filings ${missing.join(', ')} are not saved until every batch journal is posted`);
      console.log(`--apply would stop here (filings ${missing.join(', ')} not saved): ${msg}`);
      return;
    }
    console.log(`warning: ${msg} (all filings already saved — not rewritten)`);
  }

  await ensureCompanyAddress();
  for (const [code, month] of FILINGS) {
    const have = filed(code, month);
    if (have) { console.log(`filing existing ${code} ${TAX_YEAR}-${month} id ${have.id} v${have.version} — skip`); continue; }
    const doc = (await tax('prefill', { code, year: TAX_YEAR, month })).data ?? {};
    doc.values ??= {};
    const set = applyHeader(code, doc.values, await formKeys(code));
    if (!APPLY) { console.log(`would save filing ${code} ${TAX_YEAR}-${month} (header: ${set.join(',') || '-'})`); continue; }
    const r = await tax('save', { code, year: TAX_YEAR, month, document: doc });
    manifest.filings.push({ code, year: TAX_YEAR, month, id: r.data?.id, version: r.data?.version }); save();
    console.log(`saved filing ${code} ${TAX_YEAR}-${month} id ${r.data?.id}`);
  }
}

// ---------- verify (อ่านอย่างเดียว) ----------
const rowsOf = (data) => (Array.isArray(data) ? data : data?.records ?? data?.rows ?? []);
const total = (rows, key) => baht(rows.reduce((s, r) => s + sat(r[key]), 0n));
async function download(label, urlPath, body, fallbackName, outDir) {
  const res = await call('POST', urlPath, body, { raw: true });
  const buf = Buffer.from(await res.arrayBuffer());
  if (!res.ok) { console.log(`${label}: HTTP ${res.status} ${buf.toString('utf8').slice(0, 300)}`); return { label, status: res.status }; }
  const name = (res.headers.get('content-disposition') ?? '').match(/filename="?([^";]+)"?/)?.[1] ?? fallbackName;
  writeFileSync(path.join(outDir, path.basename(name)), buf);
  console.log(`${label}: ${res.status} ${res.headers.get('content-type')} ${buf.length}B → ${path.basename(name)} (starts ${JSON.stringify(buf.subarray(0, 5).toString('latin1'))})`);
  return { label, status: res.status, file: path.basename(name), bytes: buf.length };
}

async function verify() {
  const outDir = path.join(ROOT, 'tmp', 'sample-data', 'out', HOST);
  mkdirSync(outDir, { recursive: true });
  console.log(`VERIFY ${BASE} (read-only) → ${path.relative(ROOT, outDir)}`);
  const co = { holdingcode: HOLDING, businesscode: COMPANY };
  const report = { base: BASE, at: new Date().toISOString(), vat: [], wht: [], files: [] };
  for (const month of [7, 8, 9]) {
    for (const type of ['purchase', 'sale']) {
      const rows = rowsOf((await call('POST', '/api/goapi/api/report/tax/vat-register', { ...co, year: TAX_YEAR, month, type })).data);
      const line = { month, type, count: rows.length, base: total(rows, 'amountbeforevat'), vat: total(rows, 'vatamount'), invoices: rows.map((r) => r.taxinvoiceno) };
      report.vat.push(line); console.log(`VAT ${type} ${month}: ${line.count} rows base ${line.base} vat ${line.vat} [${line.invoices.join(' ')}]`);
    }
    for (const direction of ['paid', 'received']) {
      const rows = rowsOf((await call('POST', '/api/goapi/api/report/tax/wht', { ...co, year: TAX_YEAR, month, direction })).data);
      const line = { month, direction, count: rows.length, base: total(rows, 'baseamount'), wht: total(rows, 'whtamount'), items: rows.map((r) => `${r.formtype ?? ''}:${r.ratepercent ?? ''}%:${r.certificateno ?? ''}`) };
      report.wht.push(line); console.log(`WHT ${direction} ${month}: ${line.count} rows base ${line.base} wht ${line.wht} [${line.items.join(' ')}]`);
    }
  }
  const filings = (await tax('list', { year: TAX_YEAR })).data ?? [];
  for (const [code, month] of FILINGS) {
    const f = filings.find((x) => x.code === code && x.year === TAX_YEAR && x.month === month && (x.filingseq ?? 0) === 0);
    if (!f) { console.log(`filing ${code} ${month}: not saved`); report.files.push({ label: `${code}-${month}`, status: 'missing' }); continue; }
    if (RD_FILE_CODES.includes(code)) report.files.push(await download(`RD file ${code} ${month}`, '/api/goapi/api/report/tax/form/rdfile', { ...co, id: f.id, version: f.version, code, year: TAX_YEAR, month, rdfile: {} }, `${code}-${TAX_YEAR}-${month}.txt`, outDir));
    const cur = (await tax('load', { id: f.id })).data;
    report.files.push(await download(`form PDF ${code} ${month}`, '/api/goapi/api/report/tax/form/pdf', { ...co, id: cur.id, code: cur.code, year: cur.year, month: cur.month, document: cur.document }, `${code}-${TAX_YEAR}-${String(month).padStart(2, '0')}.pdf`, outDir));
  }
  // 50 ทวิ: หาใบจาก scenario key → สมุดที่เลือก (manifest หรือเลือกใหม่แบบอ่านอย่างเดียว) → id ใน manifest หรือรายการจริง
  const journals = await listAll('journals');
  const books = await resolveBooks(journals);
  const byDocno = new Map(journals.map((j) => [j.docno, j]));
  const o = HEAD_OFFICE.address;
  const payerAddress = `${o.addr_no} หมู่ ${o.addr_moo} ถนน${o.addr_road} ตำบล${o.addr_subdistrict} อำเภอ${o.addr_district} จังหวัด${o.addr_province} ${o.addr_postcode}`;
  for (const key of CERT_SCENARIOS) {
    const s = JOURNALS.find((x) => x.key === key);
    const docno = `${books[s.t].code}${s.no}`;
    const id = manifest.journals[docno]?.id ?? byDocno.get(docno)?.id;
    if (!id) { console.log(`50 ทวิ ${key}: journal ${docno} not found`); report.files.push({ label: `50tawi ${key}`, status: 'missing' }); continue; }
    const jr = (await call('GET', `/api/gl/journals/${encodeURIComponent(id)}`)).data ?? {};
    for (const w of (jr.details?.withholdings ?? []).filter((x) => x.wht_direction === 1)) {
      report.files.push(await download(`50 ทวิ ${key} ${docno} ${w.form_type}`, '/api/goapi/api/report/tax/wht/certificate', { ...co, journalid: id, withholdingid: w.id, certificate: { payer: { name: '', taxid: '', address: payerAddress }, archivecopy: true } }, `50tawi-${docno}.pdf`, outDir));
    }
  }
  writeFileSync(path.join(outDir, 'verify-report.json'), `${JSON.stringify(report, null, 2)}\n`);
  console.log(`report → ${path.relative(ROOT, path.join(outDir, 'verify-report.json'))}`);
}

await login();
await (MODE === 'verify' ? verify() : seed());
