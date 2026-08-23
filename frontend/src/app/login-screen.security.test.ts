import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(fileURLToPath(new URL("./login-screen.tsx", import.meta.url)), "utf8");

describe("login screen security", () => {
  it("shows Dev Login only after checking the browser loopback hostname", () => {
    expect(source).toContain("setIsLoopback(isLoopbackHostname(window.location.hostname))");
    expect(source).toMatch(/\{isLoopback \? \([\s\S]*เข้าทดสอบระบบ \(Dev Login\)/);
    expect(source).toContain('fetch("/api/auth/dev-login"');
    expect(source).not.toMatch(/BCAI_DEV_LOGIN_SECRET|X-BC-Dev-Login-Secret/);
  });

  it("has no inline credential login shortcut", () => {
    expect(source).not.toMatch(
      /performLogin\(\s*\{\s*username:\s*["'][^"']+["'],\s*password:\s*["'][^"']+["']/s,
    );
  });

  it("does not block a verified Google login on the optional profile request", () => {
    const googleFlow = source.slice(
      source.indexOf("async function handleGoogleCredential"),
      source.indexOf("async function handleLogin"),
    );

    expect(googleFlow).not.toContain("loadLoginProfile");
    expect(googleFlow.indexOf("persistLogin(")).toBeLessThan(googleFlow.indexOf('router.push("/holding")'));
  });
});
