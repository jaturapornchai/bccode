import type { AuthSession } from "./workspace-models";

export type UserShortcutsAuth = Pick<AuthSession, "backendUrl" | "username" | "profile"> | null | undefined;
export type ReadableStorage = Pick<Storage, "getItem">;
export type WritableStorage = Pick<Storage, "getItem" | "setItem">;
export type RemovableStorage = Pick<Storage, "removeItem">;

const STORAGE_PREFIX = "bc_user_shortcuts_v1";

function hashKeyPart(value: string): string {
  let hash = 2166136261;
  for (let index = 0; index < value.length; index += 1) {
    hash ^= value.charCodeAt(index);
    hash = Math.imul(hash, 16777619);
  }
  return (hash >>> 0).toString(36);
}

/** Generate a distinct storage key per user identity and backend URL. */
export function userShortcutsStorageKey(auth: UserShortcutsAuth): string {
  const login = auth?.profile?.email?.trim() || auth?.username?.trim() || "anonymous";
  const backend = auth?.backendUrl?.trim() || "default";
  return `${STORAGE_PREFIX}:${hashKeyPart(backend)}:${hashKeyPart(login.toLowerCase())}`;
}

/**
 * Reads user shortcuts from storage.
 * Returns `null` if the user has never customized their shortcuts (or data is invalid),
 * signaling that system default/smart fallback shortcuts should be used.
 */
export function readUserShortcuts(
  storage: ReadableStorage | null | undefined,
  key: string,
  allowedMenuIds?: Set<string>,
): string[] | null {
  if (!storage) return null;
  try {
    const raw = storage.getItem(key);
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return null;

    const seen = new Set<string>();
    const valid: string[] = [];
    for (const item of parsed) {
      if (typeof item === "string" && item.trim()) {
        const id = item.trim();
        if (!seen.has(id)) {
          seen.add(id);
          if (!allowedMenuIds || allowedMenuIds.has(id)) {
            valid.push(id);
          }
        }
      }
    }
    return valid;
  } catch {
    return null;
  }
}

/** Saves user shortcuts to storage. */
export function writeUserShortcuts(
  storage: WritableStorage | null | undefined,
  key: string,
  shortcutIds: string[],
): void {
  if (!storage) return;
  try {
    const unique = Array.from(new Set(shortcutIds.filter((id) => typeof id === "string" && id.trim())));
    storage.setItem(key, JSON.stringify(unique));
  } catch {
    // Fail silently if storage is blocked or full
  }
}

/** Clears user shortcuts from storage (reverting to system default recommendations). */
export function clearUserShortcuts(
  storage: RemovableStorage | null | undefined,
  key: string,
): void {
  if (!storage) return;
  try {
    storage.removeItem(key);
  } catch {
    // Fail silently
  }
}

/** Adds a menu ID to the shortcut list if not already present. */
export function addUserShortcut(currentIds: string[], menuId: string): string[] {
  const trimmed = menuId.trim();
  if (!trimmed || currentIds.includes(trimmed)) return currentIds;
  return [...currentIds, trimmed];
}

/** Removes a menu ID from the shortcut list. */
export function removeUserShortcut(currentIds: string[], menuId: string): string[] {
  const trimmed = menuId.trim();
  return currentIds.filter((id) => id !== trimmed);
}

/** Moves an item from one index to another in the shortcut list. */
export function moveUserShortcut(currentIds: string[], fromIndex: number, toIndex: number): string[] {
  if (
    fromIndex < 0 ||
    fromIndex >= currentIds.length ||
    toIndex < 0 ||
    toIndex >= currentIds.length ||
    fromIndex === toIndex
  ) {
    return currentIds;
  }

  const next = [...currentIds];
  const [removed] = next.splice(fromIndex, 1);
  next.splice(toIndex, 0, removed);
  return next;
}
