import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { CurrencyScreen } from "./currency-screen";

export default async function CurrencyPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <CurrencyScreen {...initialLanguage} />;
}
