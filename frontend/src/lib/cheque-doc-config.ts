// ป้ายและกติกาของจอทะเบียนเช็ค (เช็ครับ/เช็คจ่าย) และเอกสารเปลี่ยนสถานะ 8 ชนิด (419-426)
// ไฟล์นี้ไม่ import อะไรเลย เพื่อให้ menu-screen-status และจอใช้ร่วมกันได้; กติกาสมุดเงินฝากสะท้อน backend cheque/state.go
// (bookPolicyFor) ซึ่งเป็นตัวตัดสินจริง — จอใช้เพื่อเปิด/ปิดช่องและทำเครื่องหมาย "บังคับ" เท่านั้น

/** ชนิดเช็คในทะเบียน: 0 = เช็ครับ, 1 = เช็คจ่าย (ตรงกับ backend cheque.TypeReceived / TypeIssued) */
export type ChequeKind = 0 | 1;
export const CHEQUE_TYPE_RECEIVED: ChequeKind = 0;
export const CHEQUE_TYPE_ISSUED: ChequeKind = 1;

/** [key ใน languages.tsv, ข้อความไทยสำรอง] — ป้ายที่เลือกด้วยค่าสถานะ จึงถือเป็นคู่ key/ข้อความ แล้วส่งเข้า tr() ตอนแสดง */
export type LabelPair = readonly [key: string, fallback: string];

/** สถานะเช็ครับ 0..6 (Champ BCChqIn.Status) */
export const CHEQUE_STATUS_IN: readonly LabelPair[] = [
  ["cheque_status_in_0", "เช็คในมือ"],
  ["cheque_status_in_1", "เช็คฝากธนาคาร"],
  ["cheque_status_in_2", "เช็คผ่าน"],
  ["cheque_status_in_3", "เช็คคืน"],
  ["cheque_status_in_4", "เช็คยกเลิก"],
  ["cheque_status_in_5", "ขายลดเช็ค"],
  ["cheque_status_in_6", "เช็คคืนฝากธนาคาร"],
];

/** สถานะเช็คจ่าย 0..3 (Champ BCChqOut.Status; 3 = เตรียมจ่าย) */
export const CHEQUE_STATUS_OUT: readonly LabelPair[] = [
  ["cheque_status_out_0", "เช็คในมือ"],
  ["cheque_status_out_1", "เช็คผ่าน"],
  ["cheque_status_out_2", "เช็คยกเลิก"],
  ["cheque_status_out_3", "เตรียมจ่าย"],
];

/** ชนิดการยกเลิกเช็ครับของเอกสาร 423 (Champ CancelStatus 0..3) */
export const CHEQUE_CANCEL_TYPES: readonly LabelPair[] = [
  ["cheque_cancel_type_0", "ยกเลิกเช็คในมือ"],
  ["cheque_cancel_type_1", "ยกเลิกเช็คที่ผ่านแล้ว"],
  ["cheque_cancel_type_2", "ยกเลิกเช็คฝาก"],
  ["cheque_cancel_type_3", "ยกเลิกเช็คคืนฝากธนาคาร"],
];

const CHEQUE_STATUS_UNKNOWN: LabelPair = ["cheque_status_unknown", "ไม่ทราบสถานะ"];

export function chequeStatusLabel(chequeType: number, status: number): LabelPair {
  return (chequeType === CHEQUE_TYPE_ISSUED ? CHEQUE_STATUS_OUT : CHEQUE_STATUS_IN)[status] ?? CHEQUE_STATUS_UNKNOWN;
}

/** โทนป้ายสถานะในรายการ: เช็คผ่าน = success, เช็คยกเลิก = warning, อื่นๆ = outline (ข้อความสถานะแสดงเสมอ ไม่พึ่งสีอย่างเดียว) */
export function chequeStatusTone(chequeType: number, status: number): "success" | "warning" | "outline" {
  const [cleared, cancelled] = chequeType === CHEQUE_TYPE_ISSUED ? [1, 2] : [2, 4];
  if (status === cleared) return "success";
  return status === cancelled ? "warning" : "outline";
}

/**
 * สถานะที่ตั้งในจอทะเบียนได้ (backend validRegisterStatus): ตอนสร้าง เช็ครับ 0..6 / เช็คจ่าย 0 หรือ 3;
 * ตอนแก้ (D8 ต้องยืนยันก่อน) เลือกได้ทุกสถานะของชนิดนั้น
 */
export function registerStatusChoices(chequeType: number, creating: boolean): number[] {
  if (chequeType !== CHEQUE_TYPE_ISSUED) return CHEQUE_STATUS_IN.map((_, status) => status);
  return creating ? [0, 3] : CHEQUE_STATUS_OUT.map((_, status) => status);
}

/** ป้ายหัวจอทะเบียน (ตรงกับชื่อเมนู) */
export const CHEQUE_REGISTER_TITLES: Record<ChequeKind, LabelPair> = {
  0: ["cheque_received", "บันทึกเช็ครับ"],
  1: ["cheque_issued", "บันทึกเช็คจ่าย"],
};

export const CHEQUE_REGISTER_ROUTES: Record<ChequeKind, string> = {
  0: "/transaction/chequereceived",
  1: "/transaction/chequeissued",
};

/** สถานะการลงบัญชีของเอกสาร (gl_status จาก backend): none = ยังไม่มีใบสำคัญ, no_mapping = เอกสารนี้ยังไม่มีรูปแบบลงบัญชี */
export const CHEQUE_GL_STATUS: Record<string, { label: LabelPair; tone: "success" | "warning" | "outline" }> = {
  posted: { label: ["cheque_gl_posted", "ลงบัญชีแล้ว"], tone: "success" },
  none: { label: ["cheque_gl_none", "ยังไม่ลงบัญชี"], tone: "outline" },
  no_mapping: { label: ["cheque_gl_no_mapping", "ยังไม่มีรูปแบบลงบัญชีของเอกสารนี้ จึงไม่สร้างใบสำคัญ"], tone: "warning" },
  not_required: { label: ["cheque_gl_not_required", "เอกสารนี้ไม่ต้องลงบัญชี"], tone: "outline" },
};

export function chequeGlStatus(glStatus: string): { label: LabelPair; tone: "success" | "warning" | "outline" } {
  return CHEQUE_GL_STATUS[glStatus] ?? CHEQUE_GL_STATUS.none;
}

/** กติกาสมุดเงินฝากของเอกสาร: required = ต้องระบุ, match = เช็คต้องอยู่สมุดเดียวกับเอกสาร, none = เอกสารไม่ใช้สมุด */
export type ChequeBookPolicy = { required: boolean; match: boolean; none: boolean };

export type ChequeDocConfig = {
  flag: number;
  route: string;
  /** key ของ doctype ฝั่ง backend (ใช้จับคู่ชื่อรายการในตารางอ้างอิงของจอทะเบียน) */
  doctypeKey: string;
  title: LabelPair;
  chequeType: ChequeKind;
  /** false = จอนำฝาก 419: ไม่มีคอลัมน์ค่าใช้จ่าย/ยอดหัก มีแค่ "จำนวนเงินรวม" */
  hasFee: boolean;
  /** true = จอนำฝาก 419: มีคอลัมน์หมายเหตุอ่านอย่างเดียว (คำอธิบายของทะเบียนเช็ค) */
  showRemark: boolean;
  /** 423: คอมโบ "เช็คที่ต้องการยกเลิก" + ข้อความเตือนสีแดง */
  cancelTypeCombo: boolean;
  /** 426: ช่องติ๊ก "ยกเลิกเช็คที่ผ่านแล้ว" */
  cancelPassedBox: boolean;
  book: ChequeBookPolicy;
};

const BOOK_REQUIRED_MATCH: ChequeBookPolicy = { required: true, match: true, none: false };

export const CHEQUE_DOC_CONFIGS: readonly ChequeDocConfig[] = [
  { flag: 419, route: "/banking/cheques/deposit", doctypeKey: "chequedeposit", title: ["cheque_deposit", "บันทึกนำฝากเช็ครับ"], chequeType: 0,
    hasFee: false, showRemark: true, cancelTypeCombo: false, cancelPassedBox: false, book: { required: true, match: false, none: false } },
  { flag: 420, route: "/banking/cheques/received-clear", doctypeKey: "chequereceivedclear", title: ["cheque_received_clear", "บันทึกเช็ครับผ่าน"], chequeType: 0,
    hasFee: true, showRemark: false, cancelTypeCombo: false, cancelPassedBox: false, book: BOOK_REQUIRED_MATCH },
  { flag: 421, route: "/banking/cheques/received-return", doctypeKey: "chequereceivedreturn", title: ["cheque_received_return", "บันทึกเช็ครับคืน"], chequeType: 0,
    hasFee: true, showRemark: false, cancelTypeCombo: false, cancelPassedBox: false, book: BOOK_REQUIRED_MATCH },
  { flag: 422, route: "/banking/cheques/redeposit", doctypeKey: "chequeredeposit", title: ["cheque_redeposit", "บันทึกนำเช็คเข้าใหม่"], chequeType: 0,
    hasFee: true, showRemark: false, cancelTypeCombo: false, cancelPassedBox: false, book: BOOK_REQUIRED_MATCH },
  { flag: 423, route: "/banking/cheques/received-cancel", doctypeKey: "chequereceivedcancel", title: ["cheque_received_cancel", "บันทึกยกเลิกเช็ครับ"], chequeType: 0,
    hasFee: true, showRemark: false, cancelTypeCombo: true, cancelPassedBox: false, book: BOOK_REQUIRED_MATCH },
  { flag: 424, route: "/banking/cheques/discount", doctypeKey: "chequediscount", title: ["cheque_discount", "บันทึกขายลดเช็ครับ"], chequeType: 0,
    hasFee: true, showRemark: false, cancelTypeCombo: false, cancelPassedBox: false, book: { required: false, match: true, none: false } },
  { flag: 425, route: "/banking/cheques/issued-clear", doctypeKey: "chequeissuedclear", title: ["cheque_issued_clear", "บันทึกเช็คจ่ายผ่าน"], chequeType: 1,
    hasFee: true, showRemark: false, cancelTypeCombo: false, cancelPassedBox: false, book: BOOK_REQUIRED_MATCH },
  { flag: 426, route: "/banking/cheques/issued-cancel", doctypeKey: "chequeissuedcancel", title: ["cheque_issued_cancel", "บันทึกยกเลิกเช็คจ่าย"], chequeType: 1,
    hasFee: true, showRemark: false, cancelTypeCombo: false, cancelPassedBox: true, book: BOOK_REQUIRED_MATCH },
];

export function chequeDocConfig(flag: number): ChequeDocConfig | undefined {
  return CHEQUE_DOC_CONFIGS.find((config) => config.flag === flag);
}

/** ชื่อรายการเอกสารเปลี่ยนสถานะในตารางอ้างอิงใต้หัวทะเบียน (kind_key = key ของ doctype); ไม่รู้จัก = undefined */
export function chequeDocTitleByKind(kindKey: string): LabelPair | undefined {
  return CHEQUE_DOC_CONFIGS.find((config) => config.doctypeKey === kindKey)?.title;
}

/** กติกาสมุดเงินฝากตามชนิดการยกเลิก: 423 ชนิด 0 (ยกเลิกเช็คในมือ) ไม่ใช้สมุด — ชนิดอื่นต้องมีสมุดและเช็คต้องอยู่สมุดเดียวกัน */
export function chequeDocBookPolicy(config: ChequeDocConfig, cancelType: number | null): ChequeBookPolicy {
  if (config.cancelTypeCombo && (cancelType ?? 0) === 0) return { required: false, match: false, none: true };
  return config.book;
}

export type ChequeScreenTarget = { kind: "register"; chequeType: ChequeKind } | { kind: "doc"; flag: number };

/** route ของทั้ง 10 จอ: ทะเบียน 2 + เอกสารสถานะ 8 */
export const CHEQUE_SCREEN_ROUTES: readonly string[] = [...Object.values(CHEQUE_REGISTER_ROUTES), ...CHEQUE_DOC_CONFIGS.map((config) => config.route)];

/** route → จอเช็ค (null = ไม่ใช่จอเช็ค); ตัด query ออกก่อนเทียบ */
export function chequeScreenFromRoute(route: string): ChequeScreenTarget | null {
  const path = route.split("?")[0] ?? "";
  if (path === CHEQUE_REGISTER_ROUTES[0]) return { kind: "register", chequeType: CHEQUE_TYPE_RECEIVED };
  if (path === CHEQUE_REGISTER_ROUTES[1]) return { kind: "register", chequeType: CHEQUE_TYPE_ISSUED };
  const config = CHEQUE_DOC_CONFIGS.find((item) => item.route === path);
  return config ? { kind: "doc", flag: config.flag } : null;
}
