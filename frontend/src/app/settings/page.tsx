import type { Metadata } from "next";
import { SettingsScreen } from "./settings-screen";

export const metadata: Metadata = {
  title: "ศูนย์ตั้งค่าระบบ",
};

export default function SettingsPage() {
  return <SettingsScreen />;
}
