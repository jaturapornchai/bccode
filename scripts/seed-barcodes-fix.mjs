/**
 * รอบสอง: ผูกรูปเข้า barcode 20 รายการที่สร้างแล้ว (UATBC01-20)
 * - รูป 18 รูปอัปโหลดแล้วในรอบแรก (images collection เก็บ r2key) → ดึงกลับมาใช้
 * - รายการ 16/19 รูปเดิมเป็น webp โดน reject → อัปโหลดรูปสำรอง jpg ใหม่ + อัปเดตชื่อให้ตรงรูป
 * - PUT /api/product-barcode/:guid พร้อม GET full doc ก่อน (แบบเดียวกับ product screen)
 * รัน: node scripts/seed-barcodes-fix.mjs
 */
import { execSync } from 'node:child_process';

const BASE = 'http://127.0.0.1:3000';

const REPLACE = {
  6: { name: 'โคคา-โคล่าแซโร่ชูการ์กระป๋อง', img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/822e47e3012b.jpg', ext: 'jpg' },
  16: { name: 'น้ำยาล้างจานสูตรซิตรอนเนล 1000 มล.', img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/ba4f916cce29.jpg', ext: 'jpg' },
  17: { name: 'แปรงฟันสี่ด้านสีสันสดใส', img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/5ee0b035225a.jpg', ext: 'jpg' },
  19: { name: 'มะม่วงอบกรอบถุง', img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/f056fe52a6d8.jpg', ext: 'jpg' },
  20: { name: 'คุกกี้มะพร้าวกล่อง', img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/a3bfa7b4f1d5.jpg', ext: 'jpg' },
};

function mongoJson(js) {
  const out = execSync(`docker exec mongodb mongosh --quiet appdb --eval "JSON.stringify(${js})"`, { timeout: 45000 }).toString().trim();
  const start = out.indexOf('[') >= 0 ? out.indexOf('[') : out.indexOf('{');
  return JSON.parse(out.slice(start));
}

async function main() {
  // 1) login + select holding
  const login = await (await fetch(`${BASE}/api/auth/dev-login`, { method: 'POST', headers: { Origin: BASE } })).json();
  const auth = { Authorization: `Bearer ${login.token}`, 'x-bc-backend-url': 'http://mainapi:8888' };
  await fetch(`${BASE}/api/workspace/select-holding`, {
    method: 'POST',
    headers: { ...auth, 'Content-Type': 'application/json' },
    body: JSON.stringify({ backendUrl: 'http://mainapi:8888', holdingcode: 'bc001', businesscode: 'TST03' }),
  });
  console.log('session ok');

  // 2) barcode 20 รายการจาก Mongo → {itemcode, guid}
  const bcs = mongoJson(`db.productbarcodes.find({holdingcode:'bc001', itemcode:/^UATBC[0-9]{2}$/}, {itemcode:1, guidfixed:1, barcode:1}).toArray().map(d=>({itemcode:d.itemcode, guid:d.guidfixed}))`);
  const byIdx = new Map(bcs.map((b) => [Number(b.itemcode.replace('UATBC', '')), b]));
  console.log('barcodes in Mongo:', bcs.length);

  // 3) รูปรอบแรก: images collection (originalname uatbc-XX) → r2key
  const imgs = mongoJson(`db.images.find({holdingcode:'bc001', originalname:/^uatbc-[0-9]{2}\\./}, {originalname:1, r2key:1}).toArray()`);
  const imgByIdx = new Map(imgs.map((i) => [Number(i.originalname.replace('uatbc-', '').split('.')[0]), `/goapi/s3/file/${i.r2key}`]));
  console.log('reusable images:', imgs.length);

  // 4) อัปโหลดรูปสำรองสำหรับรายการที่ไม่มีรูป
  for (const [idxStr, rep] of Object.entries(REPLACE)) {
    const idx = Number(idxStr);
    const buf = Buffer.from(await (await fetch(rep.img)).arrayBuffer());
    const form = new FormData();
    form.append('file', new Blob([buf]), `uatbc-${String(idx).padStart(2, '0')}.${rep.ext}`);
    form.append('category', 'products');
    const up = await (await fetch(`${BASE}/api/product-barcode/image`, { method: 'POST', headers: auth, body: form })).json();
    const uri = up?.url ?? (up?.data?.r2key ? `/goapi/s3/file/${up.data.r2key}` : '');
    if (uri) imgByIdx.set(idx, uri);
    console.log(`  [${idx}] reupload ${rep.name} → ${uri ? 'ok' : 'FAIL ' + JSON.stringify(up).slice(0, 120)}`);
  }

  // 5) PUT ทุกรายการ: GET full doc → set imageuri/images (+names สำหรับ 16/19) → PUT
  let ok = 0;
  for (let idx = 1; idx <= 20; idx++) {
    const bc = byIdx.get(idx);
    const uri = imgByIdx.get(idx);
    if (!bc || !uri) { console.log(`  [${idx}] SKIP (bc=${!!bc}, uri=${!!uri})`); continue; }
    const getRes = await fetch(`${BASE}/api/product-barcode/${encodeURIComponent(bc.guid)}`, { headers: auth });
    const getJson = await getRes.json().catch(() => null);
    const doc = getJson?.data;
    if (!getRes.ok || !doc) { console.log(`  [${idx}] GET failed ${getRes.status}`); continue; }
    doc.imageuri = uri;
    doc.images = [{ xorder: 1, uri }];
    if (REPLACE[idx]) doc.names = [{ code: 'th', name: REPLACE[idx].name }, { code: 'en', name: `UAT Item ${idx}` }];
    const putRes = await fetch(`${BASE}/api/product-barcode/${encodeURIComponent(bc.guid)}`, {
      method: 'PUT',
      headers: { ...auth, 'Content-Type': 'application/json' },
      body: JSON.stringify({ backendUrl: 'http://mainapi:8888', data: doc }),
    });
    const putJson = await putRes.json().catch(() => ({}));
    if (putRes.ok && putJson.success !== false) ok++;
    else console.log(`  [${idx}] PUT failed ${putRes.status}: ${JSON.stringify(putJson).slice(0, 140)}`);
  }
  console.log(`updated ${ok}/20`);

  // 6) ตรวจ Mongo
  const check = mongoJson(`db.productbarcodes.find({holdingcode:'bc001', itemcode:/^UATBC[0-9]{2}$/}, {itemcode:1, imageuri:1, 'images':1}).toArray().map(d=>({i:d.itemcode, uri:(d.imageuri||'').slice(0,30), n:(d.images||[]).length}))`);
  const withImg = check.filter((c) => c.uri.startsWith('/goapi/s3/file/')).length;
  console.log(`verify: ${check.length} docs, ${withImg} มีรูป /goapi/s3/file/`);
  const noImg = check.filter((c) => !c.uri.startsWith('/goapi/s3/file/')).map((c) => c.i);
  if (noImg.length) console.log('MISSING:', noImg.join(','));
}

main().catch((e) => { console.error('FATAL', e); process.exit(1); });
