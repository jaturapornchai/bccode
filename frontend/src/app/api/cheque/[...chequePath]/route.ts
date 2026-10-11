import { NextResponse } from "next/server";
import { isBankCodeKind } from "@/lib/bank-code-kinds";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ chequePath: string[] }> };

// ข้อความผิดพลาดของ BFF เป็น key ใน languages.tsv (code = message = key) — จอแปลเป็นภาษาผู้ใช้เอง ไม่ hard-code ข้อความที่นี่
const rejected = (key: string, status = 400) => NextResponse.json({ success: false, code: key, message: key }, { status });

const COMMAND_RESOURCES = ["codes", "bank-contacts", "registers", "docs", "recalculate", "petty-cash", "bank-recalculate"];
// พารามิเตอร์ที่ส่งต่อได้: รายการรหัส/ผู้ติดต่อ, ตัวกรองทะเบียนเช็ค, รายการเลือกเช็คของเอกสาร, รายการเอกสาร และยอดคงเหลือ
const ALLOWED_QUERY = [
  "q", "page", "limit", "bank_code", "bank_branch_code", "active",
  "type", "status", "book", "partner", "due_from", "due_to", "transflag", "cancel_type", "cancel_passed", "doc_date", "date_from", "date_to",
];
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
// เอกสารสถานะเช็ค 419-426 (นำฝาก ผ่าน คืน นำเข้าใหม่ ยกเลิก ขายลด เช็คจ่ายผ่าน ยกเลิกเช็คจ่าย)
const DOC_FLAG = /^(419|42[0-6])$/;
// เลขที่เอกสาร varchar(30) ของ backend; กันส่วน path พิเศษ/อักขระควบคุมก่อนต่อ URL
const isDocNo = (value: string) => value.trim() !== "" && value.length <= 30 && value !== "." && value !== ".." && !/[\u0000-\u001f\u007f/\\?#]/.test(value);

// isReadPath: เส้นทาง GET ที่ backend มีจริง (/cheque/v1/...) — ที่เหลือปฏิเสธโดยไม่เรียก backend
function isReadPath(path: string[] | undefined): path is string[] {
  if (!path?.length) return false;
  const [head, second, third, fourth] = path;
  switch (head) {
    case "codes":
      return path.length === 2 && isBankCodeKind(second);
    case "bank-contacts":
    case "recalculate":
    case "petty-cash":
      return path.length === 1;
    case "registers":
      return path.length === 1 || (path.length === 2 && (second === "selectable" || UUID.test(second)));
    case "docs":
      if (path.length < 2 || path.length > 4 || !DOC_FLAG.test(second)) return false;
      return path.length === 2 || (isDocNo(third) && (path.length === 3 || fourth === "events"));
    case "balances":
      return path.length === 2 && (second === "partners" || second === "books");
    default:
      return false;
  }
}

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { chequePath } = await context.params;
  if (!isReadPath(chequePath)) return rejected("cheque_err_resource_invalid", 404);

  try {
    const query = new URLSearchParams();
    new URL(request.url).searchParams.forEach((value, key) => {
      if (ALLOWED_QUERY.includes(key)) query.set(key, value);
    });
    const targetPath = `/cheque/v1/${chequePath.map(encodeURIComponent).join("/")}?${query.toString()}`;
    return proxyMainApiJson(request, getMainApiUrl(getBackendUrlFromRequest(request)), targetPath, { method: "GET" }, { userErrorStatusOk: true });
  } catch {
    return rejected("cheque_err_connection_invalid");
  }
}

export async function POST(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { chequePath } = await context.params;
  if (chequePath?.length !== 1 || chequePath[0] !== "command") return rejected("cheque_err_resource_invalid", 404);

  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return rejected("cheque_err_invalid_payload");
  }
  if (!isRecord(body) || typeof body.resource !== "string" || !COMMAND_RESOURCES.includes(body.resource)) {
    return rejected("cheque_err_resource_invalid");
  }

  try {
    return proxyMainApiJson(
      request,
      getMainApiUrl(getBackendUrlFromRequest(request)),
      "/cheque/v1/command",
      { method: "POST", body: JSON.stringify(body) },
      { userErrorStatusOk: true },
    );
  } catch {
    return rejected("cheque_err_connection_invalid");
  }
}
