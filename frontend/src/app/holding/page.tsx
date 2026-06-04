import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { HoldingScreen } from "./holding-screen";

export default async function HoldingPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <HoldingScreen initialLanguage={initialLanguage.initialLanguage} />;
}
