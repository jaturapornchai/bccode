import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { ProductPriceHistoryScreen } from "../menu/product-price-history-screen";

export const metadata: Metadata = {
  title: "ประวัติแก้ไขราคา",
};

export default async function PriceHistoryPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <ProductPriceHistoryScreen language={initialLanguage.initialLanguage} />;
}
