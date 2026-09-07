import { describe, expect, it } from "vitest";
import { flattenMenuItems } from "./menu-data";
import { hasExplicitMenuIcon, menuIconKeyForRoute } from "./menu-icons";

describe("menu icon mapping", () => {
  it("has an explicit migrated icon for every Flutter menu route", () => {
    const items = flattenMenuItems();
    const missing = items.filter((item) => !hasExplicitMenuIcon(item.route)).map((item) => item.route);

    expect(items).toHaveLength(149);
    expect(missing).toEqual([]);
    expect(menuIconKeyForRoute("/product", "master")).toBe("package");
    expect(menuIconKeyForRoute("/productset", "master")).toBe("boxes");
    expect(menuIconKeyForRoute("/masterbrandscreen", "master")).toBe("badge");
    expect(menuIconKeyForRoute("/transaction/stocktransfer", "transaction")).toBe("truck");
    expect(menuIconKeyForRoute("/cashinginthedrawer", "finance")).toBe("money");
  });

  it("falls back to a category icon for unknown routes", () => {
    expect(menuIconKeyForRoute("/unknown", "report")).toBe("barChart");
  });
});
