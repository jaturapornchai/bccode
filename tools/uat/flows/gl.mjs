// General ledger through the real BFF (/api/gl/command): journal draft → update → post → reverse,
// draft delete, chart of accounts CRUD, journal books CRUD + guard cases. PostgreSQL checked after every step.
import { randomUUID } from 'node:crypto';
import { sameDecimal, sqlText } from '../lib.mjs';

const COMPANY = '01';
// Business errors come back as HTTP 200/4xx with success:false (BFF relays the backend error contract).
const rejected = (r) => r.status >= 400 || r.json?.success === false;
const why = (r) => `${r.status} ${r.json?.code ?? r.json?.errorcode ?? ''}`;

export async function run({ fe, psql, recorder, docdate, cleanup }) {
  const command = (resource, action, fields = {}) => {
    const requestid = randomUUID();
    return fe('POST', '/api/gl/command', { requestid, resource, action, id: '', version: 0, ...fields }).then((r) => ({ ...r, requestid }));
  };
  const rec = (kind, id) => psql(`select version, payload->>'status', coalesce(payload->>'isdeleted','false'), jsonb_array_length(coalesce(payload->'lines','[]'::jsonb)), payload->>'docno' from gl_records where company='${COMPANY}' and kind='${kind}' and id=${sqlText(id)}`).split('|');
  const lineTotals = (journalId) => psql(`select count(*), coalesce(sum(debit),0), coalesce(sum(credit),0) from gl_lines where company='${COMPANY}' and journal_id=${sqlText(journalId)}`).split('|');
  const events = (requestid) => Number(psql(`select count(*) from gl_events where company='${COMPANY}' and payload->>'requestid'=${sqlText(requestid)}`));
  const fiscalyear = String(Number(docdate.slice(0, 4)) + 543);

  const journal = (amount, description) => ({
    date: docdate,
    bookcode: 'JV',
    fiscalyear,
    branchcode: '00000',
    kind: 'manual',
    description,
    reference: 'PV6910-0153',
    lines: [
      { accountcode: '5212', description: 'ค่านายหน้าขาย นายสมศักดิ์ ใจดี', debit: amount, credit: '0.00', departmentcode: '', projectcode: '', cashflow: '' },
      { accountcode: '1111', description: 'จ่ายเงินสด', debit: '0.00', credit: amount, departmentcode: '', projectcode: '', cashflow: '' },
    ],
  });

  // ── Journal: draft → read → update → idempotent replay → stale version → post → guards → reverse
  {
    const { check } = recorder('gl:journal');
    const c = await command('journals', 'create', { journal: journal('1250.00', 'จ่ายค่านายหน้าขายเดือนตุลาคม 2569 ให้นายสมศักดิ์ ใจดี') });
    const id = c.json?.data?.id;
    if (check('C.api', c.status === 200 && id, `${c.status} ${c.text}`)) {
      cleanup.push({ kind: 'gl-journal', id });
      let [version, status, deleted, nlines, docno] = rec('journals', id);
      check('C.pg.draft', status === 'draft' && deleted === 'false' && nlines === '2' && docno, `status=${status} lines=${nlines} docno=${docno}`);
      check('C.pg.no-gl_lines', lineTotals(id)[0] === '0', 'draft must not write gl_lines');
      check('C.pg.event', events(c.requestid) === 1, 'one gl_events row per requestid');

      const g = await fe('GET', `/api/gl/journals/${id}`);
      check('R.get', g.status === 200 && (g.json?.data?.docno ?? g.json?.data?.journal?.docno) === docno, `${g.status}`);

      const u = await command('journals', 'update', { id, version: Number(version), journal: { ...journal('1375.00', 'จ่ายค่านายหน้าขายเดือนตุลาคม 2569 (แก้ยอดตามใบสำคัญจ่าย)'), docno } });
      check('U.api', u.status === 200, `${u.status} ${u.text}`);
      const after = rec('journals', id);
      check('U.pg.same-row', Number(after[0]) === Number(version) + 1 && after[4] === docno && psql(`select count(*) from gl_records where company='${COMPANY}' and kind='journals' and payload->>'docno'=${sqlText(docno)}`) === '1', `version ${version}→${after[0]}`);
      check('U.pg.amount', sameDecimal(psql(`select payload->'lines'->0->>'debit' from gl_records where company='${COMPANY}' and kind='journals' and id=${sqlText(id)}`), '1375.00'), 'line 1 debit');

      const replay = await fe('POST', '/api/gl/command', { requestid: u.requestid, resource: 'journals', action: 'update', id, version: Number(version), journal: { ...journal('1375.00', 'ส่งซ้ำ'), docno } });
      check('U.idempotent-replay', events(u.requestid) === 1 && rec('journals', id)[0] === after[0], `${replay.status} events=${events(u.requestid)}`);
      const stale = await command('journals', 'update', { id, version: Number(version), journal: { ...journal('9999.00', 'เวอร์ชันเก่า'), docno } });
      check('U.stale-version-rejected', rejected(stale) && rec('journals', id)[0] === after[0], why(stale));

      const p = await command('journals', 'post', { id, version: Number(after[0]) });
      check('P.api', p.status === 200, `${p.status} ${p.text}`);
      [version, status] = rec('journals', id);
      const [n, dr, cr] = lineTotals(id);
      check('P.pg.posted', status === 'posted', `status=${status}`);
      check('P.pg.gl_lines.balanced', n === '2' && sameDecimal(dr, '1375.00') && sameDecimal(cr, '1375.00'), `n=${n} dr=${dr} cr=${cr}`);

      const upd = await command('journals', 'update', { id, version: Number(version), journal: { ...journal('1.00', 'แก้หลังผ่านรายการ'), docno } });
      check('P.update-blocked', rejected(upd) && rec('journals', id)[0] === version, why(upd));
      const del = await command('journals', 'delete', { id, version: Number(version) });
      check('P.delete-blocked', rejected(del) && rec('journals', id)[2] === 'false', why(del));

      const next = await fe('GET', `/api/gl/journal-next-docno?bookcode=JV&from=${docdate}`);
      const revDocno = next.json?.data?.docno ?? next.json?.data ?? '';
      const r = await command('journals', 'reverse', { id, version: Number(version), docno: String(revDocno), date: docdate, reason: 'บันทึกผิดผู้รับเงิน กลับรายการเพื่อบันทึกใหม่' });
      check('V.api', r.status === 200, `${r.status} ${r.text}`);
      check('V.pg.original-reversed', rec('journals', id)[1] === 'reversed', `status=${rec('journals', id)[1]}`);
      const [revId, revStatus] = psql(`select id, payload->>'status' from gl_records where company='${COMPANY}' and kind='journals' and payload->>'reversalof'=${sqlText(id)}`).split('|');
      check('V.pg.reversal-posted', revId && revStatus === 'posted', `reversal ${revId} ${revStatus}`);
      const net = psql(`select coalesce(sum(debit - credit),0), count(*) from gl_lines where company='${COMPANY}' and journal_id in (${sqlText(id)}, ${sqlText(revId || id)}) group by account_code order by account_code`).split('\n');
      check('V.pg.net-zero', net.length === 2 && net.every((row) => sameDecimal(row.split('|')[0], '0') && row.endsWith('|2')), net.join(' ; '));
    }

    const unbalanced = await command('journals', 'create', { journal: { ...journal('500.00', 'ใบไม่สมดุล'), lines: [journal('500.00', '').lines[0], { ...journal('400.00', '').lines[1] }] } });
    check('C.unbalanced-rejected', rejected(unbalanced) && !unbalanced.json?.data?.id, why(unbalanced));
    const numeric = await fe('POST', '/api/gl/command', { requestid: randomUUID(), resource: 'journals', action: 'create', id: '', version: 0, journal: { ...journal('10.00', 'เงินเป็นตัวเลข'), lines: [{ ...journal('10.00', '').lines[0], debit: 10 }, journal('10.00', '').lines[1]] } });
    check('C.numeric-money-rejected', numeric.status === 400, why(numeric));
  }

  // ── Journal draft delete
  {
    const { check } = recorder('gl:journal-draft-delete');
    const c = await command('journals', 'create', { journal: journal('420.00', 'ค่าน้ำมันรถส่งของ สาขาสำนักงานใหญ่') });
    const id = c.json?.data?.id;
    if (check('C.api', c.status === 200 && id, `${c.status} ${c.text}`)) {
      cleanup.push({ kind: 'gl-journal', id });
      const [version] = rec('journals', id);
      const d = await command('journals', 'delete', { id, version: Number(version) });
      check('D.api', d.status === 200, `${d.status} ${d.text}`);
      check('D.pg.soft-deleted', rec('journals', id)[2] === 'true' && lineTotals(id)[0] === '0', `isdeleted=${rec('journals', id)[2]}`);
    }
  }

  // ── Chart of accounts
  {
    const { check } = recorder('gl:account');
    const code = ['5219', '5218', '5217', '5216'].find((x) => psql(`select count(*) from gl_records where company='${COMPANY}' and kind='accounts' and code='${x}'`) === '0');
    if (check('setup.free-code', code, 'no free code 5216-5219')) {
      const account = (name) => ({ accountcode: code, names: [{ code: 'th', name }], accounttype: 'expense', normalbalance: 'debit', allowposting: true, isactive: true, parentaccountcode: '5200', level: 3 });
      const c = await command('accounts', 'create', { account: account('ค่าส่งเสริมการขาย ณ จุดขาย') });
      const id = c.json?.data?.id;
      if (check('C.api', c.status === 200 && id, `${c.status} ${c.text}`)) {
        cleanup.push({ kind: 'gl-account', id, code });
        const row = () => psql(`select version, payload->'names'->0->>'name', coalesce(payload->>'isdeleted','false') from gl_records where company='${COMPANY}' and kind='accounts' and id=${sqlText(id)}`).split('|');
        check('C.pg.row', row()[1] === 'ค่าส่งเสริมการขาย ณ จุดขาย', row().join('|'));
        const u = await command('accounts', 'update', { id, version: Number(row()[0]), account: account('ค่าส่งเสริมการขายและสาธิตสินค้า ณ จุดขาย') });
        check('U.api', u.status === 200, `${u.status} ${u.text}`);
        check('U.pg.same-row', row()[1] === 'ค่าส่งเสริมการขายและสาธิตสินค้า ณ จุดขาย' && psql(`select count(*) from gl_records where company='${COMPANY}' and kind='accounts' and code='${code}'`) === '1', row().join('|'));
        const d = await command('accounts', 'delete', { id, version: Number(row()[0]) });
        check('D.api', d.status === 200, `${d.status} ${d.text}`);
        check('D.pg.soft-deleted', row()[2] === 'true', row().join('|'));
      }
    }
    const usedId = psql(`select id from gl_records where company='${COMPANY}' and kind='accounts' and code='1111'`);
    const usedVersion = psql(`select version from gl_records where company='${COMPANY}' and kind='accounts' and code='1111'`);
    const blocked = await command('accounts', 'delete', { id: usedId, version: Number(usedVersion) });
    check('D.in-use-blocked', rejected(blocked) && psql(`select coalesce(payload->>'isdeleted','false') from gl_records where company='${COMPANY}' and kind='accounts' and code='1111'`) === 'false', why(blocked));
  }

  // ── Journal books (user-defined master)
  {
    const { check } = recorder('gl:journal-book');
    const code = ['JC', 'JD', 'JE', 'JF'].find((x) => psql(`select count(*) from gl_records where company='${COMPANY}' and kind='journal-books' and code='${x}'`) === '0');
    if (check('setup.free-code', code, 'no free code JC-JF')) {
      const master = (name) => ({ code, name, nameen: 'Petty cash journal', booktype: 1, isactive: true });
      const c = await command('journal-books', 'create', { master: master('สมุดรายวันเงินสดย่อย') });
      const id = c.json?.data?.id;
      if (check('C.api', c.status === 200 && id, `${c.status} ${c.text}`)) {
        cleanup.push({ kind: 'gl-journal-book', id, code });
        const row = () => psql(`select version, payload->>'name', coalesce(payload->>'isdeleted','false') from gl_records where company='${COMPANY}' and kind='journal-books' and id=${sqlText(id)}`).split('|');
        check('C.pg.row', row()[1] === 'สมุดรายวันเงินสดย่อย', row().join('|'));
        const u = await command('journal-books', 'update', { id, version: Number(row()[0]), master: master('สมุดรายวันเงินสดย่อย สำนักงานใหญ่') });
        check('U.api', u.status === 200, `${u.status} ${u.text}`);
        check('U.pg.same-row', row()[1] === 'สมุดรายวันเงินสดย่อย สำนักงานใหญ่', row().join('|'));
        const d = await command('journal-books', 'delete', { id, version: Number(row()[0]) });
        check('D.api', d.status === 200, `${d.status} ${d.text}`);
        check('D.pg.soft-deleted', row()[2] === 'true', row().join('|'));
      }
    }
    const jv = psql(`select id||'|'||version from gl_records where company='${COMPANY}' and kind='journal-books' and code='JV'`).split('|');
    const blocked = await command('journal-books', 'delete', { id: jv[0], version: Number(jv[1]) });
    check('D.in-use-blocked', rejected(blocked) && psql(`select coalesce(payload->>'isdeleted','false') from gl_records where company='${COMPANY}' and kind='journal-books' and code='JV'`) === 'false', why(blocked));
  }
}
