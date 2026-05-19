import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { LineOaLinkScreen } from "./line-oa-link-screen";

export default async function LineOaPage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <LineOaLinkScreen {...initialLanguage} />;
}
