import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { CurrencyScreen } from "./currency-screen";

export const metadata: Metadata = {
  title: "สกุลเงิน",
};

export default async function CurrencyPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <CurrencyScreen {...initialLanguage} />;
}
