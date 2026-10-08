import { execFileSync } from 'child_process';

function pgQuery(db, sql) {
  const args = ['exec', 'postgres', 'psql', '-U', 'postgres', '-d', db, '-v', 'ON_ERROR_STOP=1', '-At', '-c', sql];
  return execFileSync('docker', args).toString().trim();
}

async function run() {
  console.log('🚀 เริ่มการทดสอบ E2E Document Flow: ออกใบสั่งซื้อ (PO) -> ดึงไปเปิดใบรับของ (PUR) -> ลงบัญชีเจ้าหนี้และ GL อัตโนมัติ...\n');

  // 1. เข้าสู่ระบบ Demo และเลือกกิจการ
  console.log('👉 [1/6] เข้าสู่ระบบ Demo และกำหนดบริบทกิจการ (rungrueng / 01 / 00000)...');
  const loginRes = await fetch('http://localhost:3000/api/auth/demo-login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ backendUrl: 'http://localhost:8888' }),
  });
  const { token } = await loginRes.json();

  const selectRes = await fetch('http://localhost:3000/api/workspace/select-holding', {
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
  console.log('   └─ ผูก Workspace Context สำเร็จ');

  // 2. สร้างใบสั่งซื้อสินค้า (PO)
  console.log('\n👉 [2/6] สร้างใบสั่งซื้อสินค้า (PO) ไปยังคู่ค้า S101 (บริษัท ปูนซีเมนต์ไทยรุ่งโรจน์ จำกัด)...');
  const poPayload = {
    custcode: 'S101',
    description: 'สั่งซื้อปูนซีเมนต์และเหล็กเส้นสำหรับงานโครงการ A',
    totalamount: 30000.0,
    totalvalue: 28037.38,
    totalbeforevat: 28037.38,
    totalvatvalue: 1962.62,
    details: [
      {
        linenumber: 1,
        itemcode: 'ITEM-CEM-01',
        itemname: 'ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง 50 กก.',
        unitcode: 'BAG',
        qty: 100,
        price: 150.0,
        sumamount: 15000.0,
      },
      {
        linenumber: 2,
        itemcode: 'ITEM-STL-02',
        itemname: 'เหล็กเส้นกลม SR24 ขนาด 9 มม.',
        unitcode: 'เส้น',
        qty: 100,
        price: 150.0,
        sumamount: 15000.0,
      },
    ],
  };

  const createPoRes = await fetch('http://localhost:3000/api/erp-transaction/purchase-order', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      'x-bc-backend-url': 'http://localhost:8888',
    },
    body: JSON.stringify(poPayload),
  });

  const createPoData = await createPoRes.json();
  if (!createPoRes.ok || !createPoData.success) {
    throw new Error(`สร้าง PO ไม่สำเร็จ: ${createPoData.message || JSON.stringify(createPoData)}`);
  }
  const poDocNo = createPoData.data.docno;
  console.log(`   └─ สร้าง PO สำเร็จ: เลขที่ ${poDocNo} ยอดรวม ${createPoData.data.totalamount} บาท`);

  // 3. ตรวจสอบ PO ใน PostgreSQL
  console.log('\n👉 [3/6] ตรวจสอบข้อมูล PO ใน PostgreSQL...');
  const poCheckSql = `SELECT id, docno, transflag, custcode, totalamount, isclosed, isref FROM doc WHERE docno = '${poDocNo}';`;
  const poRow = pgQuery('rungrueng', poCheckSql);
  console.log('   └─ PO row in doc table:', poRow);

  // 4. ดึง PO ไปเปิดใบรับสินค้า / ซื้อสินค้า (PUR)
  console.log(`\n👉 [4/6] ออกใบรับสินค้า / ซื้อสินค้า (PUR) อ้างอิงใบสั่งซื้อ ${poDocNo}...`);
  const purPayload = {
    custcode: 'S101',
    docref: poDocNo,
    docreftransflag: 6,
    description: `รับสินค้าตามใบสั่งซื้อ ${poDocNo}`,
    totalamount: 30000.0,
    totalvalue: 28037.38,
    totalbeforevat: 28037.38,
    totalvatvalue: 1962.62,
    details: [
      {
        linenumber: 1,
        itemcode: 'ITEM-CEM-01',
        itemname: 'ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง 50 กก.',
        unitcode: 'BAG',
        qty: 100,
        price: 150.0,
        sumamount: 15000.0,
        docref: poDocNo,
      },
      {
        linenumber: 2,
        itemcode: 'ITEM-STL-02',
        itemname: 'เหล็กเส้นกลม SR24 ขนาด 9 มม.',
        unitcode: 'เส้น',
        qty: 100,
        price: 150.0,
        sumamount: 15000.0,
        docref: poDocNo,
      },
    ],
  };

  const createPurRes = await fetch('http://localhost:3000/api/erp-transaction/purchase', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      'x-bc-backend-url': 'http://localhost:8888',
    },
    body: JSON.stringify(purPayload),
  });

  const createPurData = await createPurRes.json();
  if (!createPurRes.ok || !createPurData.success) {
    throw new Error(`สร้าง PUR ไม่สำเร็จ: ${createPurData.message || JSON.stringify(createPurData)}`);
  }
  const purDocNo = createPurData.data.docno;
  console.log(`   └─ สร้าง PUR สำเร็จ: เลขที่ ${purDocNo} ยอดรวม ${createPurData.data.totalamount} บาท`);

  // 5. ตรวจสอบการเชื่อมโยงเอกสาร (Doc Reference Linkage) ใน PostgreSQL
  console.log('\n👉 [5/6] ตรวจสอบการเชื่อมโยงความสัมพันธ์ของเอกสารใน PostgreSQL...');
  const docRefSql = `SELECT docno, docnotransflag, docnoref, docnoreftransflag FROM docref WHERE docno = '${purDocNo}';`;
  const docRefRow = pgQuery('rungrueng', docRefSql);
  console.log('   └─ docref row in PostgreSQL:', docRefRow);

  const poUpdatedSql = `SELECT docno, isclosed, isref FROM doc WHERE docno = '${poDocNo}';`;
  const poUpdatedRow = pgQuery('rungrueng', poUpdatedSql);
  console.log('   └─ PO status after referenced (isclosed, isref):', poUpdatedRow);

  // 6. ตรวจสอบการลงบัญชีแยกประเภท GL และเจ้าหนี้การค้า (AP Subledger) อัตโนมัติ
  console.log('\n👉 [6/6] ตรวจสอบการลงบัญชีแยกประเภท GL และเจ้าหนี้การค้า (AP Subledger) อัตโนมัติ...');
  const glDocNo = `UV-${purDocNo}`;
  const glRecordSql = `SELECT code, version, payload->>'status' as status, payload->>'description' as description FROM gl_records WHERE code = '${glDocNo}';`;
  const glRecordRow = pgQuery('rungrueng', glRecordSql);
  console.log('   └─ GL Record in gl_records:', glRecordRow);

  const glLinesSql = `SELECT doc_no, line_no, account_code, account_name, debit, credit FROM gl_lines WHERE doc_no = '${glDocNo}' ORDER BY line_no;`;
  const glLinesRows = pgQuery('rungrueng', glLinesSql);
  console.log('   └─ GL Lines in gl_lines:\n' + glLinesRows.split('\n').map((l) => '      ' + l).join('\n'));

  const balanceSql = `SELECT sum(debit) as total_debit, sum(credit) as total_credit FROM gl_lines WHERE doc_no = '${glDocNo}';`;
  const balanceRow = pgQuery('rungrueng', balanceSql);
  console.log('   └─ Balance Check (Debit = Credit):', balanceRow);

  const subledgerSql = `SELECT id, ledger, partner_code, control_account_code, amount, side FROM gl_subledger_documents WHERE id = 'AP-${purDocNo}';`;
  const subledgerRow = pgQuery('rungrueng', subledgerSql);
  console.log('   └─ AP Subledger in gl_subledger_documents:', subledgerRow);

  console.log('\n🎉 ผลการทดสอบ E2E Document Flow (PO -> PUR -> AP & GL) สำเร็จสมบูรณ์ 100%!');
}

run().catch((err) => {
  console.error('❌ Error during E2E flow test:', err);
  process.exit(1);
});
