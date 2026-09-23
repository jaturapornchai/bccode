"use client";

import { useEffect, useState } from "react";
import { FileText, Loader2 } from "lucide-react";
import { useBackendText } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { DateField } from "@/components/ui/date-time-field";
import { Input } from "@/components/ui/input";
import { formatAmount } from "@/lib/general-ledger";
import type { LanguageCode } from "@/lib/i18n";
import { requestWhtCertificatePdf, type CompanyHeader, type WhtReportRow } from "@/lib/thai-tax";

// ใบ 50 ทวิ: จอนี้เตรียมข้อมูลให้นักบัญชีตรวจ/เลือก แล้วให้ backend สร้าง PDF บนแบบฟอร์มกรมสรรพากร
// ยอดเงินมาจากรายงานภาษีหัก ณ ที่จ่าย (backend) อ่านอย่างเดียว — จอไม่คำนวณยอดรวมหรือตัวอักษรเอง
// ไม่มีตัวเลือกเงินเดือน (40(1), ภ.ง.ด.1ก) เพราะ BC ไม่ทำระบบเงินเดือน

const FORM_OPTIONS = [
  { value: "53", key: "wht_cert_ui_form_53", th: "ภ.ง.ด.53 (นิติบุคคล)" },
  { value: "3", key: "wht_cert_ui_form_3", th: "ภ.ง.ด.3 (บุคคลธรรมดา)" },
  { value: "2", key: "wht_cert_ui_form_2", th: "ภ.ง.ด.2 (ดอกเบี้ย/เงินปันผล)" },
  { value: "3a", key: "wht_cert_ui_form_3a", th: "ภ.ง.ด.3ก" },
  { value: "2a", key: "wht_cert_ui_form_2a", th: "ภ.ง.ด.2ก" },
];

const INCOME_OPTIONS = [
  { value: "3_tres", key: "wht_cert_ui_income_3_tres", th: "5. ตามมาตรา 3 เตรส (ค่าบริการ ค่าเช่า ค่าขนส่ง ค่าโฆษณา ค่าจ้างทำของ ฯลฯ)" },
  { value: "40_2", key: "wht_cert_ui_income_40_2", th: "2. ค่าธรรมเนียม ค่านายหน้า ฯลฯ 40 (2)" },
  { value: "40_3", key: "wht_cert_ui_income_40_3", th: "3. ค่าแห่งลิขสิทธิ์ ฯลฯ 40 (3)" },
  { value: "40_4a", key: "wht_cert_ui_income_40_4a", th: "4. (ก) ดอกเบี้ย ฯลฯ 40 (4) (ก)" },
  { value: "40_4b_1_1", key: "wht_cert_ui_income_div_1_1", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — กำไรเสียภาษีร้อยละ 30" },
  { value: "40_4b_1_2", key: "wht_cert_ui_income_div_1_2", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — กำไรเสียภาษีร้อยละ 25" },
  { value: "40_4b_1_3", key: "wht_cert_ui_income_div_1_3", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — กำไรเสียภาษีร้อยละ 20" },
  { value: "40_4b_1_4", key: "wht_cert_ui_income_div_1_4", th: "4. (ข) เงินปันผล ได้เครดิตภาษี — อัตราอื่น (ระบุอัตรา)" },
  { value: "40_4b_2_1", key: "wht_cert_ui_income_div_2_1", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — กิจการได้รับยกเว้นภาษี" },
  { value: "40_4b_2_2", key: "wht_cert_ui_income_div_2_2", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — เงินปันผลที่ได้รับยกเว้น" },
  { value: "40_4b_2_3", key: "wht_cert_ui_income_div_2_3", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — หักผลขาดทุนยกมาไม่เกิน 5 ปี" },
  { value: "40_4b_2_4", key: "wht_cert_ui_income_div_2_4", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — วิธีส่วนได้เสีย (equity method)" },
  { value: "40_4b_2_5", key: "wht_cert_ui_income_div_2_5", th: "4. (ข) เงินปันผล ไม่ได้เครดิต — อื่น ๆ (ระบุ)" },
  { value: "other", key: "wht_cert_ui_income_other", th: "6. อื่น ๆ (ระบุ)" },
];
const INCOME_WITH_NOTE = new Set(["40_4b_1_4", "40_4b_2_5", "other"]);

const CONDITION_OPTIONS = [
  { value: "withhold", key: "wht_cert_ui_condition_withhold", th: "(1) หัก ณ ที่จ่าย" },
  { value: "always", key: "wht_cert_ui_condition_always", th: "(2) ออกให้ตลอดไป" },
  { value: "once", key: "wht_cert_ui_condition_once", th: "(3) ออกให้ครั้งเดียว" },
  { value: "other", key: "wht_cert_ui_condition_other", th: "(4) อื่น ๆ (ระบุ)" },
];

const selectClass =
  "w-full min-h-[2.6em] rounded-lg border border-border bg-background px-3 text-sm font-medium text-foreground shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20";

const today = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
};

// ทะเบียนบริษัทยังไม่มีที่อยู่ — จำที่อยู่ผู้หักที่กรอกล่าสุดไว้ในเครื่องต่อบริษัท ไม่ต้องพิมพ์ซ้ำทุกใบ
const payerAddressKey = (holding: string, business: string) => `bc.wht.payeraddress.${holding}.${business}`;

type Props = {
  row: WhtReportRow;
  company: CompanyHeader | null;
  holdingcode: string;
  businesscode: string;
  formType: string;
  language: LanguageCode;
};

export function WhtCertificatePanel({ row, company, holdingcode, businesscode, formType, language }: Props) {
  const tr = useBackendText();
  const [form, setForm] = useState(formType === "pnd3" ? "3" : formType === "pnd2" ? "2" : "53");
  const [incomeType, setIncomeType] = useState(formType === "pnd2" ? "40_4a" : "3_tres");
  const [incomeNote, setIncomeNote] = useState("");
  const [condition, setCondition] = useState("withhold");
  const [conditionNote, setConditionNote] = useState("");
  const [paidDate, setPaidDate] = useState(row.docdate);
  const [issueDate, setIssueDate] = useState(today);
  const [bookNo, setBookNo] = useState("");
  const [runNo, setRunNo] = useState(row.docno);
  const [sequenceNo, setSequenceNo] = useState("");
  const [payerAddress, setPayerAddress] = useState("");
  const [payeeName, setPayeeName] = useState(row.partnername ?? "");
  const [payeeTaxId, setPayeeTaxId] = useState(row.taxid ?? "");
  const [payeeAddress, setPayeeAddress] = useState(row.address ?? "");
  const [archiveCopy, setArchiveCopy] = useState(false);
  const [replacement, setReplacement] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [pdfUrl, setPdfUrl] = useState("");

  useEffect(() => {
    setPayerAddress(window.localStorage.getItem(payerAddressKey(holdingcode, businesscode)) ?? "");
  }, [holdingcode, businesscode]);
  useEffect(() => () => { if (pdfUrl) URL.revokeObjectURL(pdfUrl); }, [pdfUrl]);

  const generate = async () => {
    setBusy(true);
    setError("");
    window.localStorage.setItem(payerAddressKey(holdingcode, businesscode), payerAddress.trim());
    const result = await requestWhtCertificatePdf(holdingcode, businesscode, {
      bookno: bookNo.trim(),
      runno: runNo.trim(),
      sequenceno: sequenceNo.trim(),
      form,
      condition,
      conditionnote: condition === "other" ? conditionNote.trim() : "",
      issuedate: issueDate,
      archivecopy: archiveCopy,
      replacement,
      // ชื่อ/เลขผู้หักภาษี backend ยึดจากทะเบียนบริษัท ค่าที่ส่งเป็นค่าสำรองเมื่อทะเบียนยังว่าง
      payer: { name: company?.name ?? "", address: payerAddress.trim(), taxid: company?.taxid ?? "" },
      payee: { name: payeeName.trim(), address: payeeAddress.trim(), taxid: payeeTaxId.trim() },
      incomes: [{
        type: incomeType,
        paiddate: paidDate,
        amount: row.baseamount,
        tax: row.whtamount,
        note: INCOME_WITH_NOTE.has(incomeType) ? incomeNote.trim() : "",
      }],
    });
    setBusy(false);
    if (result.ok) {
      setPdfUrl(URL.createObjectURL(result.pdf));
      return;
    }
    setPdfUrl("");
    // code จาก backend เป็น language key (wht_cert_*) → แปลตามภาษาบนจอ ไม่ยึด Accept-Language ของเบราว์เซอร์
    const fallback = tr("wht_cert_render_failed", "สร้างหนังสือรับรอง 50 ทวิ ไม่สำเร็จ กรุณาลองใหม่อีกครั้ง");
    setError(result.error.startsWith("wht_cert_") ? tr(result.error, result.message || fallback) : result.message || fallback);
  };

  const label = (key: string, th: string) => <span className="text-[0.9rem] font-semibold text-foreground">{tr(key, th)}</span>;
  const field = "grid gap-1.5";

  return (
    <div className="grid gap-4 xl:grid-cols-[minmax(22rem,30rem)_1fr] print:hidden">
      <Card className="space-y-4 p-5 shadow-sm">
        <div className="flex items-start gap-3">
          <FileText className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden />
          <div className="space-y-0.5">
            <h2 className="text-base font-bold text-foreground">{tr("wht_cert_ui_title", "หนังสือรับรองการหักภาษี ณ ที่จ่าย (50 ทวิ)")}</h2>
            <p className="text-sm leading-relaxed text-muted-foreground">
              {tr("wht_cert_ui_hint", "ตรวจข้อมูลแล้วกดสร้าง ระบบพิมพ์ลงแบบฟอร์มของกรมสรรพากร ได้ฉบับที่ 1 และฉบับที่ 2 ในไฟล์เดียว")}
            </p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3 rounded-xl border border-border bg-muted/30 p-3 text-sm">
          <div>{tr("wht_cert_ui_amount", "จำนวนเงินที่จ่าย")}<div className="font-mono text-base font-bold text-foreground">{formatAmount(row.baseamount, 2)}</div></div>
          <div>{tr("wht_cert_ui_tax", "ภาษีที่หักและนำส่ง")}<div className="font-mono text-base font-bold text-primary">{formatAmount(row.whtamount, 2)}</div></div>
        </div>

        <label className={field}>
          {label("wht_cert_ui_form", "ลำดับที่ในแบบ (แบบยื่นรายการ)")}
          <select className={selectClass} value={form} onChange={(e) => setForm(e.target.value)} data-field="form">
            {FORM_OPTIONS.map((o) => <option key={o.value} value={o.value}>{tr(o.key, o.th)}</option>)}
          </select>
        </label>

        <label className={field}>
          {label("wht_cert_ui_income", "ประเภทเงินได้พึงประเมินที่จ่าย")}
          <select className={selectClass} value={incomeType} onChange={(e) => setIncomeType(e.target.value)} data-field="incomes[0].type">
            {INCOME_OPTIONS.map((o) => <option key={o.value} value={o.value}>{tr(o.key, o.th)}</option>)}
          </select>
        </label>
        {INCOME_WITH_NOTE.has(incomeType) && (
          <label className={field}>
            {label("wht_cert_ui_income_note", incomeType === "40_4b_1_4" ? "อัตราภาษีของกำไรสุทธิ (ร้อยละ)" : "ระบุประเภทเงินได้")}
            <Input value={incomeNote} onChange={(e) => setIncomeNote(e.target.value)} data-field="incomes[0].note" />
          </label>
        )}

        <div className="grid grid-cols-2 gap-3">
          <DateField label={tr("wht_cert_ui_paid_date", "วันที่จ่ายเงิน")} language={language} yearType="buddhist" value={paidDate} onChange={(e) => setPaidDate(e.target.value)} />
          <DateField label={tr("wht_cert_ui_issue_date", "วันที่ออกหนังสือรับรอง")} language={language} yearType="buddhist" value={issueDate} onChange={(e) => setIssueDate(e.target.value)} />
        </div>

        <fieldset className="grid gap-2">
          <legend className="mb-1.5 text-[0.9rem] font-semibold text-foreground">{tr("wht_cert_ui_condition", "ผู้จ่ายเงิน")}</legend>
          <div className="grid grid-cols-2 gap-2">
            {CONDITION_OPTIONS.map((o) => (
              <label key={o.value} className={`flex min-h-[2.6em] cursor-pointer items-center gap-2 rounded-lg border px-3 text-sm ${condition === o.value ? "border-primary bg-primary/10 font-semibold text-foreground" : "border-border text-muted-foreground"}`}>
                <input type="radio" name="wht-condition" value={o.value} checked={condition === o.value} onChange={() => setCondition(o.value)} className="accent-[var(--primary)]" />
                {tr(o.key, o.th)}
              </label>
            ))}
          </div>
          {condition === "other" && (
            <Input value={conditionNote} onChange={(e) => setConditionNote(e.target.value)} placeholder={tr("wht_cert_ui_condition_note", "ระบุเงื่อนไขอื่น ๆ")} data-field="conditionnote" />
          )}
        </fieldset>

        <label className={field}>
          {label("wht_cert_ui_payer_address", "ที่อยู่ผู้มีหน้าที่หักภาษี (บริษัทเรา)")}
          <Input value={payerAddress} onChange={(e) => setPayerAddress(e.target.value)} data-field="payer.address" />
        </label>
        <label className={field}>
          {label("wht_cert_ui_payee_name", "ชื่อผู้ถูกหักภาษี")}
          <Input value={payeeName} onChange={(e) => setPayeeName(e.target.value)} data-field="payee.name" />
        </label>
        <label className={field}>
          {label("wht_cert_ui_payee_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้ถูกหัก (13 หลัก)")}
          <Input value={payeeTaxId} inputMode="numeric" maxLength={17} onChange={(e) => setPayeeTaxId(e.target.value)} className="font-mono" data-field="payee.taxid" />
        </label>
        <label className={field}>
          {label("wht_cert_ui_payee_address", "ที่อยู่ผู้ถูกหักภาษี")}
          <Input value={payeeAddress} onChange={(e) => setPayeeAddress(e.target.value)} data-field="payee.address" />
        </label>

        <div className="grid grid-cols-3 gap-3">
          <label className={field}>{label("wht_cert_ui_book_no", "เล่มที่")}<Input value={bookNo} onChange={(e) => setBookNo(e.target.value)} /></label>
          <label className={field}>{label("wht_cert_ui_run_no", "เลขที่")}<Input value={runNo} onChange={(e) => setRunNo(e.target.value)} /></label>
          <label className={field}>{label("wht_cert_ui_sequence_no", "ลำดับที่ในใบแนบ")}<Input value={sequenceNo} onChange={(e) => setSequenceNo(e.target.value)} /></label>
        </div>

        <div className="grid gap-2 text-sm">
          <label className="flex min-h-[2.2em] cursor-pointer items-center gap-2">
            <Checkbox checked={archiveCopy} onCheckedChange={setArchiveCopy} />
            {tr("wht_cert_ui_archive_copy", "พิมพ์สำเนาคู่ฉบับ (ฉบับที่ 3 เก็บไว้ที่บริษัท)")}
          </label>
          <label className="flex min-h-[2.2em] cursor-pointer items-center gap-2">
            <Checkbox checked={replacement} onCheckedChange={setReplacement} />
            {tr("wht_cert_ui_replacement", "เป็นใบแทน (กรณีฉบับเดิมชำรุดหรือสูญหาย)")}
          </label>
        </div>

        {error && (
          <div role="alert" className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </div>
        )}

        <Button onClick={generate} disabled={busy} className="min-h-[2.8em] w-full gap-2 text-sm font-bold">
          {busy ? <Loader2 className="size-4 animate-spin" aria-hidden /> : <FileText className="size-4" aria-hidden />}
          {busy ? tr("wht_cert_ui_generating", "กำลังสร้างหนังสือรับรอง...") : tr("wht_cert_ui_generate", "สร้างหนังสือรับรอง (PDF)")}
        </Button>
      </Card>

      <Card className="min-h-[70vh] overflow-hidden p-0 shadow-sm">
        {pdfUrl ? (
          <iframe src={pdfUrl} title={tr("wht_cert_ui_title", "หนังสือรับรองการหักภาษี ณ ที่จ่าย (50 ทวิ)")} className="h-full min-h-[70vh] w-full border-0" />
        ) : (
          <div className="flex h-full min-h-[70vh] items-center justify-center p-8 text-center text-sm leading-relaxed text-muted-foreground">
            {tr("wht_cert_ui_empty", "กด \"สร้างหนังสือรับรอง (PDF)\" เพื่อดูตัวอย่าง พิมพ์ หรือดาวน์โหลดได้จากหน้าต่างนี้")}
          </div>
        )}
      </Card>
    </div>
  );
}
