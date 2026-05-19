import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { WorkspaceScreen } from "./workspace-screen";

export default async function WorkspacePage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <WorkspaceScreen {...initialLanguage} />;
}
