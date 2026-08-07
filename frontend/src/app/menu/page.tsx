import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { MainMenuScreen } from "./main-menu-screen";

export const metadata: Metadata = {
  title: "เมนูหลัก",
};

export default async function MenuPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <MainMenuScreen {...initialLanguage} />;
}
