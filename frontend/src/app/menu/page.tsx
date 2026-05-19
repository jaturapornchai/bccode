import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { MainMenuScreen } from "./main-menu-screen";

export default async function MenuPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <MainMenuScreen {...initialLanguage} />;
}
