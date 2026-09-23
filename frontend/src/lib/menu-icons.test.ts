import { describe, expect, it } from "vitest";
import { flattenMenuItems } from "./menu-data";
import { hasExplicitMenuIcon, menuIconKeyForRoute } from "./menu-icons";

describe("menu icon mapping", () => {
  it("has an explicit migrated icon for every Flutter menu route", () => {
    const items = flattenMenuItems();
    const missing = items.filter((item) => !hasExplicitMenuIcon(item.route)).map((item) => item.route);

    expect(items).toHaveLength(195);
    expect(missing).toEqual([]);
    expect(menuIconKeyForRoute("/product", "master")).toBe("package");
    expect(menuIconKeyForRoute("/productbarcode", "master")).toBe("qr");
    expect(menuIconKeyForRoute("/masterbrandscreen", "master")).toBe("badge");
    expect(menuIconKeyForRoute("/transaction/stocktransfer", "transaction")).toBe("truck");
    expect(menuIconKeyForRoute("/report/chequereceived", "report")).toBe("receipt");
  });

  it("falls back to a category icon for unknown routes", () => {
    expect(menuIconKeyForRoute("/unknown", "report")).toBe("barChart");
  });
});
