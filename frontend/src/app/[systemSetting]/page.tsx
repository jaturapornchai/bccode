import { notFound, redirect } from "next/navigation";
import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { getSystemSettingConfig } from "@/lib/system-setting-screens";
import { SystemSettingsScreen } from "../system-settings/system-settings-screen";

type SystemSettingPageProps = {
  params: Promise<{ systemSetting: string }>;
};

export async function generateMetadata({ params }: SystemSettingPageProps): Promise<Metadata> {
  const { systemSetting } = await params;
  const config = getSystemSettingConfig(systemSetting);
  return { title: config?.title.th ?? "ตั้งค่าระบบ" };
}

export default async function SystemSettingPage({ params }: SystemSettingPageProps) {
  const { systemSetting } = await params;
  if (systemSetting === "permissionlink") redirect("/user");
  if (systemSetting === "approvalsetting") redirect("/workspace");
  if (systemSetting === "productcategorylist") redirect("/productcategorygroupselectscreen");
  const config = getSystemSettingConfig(systemSetting);
  if (!config) notFound();
  const initialLanguage = await getInitialBackendLanguage();
  return <SystemSettingsScreen route={config.route} {...initialLanguage} />;
}
