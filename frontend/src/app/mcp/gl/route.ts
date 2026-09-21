import { proxyIntegration } from "@/lib/integration-proxy";

const proxy = (request: Request) => proxyIntegration(request, "mcp", "/mcp/gl");
export const POST = proxy;
export const GET = proxy;
export const DELETE = proxy;
