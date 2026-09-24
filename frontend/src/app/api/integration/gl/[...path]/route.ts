import { proxyIntegration } from "@/lib/integration-proxy";
import { GL_REPORTS, GL_RESOURCES } from "@/lib/general-ledger";

type Context = { params: Promise<{ path: string[] }> };
const missing = () => Response.json({ success: false, message: "Not found" }, { status: 404 });
export async function GET(request: Request, context: Context) {
  const { path } = await context.params;
  if (path.some((segment) => !/^[\p{L}\p{M}\p{N}_.-]+$/u.test(segment) || segment === "." || segment === "..")) return missing();
  const valid = path[0] === "reports"
    ? path.length === 2 && (GL_REPORTS as readonly string[]).includes(path[1])
    : path[0] === "journal-support" ? path.length === 1
    : path[0] === "journal-reviews" ? path.length === 2
    : (GL_RESOURCES as readonly string[]).includes(path[0]) && path.length <= 2;
  if (!valid) return missing();
  const query = new URLSearchParams();
  const keys = ["q", "page", "limit", "from", "to", "fiscalyear", "accountcode", "branchcode", "departmentcode", "projectcode", "bookcode", "status", "kind", "snapshot", "asof", "companywide"];
  new URL(request.url).searchParams.forEach((value, key) => { if (keys.includes(key)) query.set(key, value); });
  return proxyIntegration(request, "api", `/integration/gl/v2/${path.map(encodeURIComponent).join("/")}?${query}`);
}
export async function POST(request: Request, context: Context) {
  if ((await context.params).path.join("/") !== "command") return missing();
  return proxyIntegration(request, "api", "/integration/gl/v2/command");
}
