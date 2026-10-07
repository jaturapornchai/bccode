import { execFileSync } from 'child_process';
import { randomUUID } from 'crypto';

function pgQuery(db, sql) {
  const args = ['exec', 'postgres', 'psql', '-U', 'postgres', '-d', db, '-v', 'ON_ERROR_STOP=1', '-At', '-c', sql];
  return execFileSync('docker', args).toString().trim();
}

async function run() {
  console.log('🚀 เริ่มการทดสอบลึก GL Audit & Posting Flow...');

  // 1. Demo Login
  console.log('👉 [1/6] เข้าสู่ระบบ Demo...');
  const loginRes = await fetch('http://localhost:3000/api/auth/demo-login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ backendUrl: 'http://localhost:8888' }),
  });
  const loginData = await loginRes.json();
  const token = loginData.token;
  console.log('   └─ Token ได้รับเรียบร้อย:', token.slice(0, 16) + '...');

  // 2. Select Workspace Holding / Company / Branch context
  console.log('👉 [2/6] กำหนดบริบทกิจการ (rungrueng / 01 / 00000)...');
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

  const now = new Date();
  const dateStr = now.toISOString().split('T')[0];
  const docNo = `JV-AUDIT-${Date.now().toString().slice(-6)}`;
  const requestId = randomUUID();

  // 3. Create Journal Entry
  console.log(`👉 [3/6] บันทึกรายวัน JV ใหม่ (${docNo})...`);
  const journalPayload = {
    resource: 'journals',
    action: 'create',
    requestid: requestId,
    journal: {
      docno: docNo,
      date: dateStr,
      bookcode: 'JV',
      branchcode: '00000',
      fiscalyear: '2569',
      kind: 'manual',
      description: 'ทดสอบ UAT GL Audit & Posting Flow (จ่ายค่าขนส่งสินค้าด้วยเงินสด)',
      lines: [
        {
          accountcode: '52300',
          accountname: 'ค่าขนส่งสินค้า',
          debit: '12500.00',
          credit: '0.00',
          description: 'ค่าขนส่งสินค้าเข้าคลัง',
        },
        {
          accountcode: '11110',
          accountname: 'เงินสดในมือ',
          debit: '0.00',
          credit: '12500.00',
          description: 'จ่ายเงินสดในมือ',
        },
      ],
    },
  };

  const createRes = await fetch('http://localhost:3000/api/gl/command', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      'x-bc-backend-url': 'http://localhost:8888',
    },
    body: JSON.stringify(journalPayload),
  });

  const createData = await createRes.json();
  console.log('   └─ Create Response Status:', createRes.status);
  console.log('   └─ Create Result:', createData);

  if (!createRes.ok || !createData.success) {
    throw new Error(`บันทึก JV ไม่สำเร็จ: ${createData.message || JSON.stringify(createData)}`);
  }

  const createdId = createData.data?.id;
  if (!createdId) throw new Error('ไม่ได้รับ ID ของ Journal ที่สร้าง');

  // 4. Verify in PostgreSQL immediately after CREATE
  console.log('👉 [4/6] ตรวจสอบข้อมูลใน PostgreSQL หลัง CREATE...');
  const countSql = `SELECT count(*) FROM gl_records WHERE id = '${createdId}';`;
  const count = pgQuery('rungrueng', countSql);
  console.log('   └─ PostgreSQL gl_records count:', count);
  if (count !== '1') throw new Error(`ไม่พบเอกสาร ${createdId} ใน PostgreSQL!`);

  const recordSql = `SELECT company, kind, id, code, version, payload->>'status' as status, payload->>'description' as description FROM gl_records WHERE id = '${createdId}';`;
  const recordRow = pgQuery('rungrueng', recordSql);
  console.log('   └─ Record row in PostgreSQL:', recordRow);

  // 5. Post Journal (ผ่านรายการ)
  console.log(`👉 [5/6] ดำเนินการผ่านรายการ (Post Journal: ${docNo})...`);
  const postPayload = {
    resource: 'journals',
    action: 'post',
    requestid: randomUUID(),
    id: createdId,
    version: 1,
  };

  const postRes = await fetch('http://localhost:3000/api/gl/command', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      'x-bc-backend-url': 'http://localhost:8888',
    },
    body: JSON.stringify(postPayload),
  });

  const postData = await postRes.json();
  console.log('   └─ Post Response Status:', postRes.status);
  console.log('   └─ Post Result:', postData);

  if (!postRes.ok || !postData.success) {
    throw new Error(`ผ่านรายการไม่สำเร็จ: ${postData.message || JSON.stringify(postData)}`);
  }

  // 6. Verify in PostgreSQL immediately after POST
  console.log('👉 [6/6] ตรวจสอบสถานะใน PostgreSQL หลัง POST...');
  const postStatusSql = `SELECT company, id, code, version, payload->>'status' as status, payload->>'postedat' as postedat FROM gl_records WHERE id = '${createdId}';`;
  const postStatusRow = pgQuery('rungrueng', postStatusSql);
  console.log('   └─ Post status in gl_records:', postStatusRow);

  const linesSql = `SELECT doc_no, line_no, account_code, debit, credit FROM gl_lines WHERE doc_no = '${docNo}' ORDER BY line_no;`;
  const linesRows = pgQuery('rungrueng', linesSql);
  console.log('   └─ Lines in gl_lines:\n' + linesRows.split('\n').map((l) => '      ' + l).join('\n'));

  const balanceCheckSql = `SELECT sum(debit) as total_debit, sum(credit) as total_credit FROM gl_lines WHERE doc_no = '${docNo}';`;
  const balanceRow = pgQuery('rungrueng', balanceCheckSql);
  console.log('   └─ Balance Check (Debit/Credit):', balanceRow);

  console.log('\n🎉 ผลการทดสอบ GL Audit & Posting Flow สำเร็จสมบูรณ์ 100%!');
}

run().catch((err) => {
  console.error('❌ Error during GL Audit:', err);
  process.exit(1);
});

