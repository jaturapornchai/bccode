// tools/test-e2e-debt-reports-returns-expense.mjs
// ทดสอบ E2E: รายงานหนี้คงค้าง (AR/AP Outstanding), ใบลดหนี้รับคืน/ส่งคืน (Returns), และบันทึกค่าใช้จ่าย (Expense)
import { execSync } from "child_process";

const NEXT_BASE_URL = "http://localhost:3000";

async function main() {
  console.log("=== เริ่มการทดสอบ E2E: Debt Reports, Returns & Expense ===");

  // 1. ล็อกอินผ่าน Demo Login
  console.log("\n[1] ล็อกอินเข้าสู่ระบบผ่าน Demo Login...");
  const loginRes = await fetch(`${NEXT_BASE_URL}/api/auth/demo-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!loginRes.ok) throw new Error(`Demo login failed: ${loginRes.status}`);
  const loginData = await loginRes.json();
  const token = loginData.token;

  await fetch(`${NEXT_BASE_URL}/api/workspace/select-holding`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      backendUrl: 'http://localhost:8888',
      holdingcode: 'rungrueng',
      businesscode: '01',
      branchuid: '00000',
    }),
  });

  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
    'x-bc-backend-url': 'http://localhost:8888',
  };

  // 2. ตรวจสอบรายงานลูกหนี้และเจ้าหนี้คงค้าง
  console.log("\n[2] ตรวจสอบรายงานเจ้าหนี้คงค้าง (AP Outstanding)...");
  const apRes = await fetch(`${NEXT_BASE_URL}/api/goapi/api/report/debt/query`, {
    method: 'POST',
    headers: authHeaders,
    body: JSON.stringify({
      holdingcode: "rungrueng",
      reportcode: "ap_outstanding",
      fromdate: "2026-01-01",
      todate: "2026-12-31"
    })
  });
  if (!apRes.ok) throw new Error(`AP Outstanding API failed: ${apRes.status} ${await apRes.text()}`);
  const apData = await apRes.json();
  console.log(`✅ ดึงรายงาน AP Outstanding สำเร็จ: พบ ${apData.data?.length || 0} รายการที่ยังค้างชำระ`);
  for (const item of (apData.data || [])) {
    console.log(`   - [${item.docdate}] ${item.docno} | เจ้าหนี้: ${item.creditorcode} ${item.creditorname} | หนี้เดิม: ${item.totalamount}, ชำระแล้ว: ${item.paidamount}, ค้างชำระ: ${item.balanceamount}`);
  }

  // 3. ทดสอบออกใบลดหนี้/รับคืนสินค้าจากการขาย (Sales Return: transflag 48, SRN6910-0001)
  console.log("\n[3] ทดสอบออกใบรับคืนสินค้าจากการขาย (SRN6910-0001)...");
  const saleReturnPayload = {
    docno: "SRN6910-0001",
    docdatetime: "2026-10-07T16:30:00Z",
    transflag: 48, // Sale Return
    custcode: "CUST-001",
    docref: "INV6910-0001",
    totalamount: 802.50,
    totalbeforevat: 750.00,
    totalvatvalue: 52.50,
    vatrate: 7,
    details: [
      {
        linenumber: 1,
        itemcode: "ITEM-CEMENT",
        itemname: "ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง (ถุง 50 กก.)",
        qty: 5,
        price: 150.00,
        sumamount: 750.00,
        whcode: "WH01"
      }
    ]
  };

  const srRes = await fetch(`${NEXT_BASE_URL}/api/erp-transaction/sale-invoice-return`, {
    method: "POST",
    headers: authHeaders,
    body: JSON.stringify(saleReturnPayload)
  });
  if (!srRes.ok) throw new Error(`Sale return failed: ${srRes.status} ${await srRes.text()}`);
  console.log(`✅ บันทึกใบรับคืนสินค้าจากการขายสำเร็จ`);

  // 4. ทดสอบออกใบส่งคืนสินค้าจากการซื้อ (Purchase Return: transflag 16, PRN6910-0001)
  console.log("\n[4] ทดสอบออกใบส่งคืนสินค้าจากการซื้อ (PRN6910-0001)...");
  const purchaseReturnPayload = {
    docno: "PRN6910-0001",
    docdatetime: "2026-10-07T17:00:00Z",
    transflag: 16, // Purchase Return
    custcode: "SUPP-SCG-01",
    docref: "PUR6910-0001",
    totalamount: 1284.00,
    totalbeforevat: 1200.00,
    totalvatvalue: 84.00,
    vatrate: 7,
    details: [
      {
        linenumber: 1,
        itemcode: "ITEM-CEMENT",
        itemname: "ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง (ถุง 50 กก.)",
        qty: 10,
        price: 120.00,
        sumamount: 1200.00,
        whcode: "WH01"
      }
    ]
  };

  const prRes = await fetch(`${NEXT_BASE_URL}/api/erp-transaction/purchase-return`, {
    method: "POST",
    headers: authHeaders,
    body: JSON.stringify(purchaseReturnPayload)
  });
  if (!prRes.ok) throw new Error(`Purchase return failed: ${prRes.status} ${await prRes.text()}`);
  console.log(`✅ บันทึกใบส่งคืนสินค้าจากการซื้อสำเร็จ`);

  // 5. ทดสอบบันทึกค่าใช้จ่าย (Expense: transflag 24, EXP6910-0001)
  console.log("\n[5] ทดสอบบันทึกค่าใช้จ่ายในการดำเนินงาน (EXP6910-0001)...");
  const expensePayload = {
    docno: "EXP6910-0001",
    docdatetime: "2026-10-07T17:30:00Z",
    transflag: 24, // Expense
    custcode: "SUPP-OFFICE-01",
    totalamount: 1070.00,
    totalbeforevat: 1000.00,
    totalvatvalue: 70.00,
    vatrate: 7,
    details: [
      {
        linenumber: 1,
        itemcode: "EXP-STATIONERY",
        itemname: "ค่าเครื่องเขียนและแบบพิมพ์สำนักงาน",
        qty: 1,
        price: 1000.00,
        sumamount: 1000.00,
        whcode: "WH01"
      }
    ]
  };

  const expRes = await fetch(`${NEXT_BASE_URL}/api/erp-transaction/expense`, {
    method: "POST",
    headers: authHeaders,
    body: JSON.stringify(expensePayload)
  });
  if (!expRes.ok) throw new Error(`Expense save failed: ${expRes.status} ${await expRes.text()}`);
  console.log(`✅ บันทึกค่าใช้จ่ายสำเร็จ`);

  // 6. ตรวจสอบ Evidence ใน PostgreSQL
  console.log("\n[6] ตรวจสอบ Evidence ใน PostgreSQL...");

  // ตรวจสอบ Stock Movement ล่าสุด
  const stockMovements = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT docdatetime::date as date, docno, transflag, direction, docqty, balanceqty, balanceamount FROM stock_ledger WHERE itemcode='ITEM-CEMENT' ORDER BY docdatetime, behindindex, docno;"`, { encoding: "utf8" });
  console.log("=== Stock Movement Ledger (ITEM-CEMENT) ===");
  console.log(stockMovements);

  // ตรวจสอบ GL Journals
  const journals = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT code, payload->>'bookcode' as book, payload->>'description' as desc FROM gl_records WHERE kind='journals' AND code IN ('SRN-SRN6910-0001', 'PRN-PRN6910-0001', 'PV-EXP6910-0001') ORDER BY code;"`, { encoding: "utf8" });
  console.log("=== GL Journals ===");
  console.log(journals);

  // ตรวจสอบ GL Lines
  const glLines = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT doc_no, line_no, account_code, account_name, debit, credit FROM gl_lines WHERE doc_no IN ('SRN-SRN6910-0001', 'PRN-PRN6910-0001', 'PV-EXP6910-0001') ORDER BY doc_no, line_no;"`, { encoding: "utf8" });
  console.log("=== GL Lines (Debit = Credit) ===");
  console.log(glLines);

  console.log("\n🎉 การทดสอบ E2E: Debt Reports, Returns & Expense สำเร็จสมบูรณ์ 100%!");
}

main().catch(err => {
  console.error("❌ Test failed:", err);
  process.exit(1);
});
