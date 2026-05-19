import { afterEach, describe, expect, it, vi } from "vitest";
import { POST } from "./route";

describe("register username route", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("requires a username", async () => {
    const response = await POST(new Request("http://localhost/api/auth/register-username", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: " ",
        password: "secret",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(400);
    expect(json).toMatchObject({ success: false });
  });

  it("registers a username/password user through mainapi", async () => {
    const fetchMock = vi.fn(async (url: string | URL | Request, init?: RequestInit) => {
      expect(String(url)).toBe("http://localhost:8888/register-username");
      expect(init?.method).toBe("POST");
      expect(JSON.parse(String(init?.body))).toEqual({
        username: "new-user",
        password: "secret",
        name: "New User",
      });
      return Response.json({ success: true, id: "user-id" }, { status: 201 });
    });
    vi.stubGlobal("fetch", fetchMock);

    const response = await POST(new Request("http://localhost/api/auth/register-username", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        backendUrl: "http://localhost:8888/goapi",
        username: " new-user ",
        password: "secret",
        name: " New User ",
      }),
    }));
    const json = await response.json();

    expect(response.status).toBe(201);
    expect(json).toMatchObject({
      success: true,
      id: "user-id",
      backendUrl: "http://localhost:8888/goapi",
      mainApiUrl: "http://localhost:8888",
    });
  });
});
