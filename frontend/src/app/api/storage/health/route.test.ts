import { afterEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("storage health route", () => {
  afterEach(() => vi.restoreAllMocks());

  it("fails closed without fetching a caller-supplied endpoint", async () => {
    const fetchSpy = vi.spyOn(globalThis, "fetch");

    const response = await POST();

    expect(response.status).toBe(410);
    expect(fetchSpy).not.toHaveBeenCalled();
    expect(await response.json()).toEqual({
      success: false,
      message: "Storage health check ถูกปิดจนกว่าจะมี Control Plane Authentication ที่แยกจาก Tenant",
    });
  });
});
