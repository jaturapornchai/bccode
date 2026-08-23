import { NextResponse } from "next/server";

// The legacy bridge returned profile fields without a Google ID token that
// mainapi could verify. Keep the endpoint fail-closed until the bridge can
// forward the original credential; the supported browser flow is /google/verify.
export async function POST() {
  return NextResponse.json(
    { success: false, status: "disabled", message: "กรุณาเข้าสู่ระบบด้วยปุ่ม Google" },
    { status: 410 },
  );
}
