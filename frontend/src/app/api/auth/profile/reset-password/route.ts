import { NextResponse } from "next/server";

export async function PUT(request: Request) {
  try {
    await request.json();
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }
  return NextResponse.json(
    { success: false, message: "ระบบส่งลิงก์รีเซ็ตรหัสผ่านยังไม่พร้อมใช้งาน" },
    { status: 501 },
  );
}
