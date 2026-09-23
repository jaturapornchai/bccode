import { redirect } from "next/navigation";

// 2026-09-23: backend LINE OA (Mongo-based /api/user/lineoa/*) ถูกถอดแล้ว —
// ระบบใช้ PostgreSQL อย่างเดียว ยังไม่มี API รองรับฟีเจอร์นี้ ส่งกลับไปหน้าเวิร์กสเปซ
// แทนการแสดงฟอร์มที่เรียก API ที่ไม่มีอยู่จริง (ดู lib/menu-screen-status.ts RETIRED_BACKEND_ROUTES)
export default function LineOaPage() {
  redirect("/workspace");
}
