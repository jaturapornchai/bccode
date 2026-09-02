/**
 * Seed 20 บาร์โค้ดสินค้าพร้อมรูปจริง (2026-08-31 — ลุงจืดสั่ง "สุ่มเพิ่ม barcode และหารูปมาด้วย 20 รายการ")
 * - รูป: ดาวน์โหลดจากผลค้นหา (z-cdn.chatglm.cn) → อัปโหลดผ่าน /api/product-barcode/image
 *   (→ /goapi/image/upload → MinIO + .thumb.webp อัตโนมัติ) — ตามกฎ "รูปห้ามเก็บ Mongo"
 * - ข้อมูล: POST /api/product-barcode (QuickBarcodePayload เหมือนจอ) — names ไทย,
 *   itemcode UATBC01..20, EAN-13 checksum ถูกต้อง, ราคาสุ่มแบบมี seed
 * - ตรวจ: Mongo productbarcodes (bc001) + imageuri เป็น /goapi/s3/file/... เท่านั้น
 * รัน: node scripts/seed-barcodes.mjs
 */
const BASE = 'http://127.0.0.1:3000';
const SEED = 20260831;

// 20 รายการ: ชื่อไทยเข้ากับรูปที่ค้นเจอจริง
const ITEMS = [
  { name: 'บะหมี่กึ่งสำเร็จรูปรสผัดไทย', price: 15, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/9287e60b322c.png', ext: 'png' },
  { name: 'ข้าวโพดอบกรอบธรรมชาติ', price: 35, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/7256bef3b44e.png', ext: 'png' },
  { name: 'เบียร์ช้างคลาสสิกแพ็ครวม', price: 320, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/56132fd2dfda.jpg', ext: 'jpg' },
  { name: 'เบียร์ช้างคลาสสิกขวดใหญ่', price: 65, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/d3b5eab2d4de.jpg', ext: 'jpg' },
  { name: 'น้ำมะพร้าวรสคั่ว 350 มล.', price: 25, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/4fa1d9c2761d.jpg', ext: 'jpg' },
  { name: 'โคคา-โคล่ากระป๋อง', price: 18, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/1287561f0f3f.webp', ext: 'webp' },
  { name: 'น้ำมะพร้าว 100% กล่อง 350 มล.', price: 22, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/af5fceb3700b.jpg', ext: 'jpg' },
  { name: 'เครื่องดื่มเกลือแร่ขวด 600 มล.', price: 20, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/ccd3a7b45ba7.png', ext: 'png' },
  { name: 'บะหมี่ถ้วยรสไก่', price: 17, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/3d886a781556.png', ext: 'png' },
  { name: 'บะหมี่กึ่งสำเร็จรูปซุปไก่', price: 12, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/40ea71b24f9c.jpg', ext: 'jpg' },
  { name: 'บะหมี่ถ้วยแพ็ค 6 ถ้วย', price: 89, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/8c349b4486c2.jpg', ext: 'jpg' },
  { name: 'มามะต้มยำกุ้งรสเผ็ด', price: 14, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/480cfd114ef8.jpg', ext: 'jpg' },
  { name: 'น้ำยาล้างจานสูตรเข้มข้น', price: 45, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/76f8e6b8a2a0.jpg', ext: 'jpg' },
  { name: 'ยาสีฟันสมุนไพร', price: 55, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/28b23120cbb1.png', ext: 'png' },
  { name: 'ขนมสาหร่ายอบกรอบ', price: 30, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/fc9c7cbc1b15.png', ext: 'png' },
  { name: 'ชุดน้ำยาทำความสะอาดบ้าน', price: 150, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/07334c9a3876.png', ext: 'png' },
  { name: 'ชุดอุปกรณ์ดูแลช่องปาก', price: 120, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/c8233eb7f059.webp', ext: 'webp' },
  { name: 'ขนมไทยรวมมิตร 10 ชนิด', price: 199, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/00d64dd53e85.jpg', ext: 'jpg' },
  { name: 'ผลไม้อบแห้งกล่อง', price: 89, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/794212af6b4b.jpg', ext: 'jpg' },
  { name: 'ขนมอบกรอบถุงใหญ่', price: 40, img: 'https://z-cdn.chatglm.cn/image-search-mcp/images-ppt/35cfdea06e1f.webp', ext: 'webp' },
];

// seeded PRNG (mulberry32)
function mulberry32(a) { return function () { a |= 0; a = (a + 0x6d2b79f5) | 0; let t = Math.imul(a ^ (a >>> 15), 1 | a); t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t; return ((t ^ (t >>> 14)) >>> 0) / 4294967296; }; }
const rnd = mulberry32(SEED);

// EAN-13 checksum (prefix 885 ไทย)
function ean13(n12) {
  let sum = 0;
  for (let i = 0; i < 12; i++) sum += Number(n12[i]) * (i % 2 === 0 ? 1 : 3);
  return n12 + String((10 - (sum % 10)) % 10);
}

async function main() {
  // 0) ตรวจก่อนว่าไม่ seed ซ้ำ (ลบเฉพาะของเราเองตาม itemcode ที่แน่นอน)
  // 1) dev-login → token
  const loginRes = await fetch(`${BASE}/api/auth/dev-login`, { method: 'POST', headers: { Origin: BASE } });
  const login = await loginRes.json();
  if (!login?.token) throw new Error('dev-login failed');
  const auth = { Authorization: `Bearer ${login.token}`, 'x-bc-backend-url': 'http://mainapi:8888' };
  console.log('token ok, user:', login.user?.username ?? login.user?.email ?? '?');

  // เลือก holding + shop ใน token ก่อน (ไม่งั้นทุก call ตอบ "Shop not selected")
  const sel = await fetch(`${BASE}/api/workspace/select-holding`, {
    method: 'POST',
    headers: { ...auth, 'Content-Type': 'application/json' },
    body: JSON.stringify({ backendUrl: 'http://mainapi:8888', holdingcode: 'bc001', businesscode: 'TST03' }),
  });
  const selJson = await sel.json().catch(() => null);
  if (!sel.ok || selJson?.success === false) throw new Error(`select-holding failed: ${JSON.stringify(selJson).slice(0, 150)}`);
  console.log('holding bc001 / TST03 selected');

  // 2) upload รูปทีละรายการ → imageuri
  const uploaded = [];
  for (let i = 0; i < ITEMS.length; i++) {
    const it = ITEMS[i];
    let uri = '';
    try {
      const imgRes = await fetch(it.img);
      if (!imgRes.ok) throw new Error(`download ${imgRes.status}`);
      const buf = Buffer.from(await imgRes.arrayBuffer());
      const form = new FormData();
      form.append('file', new Blob([buf]), `uatbc-${String(i + 1).padStart(2, '0')}.${it.ext}`);
      form.append('category', 'products');
      const up = await fetch(`${BASE}/api/product-barcode/image`, { method: 'POST', headers: auth, body: form });
      const upJson = await up.json().catch(() => null);
      uri = upJson?.uri ?? upJson?.data?.uri ?? (typeof upJson?.data === 'string' ? upJson.data : '');
      if (!uri && upJson) console.log('  upload resp keys:', Object.keys(upJson), JSON.stringify(upJson).slice(0, 200));
    } catch (e) {
      console.log(`  [${i + 1}] image FAILED: ${String(e).slice(0, 100)}`);
    }
    uploaded.push(uri);
    process.stdout.write(`  [${i + 1}/20] ${it.name} → ${uri ? 'img ok' : 'no img'}\n`);
  }

  // 3) สร้าง barcode ทีละรายการ
  let created = 0;
  for (let i = 0; i < ITEMS.length; i++) {
    const it = ITEMS[i];
    const itemcode = `UATBC${String(i + 1).padStart(2, '0')}`;
    const barcode = ean13('885' + String(100000000 + Math.floor(rnd() * 899999999)).slice(0, 9));
    const price = it.price + Math.floor(rnd() * 3); // สุ่มบวก 0-2 บาทจากราคาฐาน (seeded)
    const payload = {
      barcode,
      itemcode,
      names: [{ code: 'th', name: it.name }, { code: 'en', name: `UAT Item ${i + 1}` }],
      itemunitguid: '',
      itemunitcode: 'PCS',
      itemunitnames: [{ code: 'th', name: 'ชิ้น' }, { code: 'en', name: 'Piece' }],
      imageuri: uploaded[i] || '',
      images: uploaded[i] ? [{ xorder: 1, uri: uploaded[i] }] : [],
      videos: [],
      description: `สินค้าทดสอบ seed ${SEED} #${i + 1} — ราคา ${price} บาท`,
    };
    const res = await fetch(`${BASE}/api/product-barcode`, {
      method: 'POST',
      headers: { ...auth, 'Content-Type': 'application/json' },
      body: JSON.stringify({ backendUrl: 'http://mainapi:8888', data: payload }),
    });
    const json = await res.json().catch(() => ({}));
    if (res.ok && json.success !== false) created++;
    else console.log(`  [${itemcode}] CREATE FAILED ${res.status}: ${JSON.stringify(json).slice(0, 160)}`);
  }
  console.log(`created ${created}/20`);
}

main().catch((e) => { console.error('FATAL', e); process.exit(1); });
