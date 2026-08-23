import { expect, test } from "@playwright/test";

test("Dev Login creates a refreshable localhost session without persisting tokens", async ({ page }) => {
  let accessToken = "";
  try {
    await page.goto("/");

    const button = page.getByRole("button", { name: /เข้าทดสอบระบบ \(Dev Login\)/ });
    await expect(button).toBeVisible();
    const loginResponsePromise = page.waitForResponse(
      (response) => response.url().endsWith("/api/auth/dev-login") && response.request().method() === "POST",
    );
    await button.click();
    const loginResponse = await loginResponsePromise;
    const login = (await loginResponse.json()) as { token?: string };
    accessToken = login.token ?? "";
    await page.waitForURL("**/holding");

    const storedAuth = JSON.parse(await page.evaluate(() => localStorage.getItem("bc_auth")) ?? "{}");
    expect(storedAuth).toEqual(expect.objectContaining({ method: "dev" }));
    expect(storedAuth.username).toEqual(expect.any(String));
    expect(storedAuth).not.toHaveProperty("token");
    expect(storedAuth).not.toHaveProperty("refresh");

    await page.reload();
    await expect(page).toHaveURL(/\/holding$/);
  } finally {
    if (accessToken) {
      const logout = await page.request.post(new URL("/api/auth/logout", page.url()).toString(), {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      expect(logout.ok()).toBe(true);
    }
  }
});
