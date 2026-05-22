import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { ProductBarcodeShelfScreen } from "../menu/product-barcode-shelf-screen";

export default async function ProductBarcodeShelfPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <ProductBarcodeShelfScreen language={initialLanguage.initialLanguage} />;
}
