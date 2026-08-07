import type { Metadata } from "next";
import { LoginWrapper } from "./login-wrapper";

export const metadata: Metadata = {
  title: { absolute: "เข้าสู่ระบบ | BC Ai Account" },
};

// Force Next.js Turbopack compiler to clear cache and reload
export default function HomePage() {
  return <LoginWrapper />;
}
