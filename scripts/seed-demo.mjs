/**
 * Seed ข้อมูลตัวอย่างสำหรับบัญชี Demo (ปุ่ม "ทดลองใช้ระบบ (Demo)") — SME ไทย เจ้าของคนเดียวหลายกิจการ
 * ใช้ได้ทั้ง local และ public server: ทุก call ผ่าน frontend origin (rewrite /backend/goapi → mainapi)
 *
 *   node scripts/seed-demo.mjs                       # local  http://127.0.0.1:3000
 *   SEED_BASE=https://account.bcaicloud.com node scripts/seed-demo.mjs
 *
 * ข้อมูลอยู่ใน scripts/demo-data.json (schema: holding, companies[branches, product_groups, products,
 * customers, suppliers], departments, permission_sets, employees). Idempotent: รายการที่มีอยู่แล้ว
 * (409/duplicate) ถูกข้าม จึงรันซ้ำได้.
 */
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const BASE = (process.env.SEED_BASE || 'http://127.0.0.1:3000').replace(/\/+$/, '');
const API = `${BASE}/backend`; // Next rewrite /backend/:path* → mainapi root
const here = path.dirname(fileURLToPath(import.meta.url));
const data = JSON.parse(readFileSync(path.join(here, 'demo-data.json'), 'utf8'));

let token = '';
const stats = { created: 0, skipped: 0, failed: 0 };

function th(name, en) {
  const names = [{ code: 'th', name }];
  if (en) names.push({ code: 'en', name: en });
  return names;
}

async function call(method, url, body, { quiet = false } = {}) {
  const res = await fetch(url, {
    method,
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json', Origin: BASE },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const text = await res.text();
  let json = null;
  try { json = JSON.parse(text); } catch { json = { raw: text.slice(0, 200) }; }
  const dup = res.status === 409 || /exist|ซ้ำ|มีอยู่แล้ว|duplicate|already/i.test(json?.message ?? '');
  if (res.ok && json?.success !== false) return { ok: true, dup: false, json };
  if (dup) return { ok: false, dup: true, json };
  if (!quiet) console.log(`  ! ${method} ${url.replace(BASE, '')} → ${res.status} ${JSON.stringify(json).slice(0, 160)}`);
  return { ok: false, dup: false, json, status: res.status };
}

/** POST ที่นับสถิติ created/skipped/failed */
async function create(label, url, body) {
  const r = await call('POST', url, body);
  if (r.ok) stats.created++; else if (r.dup) stats.skipped++; else stats.failed++;
  return r;
}

async function selectHolding(businesscode) {
  const r = await call('POST', `${API}/select-holding`, { holdingcode: data.holding.code, businesscode: businesscode ?? '' });
  if (!r.ok) throw new Error(`select-holding ${businesscode ?? ''} failed`);
  if (r.json?.token) token = r.json.token; // token อาจถูกออกใหม่พร้อม holding context
}

async function listAll(url) {
  const r = await call('GET', url, undefined, { quiet: true });
  const d = r.json?.data;
  return Array.isArray(d) ? d : Array.isArray(d?.data) ? d.data : [];
}

async function main() {
  console.log(`seed demo → ${BASE}`);
  // 1) demo login (backend สร้าง user demo ให้อัตโนมัติครั้งแรก)
  const login = await fetch(`${BASE}/api/auth/demo-login`, { method: 'POST', headers: { Origin: BASE } }).then((r) => r.json());
  if (!login?.token) throw new Error(`demo-login failed: ${JSON.stringify(login).slice(0, 200)}`);
  token = login.token;
  console.log('demo login ok');

  // 2) holding (กลุ่มกิจการ)
  const holdings = await listAll(`${API}/list-holding?limit=100&offset=0`);
  if (!holdings.some((h) => (h.holdingcode ?? h.code) === data.holding.code)) {
    const r = await create('holding', `${API}/create-holding`, {
      holdingcode: data.holding.code,
      name1: data.holding.name_th,
      names: th(data.holding.name_th, data.holding.name_en),
      address: [], branchcode: '', images: [], logo: '', profilepicture: '',
      settings: { emailowners: [], emailstaffs: [], isusebranch: true, isusedepartment: true },
    });
    if (!r.ok && !r.dup) throw new Error('create-holding failed');
    console.log(`holding ${data.holding.code} created`);
  } else console.log(`holding ${data.holding.code} exists`);
  await selectHolding();

  // 3) ประเภทธุรกิจ / แผนก / ชุดสิทธิ์การใช้งาน (ระดับ holding)
  const businessTypes = [...new Set(data.companies.map((c) => c.businesstype).filter(Boolean))];
  for (const [i, name] of businessTypes.entries()) {
    await create('businesstype', `${API}/organization/business-type`, { code: `BT${String(i + 1).padStart(2, '0')}`, names: th(name), isdefault: i === 0 });
  }
  for (const d of data.departments ?? []) {
    await create('department', `${API}/organization/department`, { code: d.code, names: th(d.name_th) });
  }
  for (const p of data.permission_sets ?? []) {
    const permissions = [];
    for (const screen of p.screens ?? []) permissions.push(screen, `${screen}:create`, `${screen}:update`);
    await create('permission-set', `${API}/organization/role-permission`, { rolecode: p.code, names: th(p.name_th), permissions, isactive: true });
  }
  console.log(`holding-level masters done (${businessTypes.length} business types, ${data.departments?.length ?? 0} departments, ${data.permission_sets?.length ?? 0} permission sets)`);

  // 4) บริษัท + สาขา
  const existingCompanies = await listAll(`${API}/organization/company?management=true`);
  const companyUid = {}; // code → companyuid
  const branchUid = {}; // `${code}/${branch}` → branchuid
  for (const c of data.companies) {
    let rec = existingCompanies.find((x) => (x.code ?? '').toUpperCase() === c.code.toUpperCase());
    if (!rec) {
      await create('company', `${API}/organization/company`, { code: c.code, names: th(c.name_th, c.name_en), taxid: c.taxid ?? '', logouri: '', isactive: true });
      // create ตอบแค่ success → อ่าน record กลับมาเพื่อเอา companyuid
      rec = (await listAll(`${API}/organization/company?management=true`)).find((x) => (x.code ?? '').toUpperCase() === c.code.toUpperCase());
    }
    const uid = rec?.companyuid ?? rec?.guidfixed;
    if (!uid) { console.log(`  ! company ${c.code} has no uid, skip branches`); continue; }
    companyUid[c.code] = uid;
    const existingBranches = await listAll(`${API}/organization/branch?management=true`);
    for (const b of c.branches ?? []) {
      let br = existingBranches.find((x) => (x.companyuid === uid) && (x.code ?? '') === b.code);
      if (!br) {
        await create('branch', `${API}/organization/branch`, {
          companyuid: uid, code: b.code, names: th(b.name_th), logouri: '',
          timezone: 'Asia/Bangkok', timezonelabel: '(UTC+07:00) Bangkok', timezoneoffset: '+07:00',
          language: 'th', dateformat: 'DD/MM/YYYY', yeartype: 'BE', basecurrency: 'THB',
          branchtype: b.code === '00000' ? 'head' : 'branch', isvatregistered: true, companyregistrationno: c.taxid ?? '',
          email: '', managername: '', fiscalstartmonth: 1, documentformats: [],
          addresses: [{ address: b.address_th ?? c.address_th ?? '', phoneprimary: b.phone ?? c.phone ?? '', countrycode: 'TH' }],
          countrycode: 'TH', provincecode: '', districtcode: '', subdistrictcode: '', zipcode: '', etaxenabled: false, isactive: true,
        });
        br = (await listAll(`${API}/organization/branch?management=true`)).find((x) => x.companyuid === uid && (x.code ?? '') === b.code);
      }
      if (br?.branchuid ?? br?.guidfixed) branchUid[`${c.code}/${b.code}`] = br.branchuid ?? br.guidfixed;
    }
    console.log(`company ${c.code} ${c.name_th}: ${(c.branches ?? []).length} branch(es)`);
  }

  // 4b) เจ้าของ demo ต้องมี Scope ครอบทุกบริษัท ไม่งั้น select-holding รายบริษัทถูกปฏิเสธ (Scope = allow-list)
  const members = await listAll(`${API}/holding/users?limit=10&offset=0&page=1&q=`);
  const owner = members.find((m) => m.username === 'demo' || m.role === 2);
  if (owner) {
    const accessscopes = Object.entries(companyUid).map(([code, uid]) => ({ scopetype: 'company', companyuid: uid, businesscode: code, allbranches: true }));
    // keyed by useruid (username ว่าง) ให้ตรง membership ที่ create-holding สร้าง — ไม่งั้น upsert เป็น record ที่สอง
    const r = await call('PUT', `${API}/holding/permission`, { editusername: owner.useruid, username: '', useruid: owner.useruid, role: 2, accessscopes, isaccessdisabled: false, position: 'เจ้าของกิจการ', department: '' });
    console.log(`owner scopes → ${accessscopes.length} companies ${r.ok ? 'ok' : 'FAILED'}`);
    await selectHolding();
  }

  // 5) พนักงาน (holding-level, ผูกบริษัท/สาขาผ่าน accessscopes)
  for (const e of data.employees ?? []) {
    const scope = { scopetype: 'branch', companyuid: companyUid[e.company] ?? '', businesscode: e.company, branchcode: e.branch, branchuid: branchUid[`${e.company}/${e.branch}`] ?? '' };
    await create('employee', `${API}/holding/employee`, {
      code: e.code, name: e.name_th, email: e.email ?? '', pincode: '', isenabled: true, isusepos: false,
      roles: [], accessscopes: [scope], contact: { address: '', countrycode: 'TH' },
    });
  }
  console.log(`employees: ${(data.employees ?? []).length}`);

  // 6) ข้อมูลระดับกิจการ: กลุ่มสินค้า สินค้า/บาร์โค้ด ลูกค้า ผู้ขาย
  for (const c of data.companies) {
    await selectHolding(c.code);
    for (const g of c.product_groups ?? []) {
      await create('product-group', `${API}/product/group`, { code: g.code, names: th(g.name_th), parentguid: '', xsorts: [], isdisabled: false });
    }
    for (const p of c.products ?? []) {
      await create('barcode', `${API}/product/barcode`, {
        barcode: p.barcode, itemcode: p.code, names: th(p.name_th, p.name_en),
        itemunitguid: '', itemunitcode: (p.unit ?? 'ชิ้น').toUpperCase().slice(0, 6), itemunitnames: th(p.unit ?? 'ชิ้น'),
        groupcode: p.group ?? '', imageuri: '', images: [], videos: [],
        prices: [{ keynumber: 1, fromqty: 1, toqty: 999999, price: Number(p.price ?? 0) }],
        costprice: p.cost ?? 0, price: p.price ?? 0, description: '',
      });
    }
    for (const d of c.customers ?? []) {
      await create('debtor', `${API}/debtaccount/debtor`, {
        code: d.code, names: th(d.name_th), personaltype: d.taxid ? 2 : 1, taxid: d.taxid ?? '', branchnumber: d.taxid ? '00000' : '',
        creditday: d.credit_days ?? 0, email: '', customertype: 0, ismember: false, pointscode: '',
        addressforbilling: { address: [d.address_th ?? ''], countrycode: 'TH', phoneprimary: d.phone ?? '' },
      });
    }
    for (const s of c.suppliers ?? []) {
      await create('creditor', `${API}/debtaccount/creditor`, {
        code: s.code, names: th(s.name_th), personaltype: s.taxid ? 2 : 1, taxid: s.taxid ?? '', branchnumber: s.taxid ? '00000' : '',
        creditday: s.credit_days ?? 30, email: '', whtenabled: false, whtrate: 0,
        addressforbilling: { address: [s.address_th ?? ''], countrycode: 'TH', phoneprimary: s.phone ?? '' },
      });
    }
    console.log(`company ${c.code}: ${(c.product_groups ?? []).length} groups, ${(c.products ?? []).length} products, ${(c.customers ?? []).length} customers, ${(c.suppliers ?? []).length} suppliers`);
  }
  console.log(`DONE created=${stats.created} skipped(existing)=${stats.skipped} failed=${stats.failed}`);
  if (stats.failed) process.exitCode = 2;
}

main().catch((e) => { console.error('FATAL', e); process.exit(1); });
