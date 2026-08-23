import { NextResponse } from "next/server";
type SetupProxyContext = {
  params: Promise<{ setupPath: string[] }>;
};

export async function POST(_request: Request, context: SetupProxyContext) {
  await context.params;
  return NextResponse.json(
    { success: false, message: "Setup API ถูกปิดจนกว่าจะมี Control Plane Authentication ที่แยกจาก Tenant" },
    { status: 410 },
  );
}
