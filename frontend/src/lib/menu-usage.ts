import type { MenuItem } from "./menu-data";
import type { AuthSession } from "./workspace-models";

export type MenuUsageRecord = {
  count: number;
  lastUsedAt: number;
};

export type MenuUsageMap = Record<string, MenuUsageRecord>;
export type FrequentMenuEntry = {
  count: number;
  item: MenuItem;
  percent: number;
};

type MenuUsageAuth = Pick<AuthSession, "backendUrl" | "method" | "profile" | "username"> | null | undefined;
type ReadableStorage = Pick<Storage, "getItem">;
type WritableStorage = Pick<Storage, "getItem" | "setItem">;

const STORAGE_PREFIX = "bc_menu_usage_v1";

export function menuUsageStorageKey(auth: MenuUsageAuth): string {
  const login = auth?.profile?.email?.trim() || auth?.username?.trim() || "anonymous";
  const backend = auth?.backendUrl?.trim() || "default";
  const method = auth?.method?.trim() || "password";
  return `${STORAGE_PREFIX}:${hashKeyPart(backend)}:${hashKeyPart(method)}:${hashKeyPart(login.toLowerCase())}`;
}

export function readMenuUsage(storage: ReadableStorage, key: string): MenuUsageMap {
  try {
    return normalizeUsage(JSON.parse(storage.getItem(key) ?? "{}"));
  } catch {
    return {};
  }
}

export function recordMenuUsage(storage: WritableStorage, key: string, menuId: string, usedAt = Date.now()): MenuUsageMap {
  const current = readMenuUsage(storage, key);
  const previous = current[menuId];
  const next: MenuUsageMap = {
    ...current,
    [menuId]: {
      count: (previous?.count ?? 0) + 1,
      lastUsedAt: usedAt,
    },
  };

  try {
    storage.setItem(key, JSON.stringify(next));
  } catch {
    // Keep the in-memory result so the current session still updates if localStorage is full or blocked.
  }

  return next;
}

export function getFrequentMenuItems(items: MenuItem[], usage: MenuUsageMap, limit = 20): MenuItem[] {
  return getFrequentMenuEntries(items, usage, limit).map((entry) => entry.item);
}

export function getFrequentMenuEntries(items: MenuItem[], usage: MenuUsageMap, limit = 20): FrequentMenuEntry[] {
  const originalIndex = new Map(items.map((item, index) => [item.id, index]));
  const totalCount = Object.values(usage).reduce((total, record) => total + Math.max(0, record.count), 0);
  return items
    .filter((item) => (usage[item.id]?.count ?? 0) > 0)
    .sort((left, right) => {
      const leftUsage = usage[left.id];
      const rightUsage = usage[right.id];
      const countDiff = (rightUsage?.count ?? 0) - (leftUsage?.count ?? 0);
      if (countDiff !== 0) return countDiff;
      const lastUsedDiff = (rightUsage?.lastUsedAt ?? 0) - (leftUsage?.lastUsedAt ?? 0);
      if (lastUsedDiff !== 0) return lastUsedDiff;
      return (originalIndex.get(left.id) ?? 0) - (originalIndex.get(right.id) ?? 0);
    })
    .slice(0, limit)
    .map((item) => {
      const count = usage[item.id]?.count ?? 0;
      return {
        count,
        item,
        percent: totalCount > 0 ? Math.round((count / totalCount) * 100) : 0,
      };
    });
}

function normalizeUsage(value: unknown): MenuUsageMap {
  if (!isRecord(value)) return {};
  return Object.fromEntries(
    Object.entries(value).flatMap(([menuId, record]) => {
      if (!isRecord(record)) return [];
      const count = Number(record.count);
      const lastUsedAt = Number(record.lastUsedAt);
      if (!menuId || !Number.isFinite(count) || count <= 0 || !Number.isFinite(lastUsedAt)) return [];
      return [[menuId, { count, lastUsedAt } satisfies MenuUsageRecord]];
    }),
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hashKeyPart(value: string): string {
  let hash = 2166136261;
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return (hash >>> 0).toString(36);
}
