#!/usr/bin/env node
// Ask the local backend which of the endpoints the screens call actually exist.
//
//   node tools/probe-endpoints.mjs            every group
//   node tools/probe-endpoints.mjs transaction fa
//
// Why this is worth a tool: a screen whose endpoint is missing does not look
// broken. It renders "ไม่พบข้อมูล" exactly like an empty table, so a whole module
// can be a dead end without anyone noticing. Signing in first matters too —
// without a token every route answers 401 and the answer is meaningless.
//
//   200/2xx  the route is in this build and accepted the request
//   404      the route is not in this build (screen is a dead end today)
//   401/403  the route exists but rejected the session
//   4xx      the route exists and wants different parameters
//
// Target defaults to the local backend; override with BC_PROBE_API.
const api = (process.env.BC_PROBE_API ?? "http://localhost:8888").replace(/\/+$/, "");
const holdingCode = process.env.BC_PROBE_HOLDING ?? "demo";
const businessCode = process.env.BC_PROBE_BUSINESS ?? "C01";
const wanted = process.argv.slice(2).filter((arg) => !arg.startsWith("--"));

let token = "";

async function signIn() {
  const login = await fetch(`${api}/demo-login`, { method: "POST" }).then((r) => r.json());
  if (!login?.token) throw new Error(`demo login failed: ${login?.message ?? "no token"}`);
  token = login.token;
  const selected = await fetch(`${api}/select-shop`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify({ holdingcode: holdingCode, businesscode: businessCode }),
  }).then((r) => r.json());
  if (selected?.success !== true) throw new Error(`select shop failed: ${selected?.message ?? "unknown"}`);
}

async function probe(method, path, body) {
  try {
    const response = await fetch(`${api}${path}`, {
      method,
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    return response.status;
  } catch (error) {
    return `ERR ${error instanceof Error ? error.message : String(error)}`;
  }
}

// Modules the erp-transaction proxy allows, i.e. every document screen in the
// work tabs (frontend/src/app/api/erp-transaction/[...erpPath]/route.ts).
const TRANSACTION_MODULES = [
  "sale-invoice", "sale-order", "quotation", "sale-invoice-return",
  "purchase", "purchase-order", "purchase-requisition", "purchase-return",
  "paid", "pay", "stock-balance", "stock-receive-product", "stock-prickup-product",
  "stock-return-product", "stock-transfer", "stock-adjustment", "receivableother",
  "billingnote", "deposit", "depositrefund", "paidadvance", "paidadvancerefund",
  "receivedeposit", "receivedepositrefund",
  "bank/saledebitnote", "bank/purchasedebitnote", "bank/banktransferrecord",
  "bank/depositrecord", "chequereceive/chequedeposit", "chequepayment/chequepaymentdeposit",
];

const GROUPS = {
  transaction: () =>
    TRANSACTION_MODULES.map((module) => ({
      label: module,
      method: "GET",
      path: `/transaction/${module}/list?limit=1`,
    })),
  fa: () => [
    { label: "asset list", method: "GET", path: "/fa/v2/assets?limit=1" },
    { label: "asset types", method: "GET", path: "/fa/v2/types" },
    { label: "depreciation schedule", method: "GET", path: "/fa/v2/reports/schedule?year=2026" },
    { label: "tax reconciliation", method: "GET", path: "/fa/v2/reports/tax-reconciliation?year=2026" },
  ],
  approval: () => [
    // erp-operations.ts APPROVAL_KINDS: only "pr" is wired on the frontend today.
    { label: "pr pending", method: "POST", path: "/goapi/api/approval/pr-status/pending", body: { holdingcode: holdingCode } },
    { label: "po pending", method: "POST", path: "/goapi/api/approval/po-status/pending", body: { holdingcode: holdingCode } },
  ],
  tools: () => [
    // erp-tools.ts TOOL_ENDPOINTS — the rest of the 12 tool screens have no endpoint.
    { label: "product balance", method: "POST", path: "/goapi/api/process/product-balance", body: { holdingcode: holdingCode, businesscode: businessCode } },
    { label: "audit data", method: "POST", path: "/goapi/api/stockcost/check", body: { holdingcode: holdingCode, businesscode: businessCode, fromdate: "2026-01-01", todate: "2026-12-31" } },
    { label: "queue status", method: "POST", path: "/goapi/api/process/queue-status", body: { holdingcode: holdingCode } },
  ],
  report: () => [
    { label: "sales by document", method: "POST", path: "/goapi/api/report/sales/by-document", body: { holdingcode: holdingCode, businesscode: businessCode, fromdate: "2026-01-01", todate: "2026-12-31" } },
  ],
};

const label = (status) => {
  if (typeof status !== "number") return status;
  if (status === 404) return `${status}  ไม่มี route นี้ใน build`;
  if (status === 401 || status === 403) return `${status}  route มี แต่ปฏิเสธ session`;
  if (status >= 200 && status < 300) return `${status}  ใช้งานได้`;
  return `${status}`;
};

async function main() {
  await signIn();
  console.log(`probing ${api} as demo (${holdingCode}/${businessCode})\n`);

  const groups = wanted.length ? wanted : Object.keys(GROUPS);
  const missing = [];
  for (const group of groups) {
    const build = GROUPS[group];
    if (!build) {
      console.log(`${group}: ไม่รู้จักกลุ่มนี้ (${Object.keys(GROUPS).join(", ")})`);
      continue;
    }
    console.log(`== ${group} ==`);
    for (const check of build()) {
      const status = await probe(check.method, check.path, check.body);
      if (status === 404) missing.push(`${group}/${check.label}`);
      console.log(`  ${check.label.padEnd(26)} ${label(status)}`);
    }
    console.log("");
  }

  if (missing.length) {
    console.log(`ไม่มีใน backend build นี้ ${missing.length} รายการ: ${missing.join(", ")}`);
    console.log("จอที่เรียก route เหล่านี้จะขึ้นว่า \"ไม่พบข้อมูล\" ทั้งที่จริงคือยังไม่มี API");
  } else {
    console.log("ทุก endpoint ที่ตรวจมีอยู่จริงใน build นี้");
  }
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : error);
  process.exit(1);
});
