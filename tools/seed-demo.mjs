#!/usr/bin/env node
// Put sample records into the demo tenant through the real API, so screens are
// not empty during development. Nothing here is mock data on the frontend: every
// record is created by the same endpoint the screen itself calls, so a screen
// that renders seeded rows is proof the read path works too.
//
//   node tools/seed-demo.mjs            everything this tool knows how to seed
//   node tools/seed-demo.mjs fa         one module only
//   node tools/seed-demo.mjs fa --wipe  delete what a previous run created first
//
// Target defaults to the local backend; override with BC_SEED_API.
// Safe to re-run: records are matched by code and skipped when they exist.
const api = (process.env.BC_SEED_API ?? "http://localhost:8888").replace(/\/+$/, "");
const holdingCode = process.env.BC_SEED_HOLDING ?? "demo";
const businessCode = process.env.BC_SEED_BUSINESS ?? "C01";
const modules = process.argv.slice(2).filter((a) => !a.startsWith("--"));
const wipe = process.argv.includes("--wipe");

let token = "";
const call = async (path, init = {}) => {
  const response = await fetch(`${api}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers ?? {}),
    },
  });
  const text = await response.text();
  let body;
  try {
    body = JSON.parse(text);
  } catch {
    body = { success: false, message: text.slice(0, 200) };
  }
  if (response.status === 404) throw new Error(`${path} is not in this backend build (404) — rebuild mainapi`);
  return body;
};

async function signIn() {
  // The demo account is the one the login screen's demo button uses; the backend
  // decides whether it exists (BCAI_DEMO_LOGIN_ENABLED), so no secret lives here.
  const login = await call("/demo-login", { method: "POST" });
  if (!login?.token) throw new Error(`demo login failed: ${login?.message ?? "no token"}`);
  token = login.token;
  const selected = await call("/select-shop", {
    method: "POST",
    body: JSON.stringify({ holdingcode: holdingCode, businesscode: businessCode }),
  });
  if (selected?.success !== true) throw new Error(`select shop failed: ${selected?.message ?? "unknown"}`);
  console.log(`signed in as demo, working in ${holdingCode}/${businessCode} on ${api}`);
}

const uuid = () => crypto.randomUUID();
const thai = (name) => [{ code: "th", name }];

// ---------------------------------------------------------------- fa

// Account codes follow the chart the fixed-asset screen defaults to.
const FA_TYPES = [
  {
    typecode: "EQUIPMENT",
    name: "เครื่องใช้สำนักงาน",
    defaultusefullifeyears: 5,
    defaultdeprecpercent: "20.00",
    assetaccountcode: "120101",
    accumdeprecaccountcode: "129101",
    deprecexpenseaccountcode: "520103",
  },
  {
    typecode: "VEHICLE",
    name: "ยานพาหนะ",
    defaultusefullifeyears: 5,
    defaultdeprecpercent: "20.00",
    assetaccountcode: "120301",
    accumdeprecaccountcode: "129301",
    deprecexpenseaccountcode: "520303",
  },
  {
    typecode: "COMPUTER",
    name: "คอมพิวเตอร์และอุปกรณ์",
    defaultusefullifeyears: 3,
    defaultdeprecpercent: "33.33",
    assetaccountcode: "120401",
    accumdeprecaccountcode: "129401",
    deprecexpenseaccountcode: "520403",
  },
];

const FA_ASSETS = [
  {
    assetcode: "FA-0001",
    name: "โต๊ะทำงานผู้จัดการ",
    assettypecode: "EQUIPMENT",
    cost: "24000.00",
    usefullifeyears: 5,
    deprecpercent: "20.00",
    purchasedate: "2025-01-15",
  },
  {
    assetcode: "FA-0002",
    name: "รถกระบะส่งของ",
    assettypecode: "VEHICLE",
    cost: "685000.00",
    usefullifeyears: 5,
    deprecpercent: "20.00",
    purchasedate: "2024-07-01",
  },
  {
    assetcode: "FA-0003",
    name: "คอมพิวเตอร์โน้ตบุ๊ก ฝ่ายบัญชี",
    assettypecode: "COMPUTER",
    cost: "38900.00",
    usefullifeyears: 3,
    deprecpercent: "33.33",
    purchasedate: "2025-03-20",
    firstyearpercent: "40.00", // SME first-year allowance for computers
  },
  {
    assetcode: "FA-0004",
    name: "เครื่องปรับอากาศ หน้าร้าน",
    assettypecode: "EQUIPMENT",
    cost: "45000.00",
    usefullifeyears: 5,
    deprecpercent: "20.00",
    purchasedate: "2024-11-05",
  },
  {
    assetcode: "FA-0005",
    name: "ชั้นวางสินค้าเหล็ก คลังใหญ่",
    assettypecode: "EQUIPMENT",
    cost: "128000.00",
    usefullifeyears: 10,
    deprecpercent: "10.00",
    purchasedate: "2023-09-12",
  },
];

async function seedFixedAssets() {
  const existingTypes = await call("/fa/v2/types");
  const typeCodes = new Set((existingTypes?.items ?? []).map((t) => t.typecode));
  let created = 0;
  let skipped = 0;

  for (const type of FA_TYPES) {
    if (typeCodes.has(type.typecode)) {
      skipped += 1;
      continue;
    }
    const result = await call("/fa/v2/command", {
      method: "POST",
      body: JSON.stringify({
        resource: "types",
        action: "create",
        requestid: uuid(),
        assettype: {
          typecode: type.typecode,
          names: thai(type.name),
          defaultusefullifeyears: type.defaultusefullifeyears,
          defaultdeprecpercent: type.defaultdeprecpercent,
          assetaccountcode: type.assetaccountcode,
          accumdeprecaccountcode: type.accumdeprecaccountcode,
          deprecexpenseaccountcode: type.deprecexpenseaccountcode,
          isactive: true,
        },
      }),
    });
    if (result?.success !== true) throw new Error(`type ${type.typecode}: ${result?.message ?? "failed"}`);
    created += 1;
  }
  console.log(`  asset types: created ${created}, already there ${skipped}`);

  const existingAssets = await call("/fa/v2/assets?limit=200");
  const assetCodes = new Set((existingAssets?.items ?? []).map((a) => a.assetcode));
  created = 0;
  skipped = 0;

  for (const asset of FA_ASSETS) {
    if (assetCodes.has(asset.assetcode)) {
      skipped += 1;
      continue;
    }
    const type = FA_TYPES.find((t) => t.typecode === asset.assettypecode);
    const result = await call("/fa/v2/command", {
      method: "POST",
      body: JSON.stringify({
        resource: "assets",
        action: "create",
        requestid: uuid(),
        asset: {
          assetcode: asset.assetcode,
          names: thai(asset.name),
          assettypecode: asset.assettypecode,
          cost: asset.cost,
          scrapvalue: "1.00",
          usefullifeyears: asset.usefullifeyears,
          deprecpercent: asset.deprecpercent,
          purchasedate: asset.purchasedate,
          startcalcdate: asset.purchasedate,
          firstyearpercent: asset.firstyearpercent ?? "0.00",
          assetaccountcode: type.assetaccountcode,
          accumdeprecaccountcode: type.accumdeprecaccountcode,
          deprecexpenseaccountcode: type.deprecexpenseaccountcode,
          status: "active",
        },
      }),
    });
    if (result?.success !== true) throw new Error(`asset ${asset.assetcode}: ${result?.message ?? "failed"}`);
    created += 1;
  }
  console.log(`  assets: created ${created}, already there ${skipped}`);
}

async function wipeFixedAssets() {
  const existing = await call("/fa/v2/assets?limit=200");
  const seeded = new Set(FA_ASSETS.map((a) => a.assetcode));
  let removed = 0;
  // Only codes this tool creates: never a broad match on names (UAT rule 4).
  for (const asset of existing?.items ?? []) {
    if (!seeded.has(asset.assetcode)) continue;
    const result = await call("/fa/v2/command", {
      method: "POST",
      body: JSON.stringify({
        resource: "assets",
        action: "delete",
        requestid: uuid(),
        id: asset.id,
        version: asset.version,
      }),
    });
    if (result?.success !== true) throw new Error(`delete ${asset.assetcode}: ${result?.message ?? "failed"}`);
    removed += 1;
  }
  console.log(`  assets removed: ${removed}`);
}

const MODULES = {
  fa: { title: "สินทรัพย์ถาวรและค่าเสื่อมราคา", seed: seedFixedAssets, wipe: wipeFixedAssets },
};

const wanted = modules.length ? modules : Object.keys(MODULES);
const unknown = wanted.filter((name) => !MODULES[name]);
if (unknown.length) {
  console.error(`unknown module: ${unknown.join(", ")} (have: ${Object.keys(MODULES).join(", ")})`);
  process.exit(1);
}

await signIn();
for (const name of wanted) {
  const module = MODULES[name];
  console.log(`${name} — ${module.title}`);
  if (wipe) await module.wipe();
  else await module.seed();
}
