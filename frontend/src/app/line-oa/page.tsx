import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { LineOaLinkScreen } from "./line-oa-link-screen";

export const metadata: Metadata = {
  title: "เชื่อมต่อ LINE OA",
};

export default async function LineOaPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <LineOaLinkScreen {...initialLanguage} />;
}
