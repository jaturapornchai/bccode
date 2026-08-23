import { NextResponse } from "next/server";

// LINE is not an approved user authentication method. LINE account linking is
// handled by its authenticated profile flow and must not mint a login session.
export async function POST() {
  return NextResponse.json(
    { success: false, status: "disabled", message: "LINE Login ไม่ได้เปิดใช้งาน" },
    { status: 410 },
  );
}
