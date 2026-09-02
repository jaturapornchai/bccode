// Screen permissions are stored as flat strings on a role permission record:
//   "<screen>"          = เข้า  (may open the screen)
//   "<screen>:create"   = เพิ่ม
//   "<screen>:update"   = แก้ไข
//   "<screen>:delete"   = ลบ
//   "*"                 = everything (ADMIN/OWNER default from the backend)
// Legacy records that only list "<screen>" keep working as entry-only.

export type PermissionAction = "create" | "update" | "delete";

export const PERMISSION_ACTIONS: readonly PermissionAction[] = ["create", "update", "delete"];

export const PERMISSION_ACTION_LABELS: Record<PermissionAction, { th: string; en: string }> = {
  create: { th: "เพิ่ม", en: "Create" },
  update: { th: "แก้ไข", en: "Edit" },
  delete: { th: "ลบ", en: "Delete" },
};

export function actionEntry(screen: string, action: PermissionAction): string {
  return `${screen}:${action}`;
}

export function isActionEntry(entry: string): boolean {
  return entry.includes(":");
}

export function hasScreenAccess(permissions: readonly string[], screen: string): boolean {
  return permissions.includes("*") || permissions.includes(screen);
}

export function hasScreenAction(
  permissions: readonly string[],
  screen: string,
  action: PermissionAction,
): boolean {
  return permissions.includes("*") || permissions.includes(actionEntry(screen, action));
}

export type ScreenActions = Record<PermissionAction, boolean>;

export const ALL_SCREEN_ACTIONS: ScreenActions = { create: true, update: true, delete: true };

export function screenActionsFor(permissions: readonly string[], screen: string): ScreenActions {
  return {
    create: hasScreenAction(permissions, screen, "create"),
    update: hasScreenAction(permissions, screen, "update"),
    delete: hasScreenAction(permissions, screen, "delete"),
  };
}
