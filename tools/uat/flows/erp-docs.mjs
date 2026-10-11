// ERP documents: every key in the doctype registry → Create → Read → Update → Delete (soft) and
// Create → Cancel (header-level), checking PostgreSQL after each operation.
import fs from 'node:fs';
import { dec, baht, sameDecimal, sqlText, snapshot, diff } from '../lib.mjs';

const ROOT = new URL('../../../', import.meta.url);
const TABLES = ['doc', 'docdetail', 'gl_records', 'gl_lines', 'stock_ledger', 'docref'];

// Docs that post to GL / stock on save (doctype constants Purchase, PurchaseReturn, StockPickup, Expense,
// Paid, Pay, SaleInvoice, SaleInvoiceReturn): their GL/stock row changes are reported, not asserted.
const AUTO_POST = new Set(['purchase', 'sale-invoice', 'sale-invoice-return', 'purchase-return', 'paid', 'pay', 'expense', 'stock-pickup-product']);
const LINE_KEYS = new Set([
  'stock-balance', 'stock-receive-product', 'stock-pickup-product', 'stock-return-product', 'stock-transfer', 'stock-adjustment',
  'stock-issue-request', 'stock-transfer-request', 'goods-inspection', 'stock-count', 'quotation', 'sale-order', 'sale-invoice',
  'sale-invoice-return', 'ardebitnote', 'purchase-requisition', 'purchase-order', 'purchase', 'purchase-return', 'rfq', 'goodsreceive',
  'invoicereceive', 'landedcost', 'apdebitnote', 'apcreditnote',
]);
const SALES_SIDE = new Set([
  'quotation', 'sale-order', 'sale-invoice', 'sale-invoice-return', 'ardebitnote', 'debtorbeginningbalance', 'receivableother',
  'billingnote', 'paid', 'receivedeposit', 'receiveadvance', 'receiveadvancerefund', 'temporaryreceipt', 'arbaddebt',
  'chequereceived', 'otherincome', 'creditcard',
]);

// Cheque doctypes are owned by /cheque/v1 and rejected on /transaction/:module (cheque_moved_test.go)
const CHEQUE_FLAGS = new Set([61, 418, 419, 420, 421, 422, 423, 424, 425, 426]);

/** Parse `{"key", flag|Const, "PREFIX", StockX, "screen"}` rows of backend/internal/doctype/doctype.go. */
export function readRegistry() {
  const go = fs.readFileSync(new URL('backend/internal/doctype/doctype.go', ROOT), 'utf8');
  const consts = {};
  for (const m of go.matchAll(/^\s+([A-Z][A-Za-z]+)\s*=\s*([0-9]+)\s*$/gm)) consts[m[1]] = Number(m[2]);
  return [...go.matchAll(/\{"([a-z0-9-]+)",\s*([A-Za-z0-9]+),\s*"([A-Z0-9-]+)",\s*(Stock[A-Za-z]+)/g)]
    .map((m) => ({
      key: m[1],
      flag: consts[m[2]] ?? Number(m[2]),
      prefix: m[3],
      stock: m[4],
    }))
    .filter((t) => !CHEQUE_FLAGS.has(t.flag));
}

function payload(t, version, docdate) {
  const lines = LINE_KEYS.has(t.key);
  const qty = version === 1 ? 10 : 12;
  const priceSatang = 15000; // 150.00 บาท/ถุง
  const before = qty * priceSatang;
  const vat = (before * 7) / 100; // 7% of a multiple of 100 satang → exact integer satang
  const total = before + vat;
  return {
    body: {
      transflag: t.flag,
      docno: '',
      docdatetime: `${docdate}T09:00:00Z`,
      custcode: lines || SALES_SIDE.has(t.key) ? (SALES_SIDE.has(t.key) ? 'C101' : 'S101') : '',
      description:
        version === 1
          ? `${t.key} ซื้อปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง 50 กก. ให้หน้างานบ้านเลขที่ 88/12`
          : `${t.key} แก้ไขจำนวนเป็น 12 ถุง ตามใบส่งของจริง`,
      totalamount: dec(total),
      totalbeforevat: dec(before),
      totalvatvalue: dec(vat),
      vattype: 1,
      vatrate: 7,
      status: 1,
      branchcode: '00000',
      details: lines
        ? [{ linenumber: 1, itemcode: 'ITEM-CEM-01', itemname: 'ปูนซีเมนต์ปอร์ตแลนด์ ตราช้าง 50 กก.', unitcode: 'BAG', qty, price: dec(priceSatang), sumamount: dec(before), whcode: 'WH01' }]
        : [],
    },
    total: baht(total),
    lineCount: lines ? 1 : 0,
  };
}

export async function run({ api, psql, recorder, only, docdate, cleanup }) {
  const registry = readRegistry().filter((t) => !only || only.has(t.key));
  const { check: regCheck } = recorder('erp:registry');
  regCheck('registry.parsed', registry.length > 0, `${registry.length} doctypes`);

  for (const t of registry) {
    const { check } = recorder(`erp:${t.key}`);
    try {
      const v1 = payload(t, 1, docdate);
      const s0 = snapshot(psql, TABLES);
      const c = await api('POST', `/transaction/${t.key}`, v1.body);
      const docno = c.json?.data?.docno;
      if (!check('C.api', (c.status === 200 || c.status === 201) && docno, `${c.status} ${c.text}`)) continue;
      const id = c.json?.data?.id || psql(`select id from doc where docno=${sqlText(docno)} and not isdelete order by id desc limit 1`);
      cleanup.push({ kind: 'erp-doc', key: t.key, id, docno });
      const [flag, pgDocno, total, isdelete] = psql(`select transflag,docno,totalamount,isdelete from doc where id=${Number(id)}`).split('|');
      check('C.pg.row', Number(flag) === t.flag && pgDocno === docno && isdelete === 'f' && sameDecimal(total, v1.total), `flag=${flag} total=${total} want ${v1.total}`);
      check('C.docno.prefix', docno.startsWith(t.prefix), docno);
      const lines = Number(psql(`select count(*) from docdetail where docno=${sqlText(docno)}`));
      check('C.pg.lines', lines === v1.lineCount, `docdetail=${lines}`);
      const dC = diff(s0, snapshot(psql, TABLES));
      if (!AUTO_POST.has(t.key)) check('C.pg.no-gl', dC.doc === 1 && !dC.gl_lines && (t.stock !== 'StockNone' || !dC.stock_ledger), `${t.stock} ${JSON.stringify(dC)}`);
      else check('C.pg.posted(info)', true, JSON.stringify(dC));

      const g = await api('GET', `/transaction/${t.key}/${id}`);
      check('R.get', g.status === 200 && g.json?.data?.docno === docno && sameDecimal(g.json?.data?.totalamount, v1.total), `${g.status} total=${g.json?.data?.totalamount}`);
      const l = await api('GET', `/transaction/${t.key}/list`);
      check('R.list', l.status === 200 && (l.json?.data ?? []).some((x) => String(x.id) === String(id)), `${l.status} n=${l.json?.data?.length}`);

      const v2 = payload(t, 2, docdate);
      const u = await api('PUT', `/transaction/${t.key}/${id}`, { ...v2.body, id: String(id), docno });
      check('U.api', u.status === 200, `${u.status} ${u.text}`);
      const [uid, udocno, utotal] = psql(`select id,docno,totalamount from doc where id=${Number(id)}`).split('|');
      check('U.pg.same-row', uid === String(id) && udocno === docno && sameDecimal(utotal, v2.total), `total=${utotal} want ${v2.total}`);
      check('U.pg.no-orphan', psql(`select count(*) from doc where docno=${sqlText(docno)}`) === '1', 'one row per docno');

      const sD = snapshot(psql, TABLES);
      const del = await api('DELETE', `/transaction/${t.key}/${id}`);
      check('D.api', del.status === 200, `${del.status} ${del.text}`);
      check('D.pg.soft-deleted', psql(`select isdelete from doc where id=${Number(id)}`) === 't', 'isdelete');
      const l2 = await api('GET', `/transaction/${t.key}/list`);
      check('D.list.hidden', !(l2.json?.data ?? []).some((x) => String(x.id) === String(id)), '');
      const dD = diff(sD, snapshot(psql, TABLES));
      check('D.pg.others-intact', !dD.doc, JSON.stringify(dD));

      // ยกเลิกทั้งใบ (header-level cancel, ลุงจืด 2026-10-11)
      const c2 = await api('POST', `/transaction/${t.key}`, payload(t, 1, docdate).body);
      const docno2 = c2.json?.data?.docno;
      if (!check('X.create', (c2.status === 200 || c2.status === 201) && docno2, `${c2.status} ${c2.text}`)) continue;
      const id2 = c2.json?.data?.id || psql(`select id from doc where docno=${sqlText(docno2)} and not isdelete order by id desc limit 1`);
      cleanup.push({ kind: 'erp-doc', key: t.key, id: id2, docno: docno2 });
      const x = await api('POST', `/transaction/${t.key}/${id2}/cancel`, { reason: 'ยกเลิกเพราะออกเอกสารซ้ำกับใบเดิม' });
      check('X.api', x.status === 200, `${x.status} ${x.text}`);
      const [iscancel, xdel] = psql(`select iscancel,isdelete from doc where id=${Number(id2)}`).split('|');
      check('X.pg.cancelled', iscancel === 't' && xdel === 'f', `iscancel=${iscancel} isdelete=${xdel}`);
    } catch (e) {
      check('EXCEPTION', false, e?.message ?? e);
    }
  }
}
