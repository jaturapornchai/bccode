import { NextResponse } from "next/server";
import {
  getBackendUrlFromRequest,
  getMainApiUrl,
  proxyMainApiJson,
  type ApiProxyBody,
} from "@/lib/workspace-api";

type WorkspaceProxyContext = {
  params: Promise<{ workspacePath: string[] }>;
};

export async function GET(request: Request, context: WorkspaceProxyContext) {
  const { workspacePath } = await context.params;
  const path = workspacePath.join("/");

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const url = new URL(request.url);
  switch (path) {
    case "shops":
      return proxyMainApiJson(request, mainApiUrl, "/list-shop?limit=100", { method: "GET" });
    case "shop-info": {
      const shopid = url.searchParams.get("shopid")?.trim() ?? "";
      if (!shopid) return NextResponse.json({ success: false, message: "ไม่พบรหัสกิจการ" }, { status: 400 });
      return proxyMainApiJson(request, mainApiUrl, `/shop/${encodeURIComponent(shopid)}`, { method: "GET" });
    }
    case "branches": {
      const offset = url.searchParams.get("offset") ?? "0";
      const limit = url.searchParams.get("limit") ?? "100";
      const query = url.searchParams.get("q") ?? "";
      const branchPath = `/organization/branch/list?offset=${encodeURIComponent(offset)}&limit=${encodeURIComponent(limit)}&q=${encodeURIComponent(query)}`;
      return proxyMainApiJson(request, mainApiUrl, branchPath, { method: "GET" });
    }
    default:
      return NextResponse.json({ success: false, message: "ไม่พบ workspace endpoint" }, { status: 404 });
  }
}

export async function POST(request: Request, context: WorkspaceProxyContext) {
  const { workspacePath } = await context.params;
  const path = workspacePath.join("/");

  let body: ApiProxyBody;
  try {
    body = (await request.json()) as ApiProxyBody;
  } catch {
    return NextResponse.json({ success: false, message: "รูปแบบข้อมูลไม่ถูกต้อง" }, { status: 400 });
  }

  let mainApiUrl: string;
  try {
    mainApiUrl = getMainApiUrl(getBackendUrlFromRequest(request, body));
  } catch (error) {
    return NextResponse.json(
      { success: false, message: error instanceof Error ? error.message : "Backend URL ไม่ถูกต้อง" },
      { status: 400 },
    );
  }

  const { backendUrl: _backendUrl, ...payload } = body;
  void _backendUrl;

  switch (path) {
    case "select-shop": {
      const shopid = typeof payload.shopid === "string" ? payload.shopid.trim() : "";
      if (!shopid) return NextResponse.json({ success: false, message: "ไม่พบรหัสกิจการ" }, { status: 400 });
      return proxyMainApiJson(request, mainApiUrl, "/select-shop", {
        method: "POST",
        body: JSON.stringify({ shopid }),
      });
    }
    case "branch": {
      const branch = payload.branch;
      if (!branch || typeof branch !== "object") {
        return NextResponse.json({ success: false, message: "ไม่พบข้อมูลสาขา" }, { status: 400 });
      }
      return proxyMainApiJson(request, mainApiUrl, "/organization/branch", {
        method: "POST",
        body: JSON.stringify(branch),
      });
    }
    case "create-shop":
      return proxyMainApiJson(request, mainApiUrl, "/create-shop", {
        method: "POST",
        body: JSON.stringify(payload),
      });
    default:
      return NextResponse.json({ success: false, message: "ไม่พบ workspace endpoint" }, { status: 404 });
  }
}
