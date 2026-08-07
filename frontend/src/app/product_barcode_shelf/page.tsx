import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { ProductBarcodeShelfScreen } from "../menu/product-barcode-shelf-screen";

export const metadata: Metadata = {
  title: "พิมพ์ป้ายสินค้า",
};

export default async function ProductBarcodeShelfPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <ProductBarcodeShelfScreen language={initialLanguage.initialLanguage} />;
}
