import { NextResponse } from "next/server";
import { isBankCodeKind } from "@/lib/bank-code-kinds";
import { getBackendUrlFromRequest, getMainApiUrl, isRecord, proxyMainApiJson, requireBearerToken } from "@/lib/workspace-api";

type Context = { params: Promise<{ chequePath: string[] }> };

// ข้อความผิดพลาดของ BFF เป็น key ใน languages.tsv (code = message = key) — จอแปลเป็นภาษาผู้ใช้เอง ไม่ hard-code ข้อความที่นี่
const rejected = (key: string, status = 400) => NextResponse.json({ success: false, code: key, message: key }, { status });

const COMMAND_RESOURCES = ["codes", "bank-contacts"];
// พารามิเตอร์ที่ส่งต่อได้: รายการรหัส (q/page/limit/bank_code/active) และผู้ติดต่อ (bank_code/bank_branch_code)
const ALLOWED_QUERY = ["q", "page", "limit", "bank_code", "bank_branch_code", "active"];

export async function GET(request: Request, context: Context) {
  const authorization = requireBearerToken(request);
  if (typeof authorization !== "string") return authorization;

  const { chequePath } = await context.params;
  const isCodes = chequePath?.length === 2 && chequePath[0] === "codes" && isBankCodeKind(chequePath[1]);
  const isContacts = chequePath?.length === 1 && chequePath[0] === "bank-contacts";
  if (!isCodes && !isContacts) return rejected("cheque_err_resource_invalid", 404);

  try {
    const query = new URLSearchParams();
    new URL(request.url).searchParams.forEach((value, key) => {
      if (ALLOWED_QUERY.includes(key)) query.set(key, value);
    });
    const targetPath = `/cheque/v1/${chequePath.join("/")}?${query.toString()}`;
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
