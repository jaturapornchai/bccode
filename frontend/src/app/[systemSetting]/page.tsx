import { notFound } from "next/navigation";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { getSystemSettingConfig } from "@/lib/system-setting-screens";
import { SystemSettingsScreen } from "../system-settings/system-settings-screen";

type SystemSettingPageProps = {
  params: Promise<{ systemSetting: string }>;
};

export default async function SystemSettingPage({ params }: SystemSettingPageProps) {
  const { systemSetting } = await params;
  const config = getSystemSettingConfig(systemSetting);
  if (!config) notFound();
  const initialLanguage = await getInitialBackendLanguage();
  return <SystemSettingsScreen route={config.route} {...initialLanguage} />;
}
