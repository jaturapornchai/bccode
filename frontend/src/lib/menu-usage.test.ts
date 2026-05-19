import { describe, expect, it } from "vitest";
import { getFrequentMenuEntries, getFrequentMenuItems, menuUsageStorageKey, readMenuUsage, recordMenuUsage } from "./menu-usage";
import type { MenuItem } from "./menu-data";

function createStorage() {
  const data = new Map<string, string>();
  return {
    getItem: (key: string) => data.get(key) ?? null,
    setItem: (key: string, value: string) => data.set(key, value),
    raw: data,
  };
}

const items: MenuItem[] = [
  { id: "purchase", label: { th: "ซื้อ", en: "Purchase" }, route: "/purchase", category: "transaction" },
  { id: "sale", label: { th: "ขาย", en: "Sale" }, route: "/sale", category: "transaction" },
  { id: "report", label: { th: "รายงาน", en: "Report" }, route: "/report", category: "report" },
];

describe("menu usage", () => {
  it("creates separate storage keys per login user", () => {
    const firstUser = menuUsageStorageKey({ username: "first@example.com", backendUrl: "http://localhost:8888/goapi", method: "google", profile: null });
    const secondUser = menuUsageStorageKey({ username: "second@example.com", backendUrl: "http://localhost:8888/goapi", method: "google", profile: null });

    expect(firstUser).not.toBe(secondUser);
    expect(firstUser).not.toContain("first@example.com");
  });

  it("records usage and sorts by count then recent use", () => {
    const storage = createStorage();
    const key = "usage";

    recordMenuUsage(storage, key, "purchase", 1000);
    recordMenuUsage(storage, key, "sale", 2000);
    const usage = recordMenuUsage(storage, key, "purchase", 3000);

    expect(getFrequentMenuItems(items, usage).map((item) => item.id)).toEqual(["purchase", "sale"]);
    expect(getFrequentMenuEntries(items, usage).map((entry) => ({ id: entry.item.id, percent: entry.percent }))).toEqual([
      { id: "purchase", percent: 67 },
      { id: "sale", percent: 33 },
    ]);
  });

  it("ignores corrupted records", () => {
    const storage = createStorage();
    storage.setItem("usage", JSON.stringify({ purchase: { count: "x" }, sale: { count: 1, lastUsedAt: 2000 } }));

    expect(readMenuUsage(storage, "usage")).toEqual({ sale: { count: 1, lastUsedAt: 2000 } });
  });

  it("limits frequent menus to 20 entries", () => {
    const manyItems = Array.from({ length: 25 }, (_, index) => ({
      id: `menu-${index}`,
      label: { th: `เมนู ${index}`, en: `Menu ${index}` },
      route: `/menu-${index}`,
      category: "transaction" as const,
    }));
    const usage = Object.fromEntries(manyItems.map((item, index) => [item.id, { count: index + 1, lastUsedAt: index + 1 }]));

    expect(getFrequentMenuEntries(manyItems, usage)).toHaveLength(20);
    expect(getFrequentMenuEntries(manyItems, usage)[0]?.item.id).toBe("menu-24");
  });
});
