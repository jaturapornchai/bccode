import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { HoldingScreen } from "./holding-screen";

export const metadata: Metadata = {
  title: "เลือกกลุ่มกิจการ",
};

export default async function HoldingPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <HoldingScreen initialLanguage={initialLanguage.initialLanguage} />;
}
