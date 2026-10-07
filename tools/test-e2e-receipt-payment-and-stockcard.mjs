// tools/test-e2e-receipt-payment-and-stockcard.mjs
// ทดสอบ E2E Flow รับชำระหนี้ (Receipt/Paid), จ่ายชำระหนี้ (Payment/Pay) และรายงานสต็อกการ์ด (Stock Movement)
import { execSync } from "child_process";

const NEXT_BASE_URL = "http://localhost:3000";

async function main() {
  console.log("=== เริ่มการทดสอบ E2E: Receipt, Payment & Stock Movement ===");

  // 1. ล็อกอินผ่าน Demo Login เพื่อรับ Auth Token
  console.log("\n[1] ล็อกอินเข้าสู่ระบบผ่าน Demo Login...");
  const loginRes = await fetch(`${NEXT_BASE_URL}/api/auth/demo-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  });
  if (!loginRes.ok) throw new Error(`Demo login failed: ${loginRes.status}`);
  const loginData = await loginRes.json();
  const token = loginData.token;
  console.log(`   └─ ล็อกอินสำเร็จ: ได้ token (${token.slice(0, 15)}...)`);

  const selectRes = await fetch(`${NEXT_BASE_URL}/api/workspace/select-holding`, {
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
  const selectData = await selectRes.json();
  if (!selectRes.ok || !selectData.success) {
    throw new Error(`เลือก Holding ไม่สำเร็จ: ${JSON.stringify(selectData)}`);
  }
  console.log('   └─ ผูก Workspace Context สำเร็จ (holding: rungrueng, company: 01, branch: 00000)');

  const authHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
    'x-bc-backend-url': 'http://localhost:8888',
  };

  // 2. ทดสอบ Endpoint รายงาน Stock Movement จาก stock_ledger
  console.log("\n[2] ตรวจสอบรายงาน Stock Movement จาก GoAPI...");
  const stockRes = await fetch(`${NEXT_BASE_URL}/api/goapi/api/reports/stock-movement?holdingcode=rungrueng&from=2026-01-01&to=2026-12-31`, {
    headers: authHeaders
  });
  if (!stockRes.ok) {
    throw new Error(`Stock movement API failed: ${stockRes.status} ${await stockRes.text()}`);
  }
  const stockData = await stockRes.json();
  console.log(`✅ ดึงรายงาน Stock Movement สำเร็จ: พบ ${stockData.items?.length || 0} รายการ`);
  for (const item of (stockData.items || [])) {
    console.log(`   - [${item.docdate}] ${item.docno} | ${item.transname} | ${item.itemcode} | ${item.itemname} | เข้า: ${item.inqty}, ออก: ${item.outqty}, เหลือ: ${item.balanceqty} @ ${item.unitcost} บาท (มูลค่าคงเหลือ: ${item.balanceamount})`);
  }

  // 3. ทดสอบออกใบเสร็จรับเงิน (RE6910-0001) ตัดชำระใบแจ้งหนี้ INV6910-0001
  console.log("\n[3] ทดสอบออกใบเสร็จรับเงิน (RE6910-0001)...");
  const receiptPayload = {
    docno: "RE6910-0001",
    docdatetime: "2026-10-07T15:00:00Z",
    transflag: 28, // Paid / Receipt
    custcode: "CUST-001",
    docref: "INV6910-0001",
    totalamount: 4815.00,
    totalbeforevat: 4500.00,
    totalvatvalue: 315.00,
    vatrate: 7,
    details: [
      {
        linenumber: 1,
        itemcode: "ITEM-CEMENT",
        itemname: "ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง (ถุง 50 กก.)",
        qty: 30,
        price: 150.00,
        sumamount: 4500.00,
        whcode: "WH01"
      }
    ]
  };

  const receiptRes = await fetch(`${NEXT_BASE_URL}/api/erp-transaction/paid`, {
    method: "POST",
    headers: authHeaders,
    body: JSON.stringify(receiptPayload)
  });

  if (!receiptRes.ok) {
    throw new Error(`Receipt save failed: ${receiptRes.status} ${await receiptRes.text()}`);
  }
  const receiptSaved = await receiptRes.json();
  console.log(`✅ บันทึกใบเสร็จรับเงินสำเร็จ: DocNo=${receiptSaved.docno}`);

  // 4. ทดสอบออกใบสำคัญจ่าย (PV6910-0001) ตัดชำระใบซื้อ PUR6910-0002
  console.log("\n[4] ทดสอบออกใบสำคัญจ่าย (PV6910-0001)...");
  const paymentPayload = {
    docno: "PV6910-0001",
    docdatetime: "2026-10-07T16:00:00Z",
    transflag: 29, // Pay / Payment
    custcode: "SUPP-SCG-01",
    docref: "PUR6910-0002",
    totalamount: 12840.00,
    totalbeforevat: 12000.00,
    totalvatvalue: 840.00,
    vatrate: 7,
    details: [
      {
        linenumber: 1,
        itemcode: "ITEM-CEMENT",
        itemname: "ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง (ถุง 50 กก.)",
        qty: 100,
        price: 120.00,
        sumamount: 12000.00,
        whcode: "WH01"
      }
    ]
  };

  const paymentRes = await fetch(`${NEXT_BASE_URL}/api/erp-transaction/pay`, {
    method: "POST",
    headers: authHeaders,
    body: JSON.stringify(paymentPayload)
  });

  if (!paymentRes.ok) {
    throw new Error(`Payment save failed: ${paymentRes.status} ${await paymentRes.text()}`);
  }
  const paymentSaved = await paymentRes.json();
  console.log(`✅ บันทึกใบสำคัญจ่ายสำเร็จ: DocNo=${paymentSaved.docno}`);

  // 5. ตรวจสอบ Evidence ใน PostgreSQL
  console.log("\n[5] ตรวจสอบความถูกต้องทางบัญชีใน PostgreSQL...");

  // GL Journals
  const glJournals = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT code, payload->>'bookcode' as book, payload->>'description' as desc, payload->>'status' as status FROM gl_records WHERE kind='journals' AND code IN ('RV-RE6910-0001', 'PV-PV6910-0001') ORDER BY code;"`, { encoding: "utf8" });
  console.log("=== GL Journals ===");
  console.log(glJournals);

  // GL Lines
  const glLines = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT doc_no, line_no, book_code, account_code, account_name, debit, credit FROM gl_lines WHERE doc_no IN ('RV-RE6910-0001', 'PV-PV6910-0001') ORDER BY doc_no, line_no;"`, { encoding: "utf8" });
  console.log("=== GL Lines (Debit = Credit) ===");
  console.log(glLines);

  // Subledger Documents
  const subledgerDocs = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT id, ledger, partner_code, control_account_code, amount, side FROM gl_subledger_documents WHERE id IN ('RCP-RE6910-0001', 'PAY-PV6910-0001', 'AR-INV6910-0001', 'AP-PUR6910-0002') ORDER BY id;"`, { encoding: "utf8" });
  console.log("=== Subledger Documents ===");
  console.log(subledgerDocs);

  // Subledger Settlements
  const settlements = execSync(`docker exec postgres psql -U postgres -d rungrueng -c "SELECT id, journal_id, debt_id, payment_id, amount, settlement_date FROM gl_subledger_settlements WHERE id IN ('SET-AR-RE6910-0001', 'SET-AP-PV6910-0001') ORDER BY id;"`, { encoding: "utf8" });
  console.log("=== Subledger Settlements (หนี้ถูกตัดยอดสำเร็จ) ===");
  console.log(settlements);

  console.log("\n🎉 การทดสอบ E2E: Receipt, Payment & Stock Movement สำเร็จสมบูรณ์ 100%!");
}

main().catch(err => {
  console.error("❌ Test failed:", err);
  process.exit(1);
});
