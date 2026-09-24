"use client";

import { useEffect, useState } from "react";
import { FileText, Loader2, Save } from "lucide-react";
import { useBackendText } from "@/components/backend-text-provider";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { useConfirmDialog, type ConfirmDialogOptions } from "@/components/ui/confirm-dialog";
import { DateField } from "@/components/ui/date-time-field";
import { Input } from "@/components/ui/input";
import { amountUnits, formatAmount, normalizeJournalLines, type GLJournal } from "@/lib/general-ledger";
import { glRequest } from "@/lib/general-ledger-api";
import { journalDetailsProblem, normalizeBranchNo, normalizeJournalDetails, reconciliationChanges, taxDigits, type GLDetailWithholding } from "@/lib/gl-journal-details";
import { useGLCommand } from "@/app/gl/gl-common";
import type { LanguageCode } from "@/lib/i18n";
import {
  requestWhtCertificatePdf,
  whtCertificateForm,
  type CompanyHeader,
  type WhtCertificateInput,
  type WhtCertificateRecordRef,
  type WhtReportRow,
  WHT_INCOME_OPTIONS,
} from "@/lib/thai-tax";

// ใบ 50 ทวิ: จอนี้เตรียมข้อมูลให้นักบัญชีตรวจ/เลือก แล้วให้ backend สร้าง PDF บนแบบฟอร์มกรมสรรพากร
// ยอดเงินมาจากรายงานภาษีหัก ณ ที่จ่าย (backend) อ่านอย่างเดียว — จอไม่คำนวณยอดรวมหรือตัวอักษรเอง
// ฐานภาษีแก้ที่รายละเอียดใบสำคัญ GL (หมวดภาษีหัก ณ ที่จ่าย) เพื่อให้หนังสือรับรองกับแบบยื่นตรงกันเสมอ
// ไม่มีตัวเลือกเงินเดือน (40(1), ภ.ง.ด.1ก) เพราะ BC ไม่ทำระบบเงินเดือน

const FORM_OPTIONS = [
  { value: "53", key: "wht_cert_ui_form_53", th: "ภ.ง.ด.53 (นิติบุคคล)" },
  { value: "3", key: "wht_cert_ui_form_3", th: "ภ.ง.ด.3 (บุคคลธรรมดา)" },
  { value: "2", key: "wht_cert_ui_form_2", th: "ภ.ง.ด.2 (ดอกเบี้ย/เงินปันผล)" },
  { value: "3a", key: "wht_cert_ui_form_3a", th: "ภ.ง.ด.3ก" },
  { value: "2a", key: "wht_cert_ui_form_2a", th: "ภ.ง.ด.2ก" },
];

const INCOME_WITH_NOTE = new Set(["40_4b_1_4", "40_4b_2_5", "other"]);
// ชื่อช่องตาม field ที่ backend (whtcert/render.go) รายงานเมื่อข้อความยาวเกินช่องบนแบบฟอร์ม
const WHT_FIELD_LABELS: Record<string, [string, string]> = {
  bookno: ["wht_cert_ui_book_no", "เล่มที่"],
  runno: ["wht_cert_ui_run_no", "เลขที่"],
  sequenceno: ["wht_cert_ui_sequence_no", "ลำดับที่ในใบแนบ"],
  conditionnote: ["wht_cert_ui_condition_note", "ระบุเงื่อนไขอื่น ๆ"],
  "payer.name": ["wht_cert_ui_payer_name", "ชื่อผู้มีหน้าที่หักภาษี (บริษัทเรา)"],
  "payer.taxid": ["wht_cert_ui_payer_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้มีหน้าที่หักภาษี (13 หลัก)"],
  "payer.address": ["wht_cert_ui_payer_address", "ที่อยู่ผู้มีหน้าที่หักภาษี (บริษัทเรา)"],
  "payee.name": ["wht_cert_ui_payee_name", "ชื่อผู้ถูกหักภาษี"],
  "payee.taxid": ["wht_cert_ui_payee_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้ถูกหัก (13 หลัก)"],
  "payee.address": ["wht_cert_ui_payee_address", "ที่อยู่ผู้ถูกหักภาษี"],
  // ช่องแบบยื่น/ผู้จ่ายเงิน/เงินได้ — backend ชี้ช่องเหล่านี้เมื่อค่าไม่ถูกต้อง (ใบที่อ้างรายการใช้ค่าที่บันทึกเสมอ ไม่ตรวจเทียบค่าจากจอแล้ว)
  form: ["wht_cert_ui_form", "ลำดับที่ในแบบ (แบบยื่นรายการ)"],
  condition: ["wht_cert_ui_condition", "ผู้จ่ายเงิน"],
  incomes: ["wht_cert_ui_income", "ประเภทเงินได้พึงประเมินที่จ่าย"],
  "incomes[0].type": ["wht_cert_ui_income", "ประเภทเงินได้พึงประเมินที่จ่าย"],
  "incomes[0].paiddate": ["wht_cert_ui_paid_date", "วันที่จ่ายเงิน"],
  "incomes[0].amount": ["wht_cert_ui_amount", "จำนวนเงินที่จ่าย"],
  "incomes[0].tax": ["wht_cert_ui_tax", "ภาษีที่หักและนำส่ง"],
};

const CONDITION_OPTIONS = [
  { value: "withhold", key: "wht_cert_ui_condition_withhold", th: "(1) หัก ณ ที่จ่าย" },
  { value: "always", key: "wht_cert_ui_condition_always", th: "(2) ออกให้ตลอดไป" },
  { value: "once", key: "wht_cert_ui_condition_once", th: "(3) ออกให้ครั้งเดียว" },
  { value: "other", key: "wht_cert_ui_condition_other", th: "(4) อื่น ๆ (ระบุ)" },
];

// condition_type ที่บันทึกในใบสำคัญ: 1=หัก ณ ที่จ่าย 2=ออกให้ตลอดไป 3=ออกให้ครั้งเดียว ("" = ไม่ได้บันทึก)
const conditionValue = (recorded: number) => (recorded === 1 ? "withhold" : recorded === 2 ? "always" : recorded === 3 ? "once" : "");

// ฐานภาษีมาจากไหน — ใช้ทั้งจอออก 50 ทวิ และจอดูใบที่ได้รับ
const taxBaseHint = (row: WhtReportRow, tr: (key: string, fallback: string) => string) =>
  row.taxbasesource === "recorded"
    ? tr("wht_cert_ui_base_recorded", "ฐานภาษีตามที่บันทึกในใบสำคัญ — แก้ได้ที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย")
    : tr("wht_cert_ui_base_inferred", "ฐานภาษีนี้ระบบประมาณจากบรรทัดบัญชี กรุณาตรวจ — บันทึกฐานภาษีจริงได้ที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย");

// disabled (ค่าตามรายการที่บันทึก) ยังต้องอ่านชัด — ห้ามจางตาม opacity ของเบราว์เซอร์ (คน 40+)
const selectClass =
  "w-full min-h-[2.6em] rounded-lg border border-border bg-background px-3 text-sm font-medium text-foreground shadow-sm focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20 disabled:cursor-not-allowed disabled:bg-muted/40 disabled:opacity-100";

const today = () => {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
};

/** ข้อมูลผู้จ่าย/ผู้รับเงิน เล่มที่ และหมายเหตุ ที่เก็บเป็น snapshot ในรายการภาษีหักของใบสำคัญ (wht.sql) */
type WhtSnapshotField = "payer_name" | "payer_tax_id" | "payer_branch_no" | "payer_address" | "payee_name" | "payee_tax_id" | "payee_branch_no" | "payee_address" | "wht_book_no" | "remark";
export type WhtSnapshot = Record<WhtSnapshotField, string>;
const WHT_SNAPSHOT_LABELS: Record<WhtSnapshotField, [string, string]> = {
  payer_name: ["wht_cert_ui_payer_name", "ชื่อผู้มีหน้าที่หักภาษี (บริษัทเรา)"],
  payer_tax_id: ["wht_cert_ui_payer_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้มีหน้าที่หักภาษี (13 หลัก)"],
  payer_branch_no: ["wht_cert_ui_payer_branch", "สาขาของผู้มีหน้าที่หักภาษี (5 หลัก, 00000 = สำนักงานใหญ่)"],
  payer_address: ["wht_cert_ui_payer_address", "ที่อยู่ผู้มีหน้าที่หักภาษี (บริษัทเรา)"],
  payee_name: ["wht_cert_ui_payee_name", "ชื่อผู้ถูกหักภาษี"],
  payee_tax_id: ["wht_cert_ui_payee_taxid", "เลขประจำตัวผู้เสียภาษีอากรผู้ถูกหัก (13 หลัก)"],
  payee_branch_no: ["wht_cert_ui_payee_branch", "สาขาของผู้ถูกหักภาษี (5 หลัก, 00000 = สำนักงานใหญ่)"],
  payee_address: ["wht_cert_ui_payee_address", "ที่อยู่ผู้ถูกหักภาษี"],
  wht_book_no: ["wht_cert_ui_book_no", "เล่มที่"],
  remark: ["wht_cert_ui_remark", "หมายเหตุ (ไม่เกิน 500 ตัวอักษร)"],
};
const WHT_SNAPSHOT_FIELDS = Object.keys(WHT_SNAPSHOT_LABELS) as WhtSnapshotField[];

/**
 * หารายการภาษีหักในใบสำคัญที่ตรงกับแถวรายงาน: มี withholdingid = ตรงตัว, ไม่มี = เทียบ เราหักภาษี (ทิศทาง 1), คู่ค้า, ฐาน, ภาษี,
 * วันที่จ่าย, ประเภทเงินได้ และเลขที่หนังสือรับรอง (เมื่อมีทั้งสองฝั่ง) — เจอมากกว่า 1 รายการ = ไม่เดา ให้แก้ที่ใบสำคัญแทน
 */
export function findRecordedWithholding(journal: Pick<GLJournal, "details"> | null, row: WhtReportRow): { item: GLDetailWithholding } | { problem: "not_found" | "ambiguous" } {
  const items = journal?.details?.withholdings ?? [];
  // แถวที่บันทึกมี id รายการจากรายงาน → จับคู่ตรงตัว; แถวเก่าที่ไม่มี id ค่อยเทียบจากคู่ค้า/ยอด
  if (row.withholdingid) {
    const byId = items.filter((item) => item.id === row.withholdingid);
    return byId.length === 1 ? { item: byId[0] } : { problem: "not_found" };
  }
  const sameAmount = (a?: string, b?: string) => { try { return amountUnits(a || "0") === amountUnits(b || "0"); } catch { return false; } };
  const matches = items.filter((item) =>
    Number(item.wht_direction) === 1
    && item.partner_code === row.partnercode
    && sameAmount(item.base_amount, row.baseamount)
    && (!item.tax_amount || sameAmount(item.tax_amount, row.whtamount))
    && (!row.paiddate || item.payment_date === row.paiddate)
    && (!row.incometype || item.income_tax_type === row.incometype)
    && (!row.certificateno || !item.wht_cert_no || item.wht_cert_no === row.certificateno));
  if (matches.length === 1) return { item: matches[0] };
  return { problem: matches.length > 1 ? "ambiguous" : "not_found" };
}

/**
 * ค่าที่แสดงบนจอ: snapshot ที่บันทึกก่อน → ทะเบียนบริษัท/คู่ค้าจากรายงาน; ที่อยู่ผู้หักภาษี (บริษัทเรา) ที่ใบยังไม่มี
 * ใช้ที่อยู่สำหรับภาษีในทะเบียนบริษัท (addressline จาก backend) เป็นค่าตั้งต้นที่ยังแก้ได้ — ทะเบียนว่าง = ว่างให้กรอก
 * ค่าตั้งต้นจากทะเบียนเป็นค่าแสดง/พิมพ์เท่านั้น ไม่ถูกบันทึกลงใบสำคัญจนกว่าผู้ใช้แก้ช่องนั้นเอง (whtSnapshotChanges)
 */
export function whtSnapshotView(item: GLDetailWithholding | undefined, row: WhtReportRow, company: CompanyHeader | null): WhtSnapshot {
  return {
    payer_name: item?.payer_name || company?.name || "",
    payer_tax_id: item?.payer_tax_id || company?.taxid || "",
    payer_branch_no: item?.payer_branch_no ?? "",
    payer_address: item?.payer_address || company?.addressline || "",
    payee_name: item?.payee_name || row.partnerfullname || row.partnername || "",
    payee_tax_id: item?.payee_tax_id || row.taxid || "",
    payee_branch_no: item?.payee_branch_no ?? "",
    payee_address: item?.payee_address || row.address || "",
    wht_book_no: item?.wht_book_no ?? "",
    remark: item?.remark ?? "",
  };
}

const normalizeSnapshotValue = (field: WhtSnapshotField, value: string | undefined) =>
  field.endsWith("_tax_id") ? taxDigits(value) : field.endsWith("_branch_no") ? normalizeBranchNo(value) : (value ?? "").trim();

/**
 * ช่องที่จะบันทึกลงรายการภาษีหัก: เฉพาะช่องที่ผู้ใช้แก้บนจอ (ไม่ล็อก) และค่าใหม่ต่างจากค่าที่บันทึกในใบสำคัญจริง
 * before = ค่าในใบสำคัญ (ไม่ใช่ค่าตั้งต้นบนจอ) — หน้ายืนยันแสดงทุกช่องที่จะถูกเขียน ค่าตั้งต้นจากทะเบียนที่ผู้ใช้ไม่ได้แตะไม่ถูกบันทึกเงียบ ๆ
 * (review 2026-09-25: เดิมส่งทุกช่อง ที่อยู่บริษัทจากทะเบียนถูกแช่ลงใบที่ผ่านบัญชีโดยหน้ายืนยันไม่แสดง)
 */
export function whtSnapshotChanges(
  recorded: GLDetailWithholding | undefined,
  snapshot: WhtSnapshot,
  shown: WhtSnapshot,
  locked: Partial<Record<WhtSnapshotField, boolean>> = {},
): { field: WhtSnapshotField; before: string; after: string }[] {
  return WHT_SNAPSHOT_FIELDS
    .filter((field) => !locked[field] && snapshot[field].trim() !== shown[field].trim())
    .map((field) => ({ field, before: normalizeSnapshotValue(field, recorded?.[field]), after: normalizeSnapshotValue(field, snapshot[field]) }))
    .filter((change) => change.before !== change.after);
}

/**
 * รายการที่บันทึกในใบสำคัญที่ใบ 50 ทวิ อ้างถึง: แถวที่บันทึก + รู้ id รายการ (จากรายงาน หรือจับคู่ได้หลังโหลดใบสำคัญ)
 * มีค่า = backend พิมพ์ยอด ภาษี แบบยื่น เงื่อนไข ประเภทเงินได้ วันที่จ่าย และผู้รับเงินตามรายการ (ค่าบนจอที่ไม่ตรง = ปฏิเสธ)
 */
export function whtRecordRef(row: WhtReportRow, recorded?: Pick<GLDetailWithholding, "id">): WhtCertificateRecordRef | null {
  if (row.taxbasesource !== "recorded" || !row.journalid) return null;
  const withholdingid = row.withholdingid || recorded?.id || "";
  return withholdingid ? { journalid: row.journalid, withholdingid } : null;
}

/**
 * ค่าของใบ 50 ทวิ ที่ส่งให้ backend: อ้างรายการที่บันทึก = ช่องตัวเลขส่งว่าง ให้ backend เติมจากรายการเอง
 * (backend พิมพ์ยอด ภาษี วันที่ ประเภทเงินได้ แบบยื่น และเงื่อนไขจากรายการที่บันทึกเสมอ — ค่าจากจอใช้เฉพาะใบที่ไม่อ้างรายการ)
 * ยกเว้นประเภทเงินได้: ส่งประเภทที่บันทึก (values.incomeType = ค่าที่แสดง) คู่กับข้อความ "ระบุ" เพราะ backend ใช้ข้อความจากจอ
 * เฉพาะเมื่อประเภทตรงกับรายการ (applyRecordedFigures) — ส่งว่างแล้วข้อความที่ผู้ใช้พิมพ์หายเงียบ ๆ (review 2026-09-24)
 */
export function whtCertificateFigures(
  locked: boolean,
  values: { form: string; condition: string; conditionNote: string; incomeType: string; incomeNote: string; paidDate: string },
  row: Pick<WhtReportRow, "baseamount" | "whtamount">,
): Pick<WhtCertificateInput, "form" | "condition" | "conditionnote" | "incomes"> {
  const note = INCOME_WITH_NOTE.has(values.incomeType) ? values.incomeNote.trim() : "";
  if (locked) {
    return { form: "", condition: "", conditionnote: "", incomes: [{ type: values.incomeType, paiddate: "", amount: "", tax: "", note }] };
  }
  return {
    form: values.form,
    condition: values.condition,
    conditionnote: values.condition === "other" ? values.conditionNote.trim() : "",
    incomes: [{ type: values.incomeType, paiddate: values.paidDate, amount: row.baseamount, tax: row.whtamount, note }],
  };
}

/**
 * วันที่ออกหนังสือรับรอง: รายการที่บันทึกมีวันที่ออกแล้ว = backend พิมพ์วันที่นั้นเสมอ (applyRecordedFigures)
 * → แสดงอ่านอย่างเดียวและส่งค่านั้น ไม่ปล่อยให้ผู้ใช้แก้แล้วถูกเมินเงียบ ๆ (review 2026-09-24)
 */
export function whtCertificateIssueDate(locked: boolean, recordedDate: string | undefined, screenDate: string): { value: string; fromRecord: boolean } {
  const recordedValue = locked ? (recordedDate ?? "").trim() : "";
  return recordedValue ? { value: recordedValue, fromRecord: true } : { value: screenDate, fromRecord: false };
}

/** ถามก่อนทิ้งข้อมูลผู้จ่าย/ผู้รับเงินที่แก้แต่ยังไม่บันทึกลงใบสำคัญ (เปลี่ยนแท็บ/งวดแล้วแผงถูกถอด ค่าที่แก้หาย) — ไม่มีค่าค้าง = ไปต่อได้ทันที */
export async function leaveWhtCertificate(dirty: boolean, confirm: (options: ConfirmDialogOptions) => Promise<boolean>, tr: (key: string, fallback: string) => string): Promise<boolean> {
  if (!dirty) return true;
  return confirm({
    title: tr("gl_discard_unsaved_data", "ละทิ้งข้อมูลที่ยังไม่บันทึก?"),
    description: tr("wht_cert_ui_unsaved_leave", "ข้อมูลผู้จ่ายเงิน/ผู้รับเงินที่แก้ในหนังสือรับรองยังไม่ได้บันทึกลงใบสำคัญ — กด “ยกเลิก” แล้วกด “บันทึกลงใบสำคัญ” ก่อน หรือกด “ละทิ้งการแก้ไข” เพื่อไปต่อโดยไม่บันทึก"),
    confirmLabel: tr("gl_discard_changes", "ละทิ้งการแก้ไข"),
    cancelLabel: tr("common_cancel", "ยกเลิก"),
    tone: "warning",
  });
}

type Props = {
  row: WhtReportRow;
  company: CompanyHeader | null;
  holdingcode: string;
  businesscode: string;
  language: LanguageCode;
  /** แจ้งผู้เรียกว่ามีข้อมูลผู้จ่าย/ผู้รับเงินที่แก้แต่ยังไม่บันทึกลงใบสำคัญ — ใช้ถามก่อนเปลี่ยนแท็บหรืองวด */
  onDirtyChange?: (dirty: boolean) => void;
};

export function WhtCertificatePanel({ row, company, holdingcode, businesscode, language, onDirtyChange }: Props) {
  const tr = useBackendText();
  // แบบยื่นตามที่บันทึกในใบสำคัญ (PND3 → ภ.ง.ด.3 ฯลฯ) — ไม่รู้แบบ (ฐานประมาณ) = ยังไม่เลือก ให้ผู้ใช้เลือกเอง
  // (เดิมตั้ง ภ.ง.ด.53 ให้ ผู้รับเงินบุคคลธรรมดาได้ใบที่ติ๊กแบบผิดโดยไม่รู้ตัว — review 2026-09-24; ส่งว่าง backend ตอบ wht_cert_form_invalid ช่อง form)
  const [form, setForm] = useState(() => whtCertificateForm(row.formtype));
  const recordedIncome = WHT_INCOME_OPTIONS.some((o) => o.value === row.incometype) ? row.incometype : "";
  const [incomeType, setIncomeType] = useState(recordedIncome || "3_tres");
  const [incomeNote, setIncomeNote] = useState("");
  const [condition, setCondition] = useState(conditionValue(row.condition) || "withhold");
  const [conditionNote, setConditionNote] = useState("");
  const [paidDate, setPaidDate] = useState(row.paiddate || row.docdate);
  const [issueDate, setIssueDate] = useState(today);
  const [runNo, setRunNo] = useState(row.certificateno || row.docno);
  const [sequenceNo, setSequenceNo] = useState("");
  // ผู้จ่าย/ผู้รับเงิน เล่มที่ หมายเหตุ: ค่าเริ่มจาก snapshot ในใบสำคัญ (โหลดด้านล่าง) — ไม่เก็บในเครื่อง (localStorage)
  const [snapshot, setSnapshot] = useState<WhtSnapshot>(() => whtSnapshotView(undefined, row, company));
  const [shown, setShown] = useState<WhtSnapshot>(() => whtSnapshotView(undefined, row, company));
  const [journal, setJournal] = useState<GLJournal | null>(null);
  const [recordState, setRecordState] = useState<"inferred" | "loading" | "ready" | "not_found" | "ambiguous" | "failed">(row.taxbasesource === "recorded" && row.journalid ? "loading" : "inferred");
  const [saveError, setSaveError] = useState("");
  const [saveMessage, setSaveMessage] = useState("");
  const { busy: saving, execute } = useGLCommand();
  const { confirm, confirmationDialog } = useConfirmDialog({ defaultConfirmLabel: tr("common_confirm", "ยืนยัน"), defaultCancelLabel: tr("common_cancel", "ยกเลิก") });
  const match = findRecordedWithholding(journal, row);
  const recorded = "item" in match ? match.item : undefined;
  // อ้างรายการที่บันทึก → แบบยื่น/ประเภทเงินได้/วันที่จ่าย/เงื่อนไข อ่านอย่างเดียวตามรายการ (backend ปฏิเสธค่าที่ไม่ตรง)
  const recordRef = whtRecordRef(row, recorded);
  const figuresLocked = recordRef !== null;
  const shownForm = figuresLocked ? whtCertificateForm(recorded?.form_type ?? row.formtype) || form : form;
  const shownIncome = figuresLocked ? recorded?.income_tax_type ?? row.incometype : incomeType;
  const shownPaidDate = figuresLocked ? recorded?.payment_date ?? row.paiddate : paidDate;
  const shownCondition = figuresLocked ? conditionValue(recorded?.condition_type ?? row.condition) : condition;
  const issue = whtCertificateIssueDate(figuresLocked, recorded?.certificate_date, issueDate);
  // เลขที่หนังสือรับรองที่บันทึกในรายการ backend พิมพ์ค่านั้นเสมอ — แก้ได้เฉพาะเมื่อรายการยังไม่มีเลขที่
  const runNoLocked = figuresLocked && Boolean(recorded?.wht_cert_no ?? row.certificateno);
  const setSnap = (field: keyof WhtSnapshot, value: string) => { setSnapshot((current) => ({ ...current, [field]: value })); setSaveMessage(""); };
  const shownRunNo = runNoLocked ? recorded?.wht_cert_no || row.certificateno : runNo;
  // อ้างรายการที่บันทึกแต่ยังบันทึกลงใบสำคัญไม่ได้ (กำลังโหลด/โหลดไม่สำเร็จ/ไม่พบ/กลับรายการแล้ว) = ช่องผู้จ่าย/ผู้รับเงินอ่านอย่างเดียว
  // เพราะ backend พิมพ์ค่าที่บันทึก ค่าที่แก้บนจอจะไม่ถูกใช้และบันทึกไม่ได้ — ไม่ให้ผู้ใช้แก้แล้วติดทางตัน
  const snapshotEditable = !figuresLocked || (recordState === "ready" && journal?.status !== "reversed");
  // ชื่อ/เลขผู้หักภาษี: ทะเบียนบริษัทมีค่า = backend พิมพ์ตามทะเบียน จึงแก้บนจอนี้ไม่ได้ (แก้ที่ข้อมูลบริษัท)
  const payerNameLocked = Boolean(company?.name?.trim());
  const payerTaxIdLocked = Boolean(company?.taxid?.trim());
  const [archiveCopy, setArchiveCopy] = useState(false);
  const [replacement, setReplacement] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [pdfUrl, setPdfUrl] = useState("");

  useEffect(() => {
    if (row.taxbasesource !== "recorded" || !row.journalid) return;
    const controller = new AbortController();
    glRequest<GLJournal>(`journals/${encodeURIComponent(row.journalid)}`, { signal: controller.signal })
      .then((loaded) => {
        if (controller.signal.aborted) return;
        const found = findRecordedWithholding(loaded, row);
        const view = whtSnapshotView("item" in found ? found.item : undefined, row, company);
        setJournal(loaded);
        setSnapshot(view);
        setShown(view);
        // ข้อความ "ระบุ" ของประเภทเงินได้: backend ใช้คำอธิบายเงินได้ที่บันทึกเมื่อช่องว่าง — แสดงค่านั้นให้เห็นก่อนพิมพ์
        if ("item" in found) { const recordedNote = (found.item.income_description ?? "").trim(); setIncomeNote((current) => current || recordedNote); }
        setRecordState("item" in found ? "ready" : found.problem);
      })
      .catch(() => { if (!controller.signal.aborted) setRecordState("failed"); });
    return () => controller.abort();
  }, [row, company]);
  useEffect(() => () => { if (pdfUrl) URL.revokeObjectURL(pdfUrl); }, [pdfUrl]);

  const generate = async () => {
    // อ้างรายการที่บันทึก: backend พิมพ์ผู้จ่าย/ผู้รับเงินตามค่าที่บันทึกในใบสำคัญ — ค่าที่แก้ค้างบนจอจะไม่ถูกพิมพ์ จึงต้องบันทึกก่อน
    if (recordRef && dirty) {
      setPdfUrl("");
      setError(tr("wht_cert_ui_save_before_pdf", "มีข้อมูลผู้จ่าย/ผู้รับเงินที่แก้แต่ยังไม่ได้บันทึก — หนังสือรับรองพิมพ์ตามค่าที่บันทึกในใบสำคัญ กรุณากด “บันทึกลงใบสำคัญ” ก่อน แล้วจึงกดสร้าง PDF"));
      return;
    }
    setBusy(true);
    setError("");
    const result = await requestWhtCertificatePdf(holdingcode, businesscode, {
      bookno: snapshot.wht_book_no.trim(),
      runno: (runNoLocked ? shownRunNo : runNo).trim(),
      sequenceno: sequenceNo.trim(),
      issuedate: issue.value,
      archivecopy: archiveCopy,
      replacement,
      // ชื่อ/เลขผู้หักภาษี backend ยึดจากทะเบียนบริษัท ค่าที่ส่งเป็นค่าสำรองเมื่อทะเบียนยังว่าง
      payer: { name: snapshot.payer_name.trim(), address: snapshot.payer_address.trim(), taxid: taxDigits(snapshot.payer_tax_id) },
      payee: { name: snapshot.payee_name.trim(), address: snapshot.payee_address.trim(), taxid: taxDigits(snapshot.payee_tax_id) },
      ...whtCertificateFigures(figuresLocked, { form, condition, conditionNote, incomeType: shownIncome, incomeNote, paidDate }, row),
    }, { record: recordRef, language });
    setBusy(false);
    if (result.ok) {
      setPdfUrl(URL.createObjectURL(result.pdf));
      return;
    }
    setPdfUrl("");
    // code จาก backend เป็น language key (wht_cert_*) → แปลตามภาษาบนจอ ไม่ยึด Accept-Language ของเบราว์เซอร์
    const fallback = tr("wht_cert_render_failed", "สร้างหนังสือรับรอง 50 ทวิ ไม่สำเร็จ กรุณาลองใหม่อีกครั้ง");
    // user_access_expired (403) บอกให้ติดต่อผู้ดูแล — ไม่ใช่ "ลองใหม่" (review 2026-09-24)
    const translatable = result.error.startsWith("wht_cert_") || result.error === "user_access_expired";
    const message = translatable ? tr(result.error, result.message || fallback) : result.message || fallback;
    // backend บอกช่องที่ข้อความยาวเกิน (field) — บอกชื่อช่องให้ผู้ใช้รู้ว่าต้องย่อตรงไหน
    const box = result.field ? WHT_FIELD_LABELS[result.field] : undefined;
    setError(box ? `${message} — ${tr("wht_cert_ui_error_field", "ช่องที่ต้องแก้")}: ${tr(box[0], box[1])}` : message);
  };

  const changedFields = WHT_SNAPSHOT_FIELDS.filter((field) => snapshot[field].trim() !== shown[field].trim());
  const canSaveSnapshot = recordState === "ready" && journal?.status !== "reversed" && changedFields.length > 0 && !saving;
  // Dirty guard (AGENTS.md 40+ rule 7): unsaved payer/payee edits are reported to the workbench,
  // which asks before switching tab or period; closing/reloading the browser asks too.
  const dirty = changedFields.length > 0;
  useEffect(() => { onDirtyChange?.(dirty); }, [dirty, onDirtyChange]);
  useEffect(() => () => onDirtyChange?.(false), [onDirtyChange]);
  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => { event.preventDefault(); event.returnValue = ""; };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);

  const saveSnapshot = async () => {
    if (!journal?.id || !recorded || !canSaveSnapshot) return;
    setSaveError("");
    setSaveMessage("");
    // ชื่อ/เลขผู้หักภาษีที่ล็อกตามทะเบียนคงค่าที่บันทึกไว้เดิม; ช่องที่ไม่ได้แก้คงค่าในใบสำคัญ (ไม่แช่ค่าตั้งต้นจากทะเบียน)
    const changes = whtSnapshotChanges(recorded, snapshot, shown, { payer_name: payerNameLocked, payer_tax_id: payerTaxIdLocked });
    // แก้แล้วแต่ค่าตรงกับที่บันทึกอยู่แล้ว (เช่น ล้างค่าตั้งต้นที่ไม่เคยถูกบันทึก) = ไม่มีอะไรต้องเขียน
    if (!changes.length) { setShown(snapshot); setSaveMessage(tr("wht_cert_ui_snapshot_saved", "บันทึกข้อมูลลงใบสำคัญแล้ว")); return; }
    const updated: GLDetailWithholding = { ...recorded, ...Object.fromEntries(changes.map((change) => [change.field, change.after])) };
    const details = { ...journal.details, withholdings: (journal.details?.withholdings ?? []).map((item) => (item.id === recorded.id ? updated : item)) };
    const problem = journalDetailsProblem(details, tr);
    if (problem) { setSaveError(problem.message); return; }
    const notSpecified = tr("tax_not_specified", "ยังไม่ระบุ");
    if (!await confirm({
      title: tr("wht_cert_ui_snapshot_confirm_title", "บันทึกการแก้ไขลงใบสำคัญ {0}?").replace("{0}", journal.docno),
      description: tr("wht_cert_ui_snapshot_confirm_hint", "ค่าที่บันทึกไว้เดิมจะถูกแทนด้วยค่าใหม่ (ระบบเก็บค่าเดิมไว้ในประวัติ)"),
      details: (
        <ul className="grid gap-1 text-[0.9rem] leading-snug">
          {changes.map(({ field, before, after }) => (
            <li key={field} className="[overflow-wrap:anywhere]">
              <span className="font-semibold">{tr(...WHT_SNAPSHOT_LABELS[field])}</span>: {before || notSpecified} → {after || notSpecified}
            </li>
          ))}
        </ul>
      ),
      confirmLabel: tr("wht_cert_ui_save_snapshot", "บันทึกลงใบสำคัญ"),
      tone: "warning",
    })) return;
    const reason = tr("wht_cert_ui_snapshot_reason", "แก้ข้อมูลผู้จ่าย/ผู้รับเงินของหนังสือรับรอง 50 ทวิ");
    const normalized = normalizeJournalDetails(details);
    try {
      // ผ่านบัญชีแล้ว = กระทบยอดรายละเอียด (ยอด GL ไม่เปลี่ยน เก็บค่าเดิมใน audit); ฉบับร่าง = บันทึกใบสำคัญทั้งใบ
      await execute(journal.status === "posted"
        ? { resource: "journals", action: "reconcile", id: journal.id, version: journal.version, reason, journal: { details: reconciliationChanges(journal.details, normalized) } }
        : { resource: "journals", action: "update", id: journal.id, version: journal.version, reason, journal: { ...journal, lines: normalizeJournalLines(journal.lines), details: normalized } });
      const fresh = await glRequest<GLJournal>(`journals/${encodeURIComponent(journal.id)}`);
      const found = findRecordedWithholding(fresh, row);
      const view = whtSnapshotView("item" in found ? found.item : undefined, row, company);
      setJournal(fresh);
      setSnapshot(view);
      setShown(view);
      setRecordState("item" in found ? "ready" : found.problem);
      setSaveMessage(tr("wht_cert_ui_snapshot_saved", "บันทึกข้อมูลลงใบสำคัญแล้ว"));
    } catch (cause) {
      // ข้อความไทยจาก backend (เช่น เลขภาษีผิดช่องไหน) แสดงตรง ๆ
      setSaveError(cause instanceof Error && cause.message ? cause.message : tr("wht_cert_ui_snapshot_failed", "บันทึกลงใบสำคัญไม่สำเร็จ กรุณาลองใหม่"));
    }
  };

  const recordNotice = recordState === "loading"
    ? tr("wht_cert_ui_record_loading", "กำลังโหลดข้อมูลที่บันทึกในใบสำคัญ...")
    : recordState === "inferred"
      ? tr("wht_cert_ui_record_inferred", "ใบสำคัญนี้ยังไม่ได้บันทึกรายการภาษีหัก — บันทึกที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย ก่อน จึงจะเก็บข้อมูลผู้จ่าย/ผู้รับเงินลงใบสำคัญได้")
      : recordState === "not_found"
        ? tr("wht_cert_ui_record_not_found", "ไม่พบรายการภาษีหักนี้ในใบสำคัญ {0} — แก้ข้อมูลที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย").replace("{0}", row.docno)
        : recordState === "ambiguous"
          ? tr("wht_cert_ui_record_ambiguous", "ใบสำคัญ {0} มีรายการภาษีหักที่ตรงกันมากกว่า 1 รายการ — แก้ข้อมูลที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย แทน").replace("{0}", row.docno)
          : recordState === "failed"
            ? tr("wht_cert_ui_record_failed", "โหลดใบสำคัญ {0} ไม่สำเร็จ — ยังสร้าง PDF ได้ แต่บันทึกลงใบสำคัญไม่ได้ กรุณาโหลดหน้าใหม่").replace("{0}", row.docno)
            : journal?.status === "reversed"
              ? tr("wht_cert_ui_record_reversed", "ใบสำคัญนี้กลับรายการแล้ว แก้ข้อมูลที่บันทึกไม่ได้")
              : "";
  const snapInput = (field: keyof WhtSnapshot, dataField: string, options: { locked?: boolean; mono?: boolean; kind?: "taxid" | "branch" } = {}) => {
    const value = snapshot[field];
    const digits = options.kind ? taxDigits(value) : "";
    const hint = options.kind === "taxid" && digits && digits.length !== 13
      ? tr("gl_detail_tax_id_hint", "ต้องเป็นตัวเลข 13 หลัก (ตอนนี้ {0} หลัก) — วางแบบมีขีดได้ ระบบตัดขีดให้เอง").replace("{0}", String(digits.length))
      : options.kind === "branch" && digits.length > 5
        ? tr("gl_detail_branch_hint", "ใส่ได้ไม่เกิน 5 หลัก เช่น 0 หรือ 00000 = สำนักงานใหญ่")
        : field === "remark" && [...value].length > 500
          ? tr("wht_cert_ui_remark_too_long", "หมายเหตุยาวเกิน 500 ตัวอักษร (ตอนนี้ {0}) — กรุณาย่อข้อความ").replace("{0}", String([...value].length))
          : "";
    return (
      <>
        <Input
          value={value}
          readOnly={options.locked || !snapshotEditable}
          disabled={recordState === "loading"}
          inputMode={options.kind ? "numeric" : undefined}
          aria-invalid={hint ? true : undefined}
          className={`${options.mono ? "font-mono" : ""} ${hint ? "border-destructive" : ""}`}
          onChange={(e) => setSnap(field, options.kind ? taxDigits(e.target.value) : e.target.value)}
          onBlur={options.kind === "branch" ? (e) => { const padded = normalizeBranchNo(e.target.value); if (padded !== value) setSnap(field, padded); } : undefined}
          data-field={dataField}
        />
        {hint && <span role="alert" className="text-[0.9rem] leading-snug text-destructive">{hint}</span>}
      </>
    );
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
          <p className="col-span-2 leading-relaxed text-muted-foreground" data-field="taxbasesource">
            {taxBaseHint(row, tr)}
          </p>
        </div>

        {figuresLocked && (
          <p role="note" className="rounded-lg border border-border bg-muted/40 px-3 py-2 text-[0.9rem] leading-relaxed text-foreground" data-field="wht-figures-recorded">
            {tr("wht_cert_ui_figures_recorded", "แบบยื่น ประเภทเงินได้ วันที่จ่าย และผู้จ่ายเงิน ใช้ตามรายการภาษีหักที่บันทึกในใบสำคัญ {0} (ตรงกับแบบ ภ.ง.ด. ที่ยื่น) — ถ้าไม่ถูกต้องให้แก้ที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย").replace("{0}", row.docno)}
          </p>
        )}

        <label className={field}>
          {label("wht_cert_ui_form", "ลำดับที่ในแบบ (แบบยื่นรายการ)")}
          <select className={selectClass} value={shownForm} disabled={figuresLocked} onChange={(e) => setForm(e.target.value)} data-field="form">
            {!shownForm && <option value="" disabled>{tr("wht_cert_ui_form_choose", "— เลือกแบบยื่น (ภ.ง.ด.3 บุคคลธรรมดา / ภ.ง.ด.53 นิติบุคคล) —")}</option>}
            {FORM_OPTIONS.map((o) => <option key={o.value} value={o.value}>{tr(o.key, o.th)}</option>)}
          </select>
        </label>

        <label className={field}>
          {label("wht_cert_ui_income", "ประเภทเงินได้พึงประเมินที่จ่าย")}
          <select className={selectClass} value={shownIncome} disabled={figuresLocked} onChange={(e) => setIncomeType(e.target.value)} data-field="incomes[0].type">
            {WHT_INCOME_OPTIONS.map((o) => <option key={o.value} value={o.value}>{tr(o.key, o.th)}</option>)}
          </select>
        </label>
        {INCOME_WITH_NOTE.has(shownIncome) && (
          <label className={field}>
            {label("wht_cert_ui_income_note", shownIncome === "40_4b_1_4" ? "อัตราภาษีของกำไรสุทธิ (ร้อยละ)" : "ระบุประเภทเงินได้")}
            <Input value={incomeNote} onChange={(e) => setIncomeNote(e.target.value)} data-field="incomes[0].note" />
          </label>
        )}

        <div className="grid grid-cols-2 gap-3">
          <DateField label={tr("wht_cert_ui_paid_date", "วันที่จ่ายเงิน")} language={language} yearType="buddhist" value={shownPaidDate} readOnly={figuresLocked} onChange={(e) => setPaidDate(e.target.value)} data-field="incomes[0].paiddate" />
          <DateField label={tr("wht_cert_ui_issue_date", "วันที่ออกหนังสือรับรอง")} language={language} yearType="buddhist" value={issue.value} readOnly={issue.fromRecord} onChange={(e) => setIssueDate(e.target.value)} data-field="issuedate" />
        </div>
        {issue.fromRecord && (
          <p className="-mt-2 text-[0.9rem] leading-relaxed text-muted-foreground" data-field="issuedate-recorded">{tr("wht_cert_ui_issue_date_recorded", "วันที่ออกหนังสือรับรองใช้ตามที่บันทึกในใบสำคัญ — ถ้าไม่ถูกต้องให้แก้ที่รายละเอียดใบสำคัญ หมวดภาษีหัก ณ ที่จ่าย")}</p>
        )}

        <fieldset className="grid gap-2" disabled={figuresLocked} data-field="condition">
          <legend className="mb-1.5 text-[0.9rem] font-semibold text-foreground">{tr("wht_cert_ui_condition", "ผู้จ่ายเงิน")}</legend>
          <div className="grid grid-cols-2 gap-2">
            {CONDITION_OPTIONS.map((o) => (
              <label key={o.value} className={`flex min-h-[2.6em] items-center gap-2 rounded-lg border px-3 text-sm ${figuresLocked ? "cursor-not-allowed" : "cursor-pointer"} ${shownCondition === o.value ? "border-primary bg-primary/10 font-semibold text-foreground" : "border-border text-muted-foreground"}`}>
                <input type="radio" name="wht-condition" value={o.value} checked={shownCondition === o.value} onChange={() => setCondition(o.value)} className="accent-[var(--primary)]" />
                {tr(o.key, o.th)}
              </label>
            ))}
          </div>
          {!figuresLocked && condition === "other" && (
            <Input value={conditionNote} onChange={(e) => setConditionNote(e.target.value)} placeholder={tr("wht_cert_ui_condition_note", "ระบุเงื่อนไขอื่น ๆ")} data-field="conditionnote" />
          )}
        </fieldset>

        {recordNotice && <p role="status" className="rounded-lg border border-border bg-muted/40 px-3 py-2 text-[0.9rem] leading-relaxed text-foreground" data-field="wht-record-state">{recordNotice}</p>}

        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payer_name)}
          {snapInput("payer_name", "payer.name", { locked: payerNameLocked })}
        </label>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payer_tax_id)}
          {snapInput("payer_tax_id", "payer.taxid", { locked: payerTaxIdLocked, mono: true, kind: "taxid" })}
        </label>
        {(payerNameLocked || payerTaxIdLocked) && (
          <p className="-mt-2 text-[0.9rem] leading-relaxed text-muted-foreground">{tr("wht_cert_ui_payer_from_registry", "ชื่อและเลขผู้เสียภาษีของผู้หักภาษีพิมพ์ตามทะเบียนบริษัท — ถ้าไม่ถูกต้องให้แก้ที่ข้อมูลบริษัท")}</p>
        )}
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payer_branch_no)}
          {snapInput("payer_branch_no", "payer.branch", { mono: true, kind: "branch" })}
        </label>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payer_address)}
          {snapInput("payer_address", "payer.address")}
        </label>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payee_name)}
          {snapInput("payee_name", "payee.name")}
        </label>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payee_tax_id)}
          {snapInput("payee_tax_id", "payee.taxid", { mono: true, kind: "taxid" })}
        </label>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payee_branch_no)}
          {snapInput("payee_branch_no", "payee.branch", { mono: true, kind: "branch" })}
        </label>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.payee_address)}
          {snapInput("payee_address", "payee.address")}
        </label>

        <div className="grid grid-cols-3 gap-3">
          <label className={field}>{label(...WHT_SNAPSHOT_LABELS.wht_book_no)}{snapInput("wht_book_no", "bookno")}</label>
          <label className={field}>{label("wht_cert_ui_run_no", "เลขที่")}<Input value={shownRunNo} readOnly={runNoLocked} onChange={(e) => setRunNo(e.target.value)} data-field="runno" /></label>
          <label className={field}>{label("wht_cert_ui_sequence_no", "ลำดับที่ในใบแนบ")}<Input value={sequenceNo} onChange={(e) => setSequenceNo(e.target.value)} /></label>
        </div>
        <label className={field}>
          {label(...WHT_SNAPSHOT_LABELS.remark)}
          {snapInput("remark", "remark")}
        </label>

        {saveError && (
          <div role="alert" className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {saveError}
          </div>
        )}
        {saveMessage && <p role="status" className="rounded-lg border border-primary/30 bg-primary/10 px-3 py-2 text-sm text-foreground">{saveMessage}</p>}
        <Button type="button" variant="outline" onClick={() => void saveSnapshot()} disabled={!canSaveSnapshot} className="min-h-[2.8em] w-full gap-2 text-sm font-bold">
          {saving ? <Loader2 className="size-4 animate-spin" aria-hidden /> : <Save className="size-4" aria-hidden />}
          {saving ? tr("wht_cert_ui_saving_snapshot", "กำลังบันทึกลงใบสำคัญ...") : tr("wht_cert_ui_save_snapshot", "บันทึกลงใบสำคัญ")}
        </Button>

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

      {confirmationDialog}
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

// ภาษีถูกหัก ณ ที่จ่าย: ลูกค้า (ผู้จ่ายเงิน) เป็นผู้มีหน้าที่หักภาษีและเป็นผู้ออก 50 ทวิ ให้กิจการ — ถ้าจอนี้ออกใบเอง
// บทบาทผู้หัก/ผู้ถูกหักจะสลับกัน จึงแสดงเฉพาะรายละเอียดที่บันทึกในใบสำคัญให้ตรวจเทียบกับใบที่ได้รับ ไม่มีปุ่มออก/พิมพ์
export function WhtReceivedCertificateDetails({ row, company }: { row: WhtReportRow; company: CompanyHeader | null }) {
  const tr = useBackendText();
  const notSpecified = tr("tax_not_specified", "ยังไม่ระบุ");
  const formOption = FORM_OPTIONS.find((o) => o.value === whtCertificateForm(row.formtype));
  const income = WHT_INCOME_OPTIONS.find((o) => o.value === row.incometype);
  const conditionOption = CONDITION_OPTIONS.find((o) => o.value === conditionValue(row.condition));
  const details: { label: string; value: string; mono?: boolean }[] = [
    { label: tr("wht_received_ui_payer", "ผู้มีหน้าที่หักภาษี (ลูกค้า/ผู้จ่ายเงิน)"), value: row.partnerfullname || row.partnername },
    { label: tr("wht_received_ui_payer_taxid", "เลขประจำตัวผู้เสียภาษีของผู้หักภาษี"), value: row.taxid, mono: true },
    { label: tr("wht_received_ui_payer_address", "ที่อยู่ผู้หักภาษี"), value: row.address },
    { label: tr("wht_received_ui_payee", "ผู้ถูกหักภาษี (กิจการของเรา)"), value: company?.name ?? "" },
    { label: tr("wht_received_ui_payee_taxid", "เลขประจำตัวผู้เสียภาษีของกิจการเรา"), value: company?.taxid ?? "", mono: true },
    { label: tr("wht_received_ui_payee_address", "ที่อยู่ของกิจการเรา (ตามทะเบียนบริษัท)"), value: company?.addressline ?? "" },
    { label: tr("wht_received_ui_form", "แบบยื่นรายการของผู้หักภาษี"), value: formOption ? tr(formOption.key, formOption.th) : "" },
    { label: tr("wht_cert_ui_income", "ประเภทเงินได้พึงประเมินที่จ่าย"), value: income ? tr(income.key, income.th) : "" },
    { label: tr("wht_cert_ui_paid_date", "วันที่จ่ายเงิน"), value: row.paiddate, mono: true },
    { label: tr("wht_cert_ui_condition", "ผู้จ่ายเงิน"), value: conditionOption ? tr(conditionOption.key, conditionOption.th) : "" },
    { label: tr("wht_received_ui_certificate_no", "เลขที่หนังสือรับรองที่ได้รับ"), value: row.certificateno, mono: true },
    { label: tr("wht_received_ui_voucher", "ใบสำคัญที่บันทึก"), value: row.docno ? `${row.docno} (${row.docdate})` : "", mono: true },
  ];

  return (
    <Card className="space-y-4 p-5 shadow-sm print:hidden" data-field="wht-received-details">
      <div className="flex items-start gap-3">
        <FileText className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden />
        <div className="space-y-0.5">
          <h2 className="text-base font-bold text-foreground">{tr("wht_received_ui_title", "หนังสือรับรองการหักภาษี ณ ที่จ่ายที่ได้รับ (50 ทวิ)")}</h2>
          <p role="note" className="text-[0.9rem] leading-relaxed text-muted-foreground">
            {tr("wht_received_ui_hint", "ลูกค้า (ผู้จ่ายเงิน) เป็นผู้ออกหนังสือรับรองนี้ให้กิจการ — จอนี้แสดงรายละเอียดที่บันทึกในใบสำคัญเพื่อตรวจเทียบเท่านั้น ไม่ได้ออกหรือพิมพ์ 50 ทวิ")}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 rounded-xl border border-border bg-muted/30 p-3 text-sm">
        <div>{tr("wht_cert_ui_amount", "จำนวนเงินที่จ่าย")}<div className="font-mono text-base font-bold text-foreground">{formatAmount(row.baseamount, 2)}</div></div>
        <div>{tr("wht_cert_ui_tax", "ภาษีที่หักและนำส่ง")}<div className="font-mono text-base font-bold text-primary">{formatAmount(row.whtamount, 2)}</div></div>
        <p className="col-span-2 leading-relaxed text-muted-foreground" data-field="taxbasesource">{taxBaseHint(row, tr)}</p>
      </div>

      <dl className="grid gap-x-6 gap-y-3 md:grid-cols-2">
        {details.map((item) => (
          <div key={item.label} className="grid gap-0.5">
            <dt className="text-sm text-muted-foreground">{item.label}</dt>
            <dd className={`text-[0.95rem] font-medium leading-normal [overflow-wrap:anywhere] ${item.value ? "text-foreground" : "text-muted-foreground"} ${item.mono && item.value ? "font-mono" : ""}`}>
              {item.value || notSpecified}
            </dd>
          </div>
        ))}
      </dl>
    </Card>
  );
}
