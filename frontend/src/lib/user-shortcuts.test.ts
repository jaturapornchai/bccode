import { describe, expect, it } from "vitest";
import {
  addUserShortcut,
  clearUserShortcuts,
  moveUserShortcut,
  readUserShortcuts,
  removeUserShortcut,
  userShortcutsStorageKey,
  writeUserShortcuts,
} from "./user-shortcuts";

function createMockStorage(initial: Record<string, string> = {}) {
  const store = new Map<string, string>(Object.entries(initial));
  return {
    getItem(key: string) {
      return store.get(key) ?? null;
    },
    setItem(key: string, value: string) {
      store.set(key, value);
    },
    removeItem(key: string) {
      store.delete(key);
    },
    store,
  };
}

describe("user-shortcuts utility", () => {
  describe("userShortcutsStorageKey", () => {
    it("creates unique storage keys based on username and backend URL", () => {
      const key1 = userShortcutsStorageKey({ username: "somchai", backendUrl: "http://backend1:8080" });
      const key2 = userShortcutsStorageKey({ username: "somchai", backendUrl: "http://backend2:8080" });
      const key3 = userShortcutsStorageKey({ username: "somying", backendUrl: "http://backend1:8080" });

      expect(key1).toContain("bc_user_shortcuts_v1");
      expect(key1).not.toBe(key2);
      expect(key1).not.toBe(key3);
    });

    it("prefers profile email when available and case-insensitively normalizes login", () => {
      const keyA = userShortcutsStorageKey({
        username: "userA",
        profile: { email: "Test@Example.Com" },
        backendUrl: "http://srv:8080",
      });
      const keyB = userShortcutsStorageKey({
        username: "userB",
        profile: { email: "test@example.com" },
        backendUrl: "http://srv:8080",
      });

      expect(keyA).toBe(keyB);
    });

    it("handles null or undefined auth gracefully", () => {
      expect(userShortcutsStorageKey(null)).toContain("bc_user_shortcuts_v1");
      expect(userShortcutsStorageKey(undefined)).toContain("bc_user_shortcuts_v1");
    });
  });

  describe("readUserShortcuts", () => {
    it("returns null when no data is stored", () => {
      const storage = createMockStorage();
      expect(readUserShortcuts(storage, "non-existent")).toBeNull();
    });

    it("returns null for invalid JSON or non-array payloads", () => {
      const storage = createMockStorage({
        "bad-json": "{ broken",
        "bad-type": JSON.stringify({ items: ["product"] }),
      });
      expect(readUserShortcuts(storage, "bad-json")).toBeNull();
      expect(readUserShortcuts(storage, "bad-type")).toBeNull();
    });

    it("reads valid shortcut IDs and filters out duplicates and whitespace", () => {
      const storage = createMockStorage({
        "user-key": JSON.stringify(["product", "sale", "product", "  ", "purchase"]),
      });
      expect(readUserShortcuts(storage, "user-key")).toEqual(["product", "sale", "purchase"]);
    });

    it("filters out menu items that the user has no permission for", () => {
      const storage = createMockStorage({
        "user-key": JSON.stringify(["product", "secret-admin-menu", "sale"]),
      });
      const allowed = new Set(["product", "sale"]);
      expect(readUserShortcuts(storage, "user-key", allowed)).toEqual(["product", "sale"]);
    });
  });

  describe("writeUserShortcuts and clearUserShortcuts", () => {
    it("persists shortcuts to storage and deduplicates them", () => {
      const storage = createMockStorage();
      writeUserShortcuts(storage, "user-key", ["product", "sale", "product"]);

      expect(storage.getItem("user-key")).toBe(JSON.stringify(["product", "sale"]));
    });

    it("clears shortcuts from storage on clearUserShortcuts", () => {
      const storage = createMockStorage({ "user-key": JSON.stringify(["product"]) });
      clearUserShortcuts(storage, "user-key");

      expect(storage.getItem("user-key")).toBeNull();
    });
  });

  describe("addUserShortcut", () => {
    it("appends new menuId to list", () => {
      expect(addUserShortcut(["product"], "sale")).toEqual(["product", "sale"]);
    });

    it("does not add duplicates or empty strings", () => {
      expect(addUserShortcut(["product", "sale"], "product")).toEqual(["product", "sale"]);
      expect(addUserShortcut(["product"], "   ")).toEqual(["product"]);
    });
  });

  describe("removeUserShortcut", () => {
    it("removes specified menuId from list", () => {
      expect(removeUserShortcut(["product", "sale", "purchase"], "sale")).toEqual(["product", "purchase"]);
    });

    it("returns same array if item not found", () => {
      expect(removeUserShortcut(["product", "sale"], "non-existent")).toEqual(["product", "sale"]);
    });
  });

  describe("moveUserShortcut", () => {
    it("moves item from one index to another (move up / earlier)", () => {
      const list = ["product", "sale", "purchase"];
      // move "purchase" from index 2 to index 1
      expect(moveUserShortcut(list, 2, 1)).toEqual(["product", "purchase", "sale"]);
      // move "purchase" from index 2 to index 0
      expect(moveUserShortcut(list, 2, 0)).toEqual(["purchase", "product", "sale"]);
    });

    it("moves item from one index to another (move down / later)", () => {
      const list = ["product", "sale", "purchase"];
      // move "product" from index 0 to index 1
      expect(moveUserShortcut(list, 0, 1)).toEqual(["sale", "product", "purchase"]);
      // move "product" from index 0 to index 2
      expect(moveUserShortcut(list, 0, 2)).toEqual(["sale", "purchase", "product"]);
    });

    it("ignores invalid indices or identical from/to indices", () => {
      const list = ["product", "sale"];
      expect(moveUserShortcut(list, 0, 0)).toBe(list);
      expect(moveUserShortcut(list, -1, 1)).toBe(list);
      expect(moveUserShortcut(list, 0, 5)).toBe(list);
    });
  });
});
