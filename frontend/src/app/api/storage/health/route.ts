import { NextResponse } from "next/server";

// Storage probing is part of the disabled setup control plane. Never fetch a
// caller-supplied URL here: doing so would turn this public BFF route into SSRF.
export async function POST() {
  return NextResponse.json(
    {
      success: false,
      message: "Storage health check ถูกปิดจนกว่าจะมี Control Plane Authentication ที่แยกจาก Tenant",
    },
    { status: 410 },
  );
}
