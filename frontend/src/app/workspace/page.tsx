import type { Metadata } from "next";
import { getInitialBackendLanguage } from "@/lib/backend-language-server";
import { WorkspaceScreen } from "./workspace-screen";

export const metadata: Metadata = {
  title: "เลือกบริษัทและสาขา",
};

export default async function WorkspacePage() {
  const initialLanguage = await getInitialBackendLanguage();
  return <WorkspaceScreen {...initialLanguage} />;
}
