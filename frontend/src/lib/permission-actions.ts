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

// 2026-09-16: the labels above are the fallback — these keys read languages.tsv.
export const PERMISSION_ACTION_KEYS: Record<PermissionAction, string> = {
  create: "add",
  update: "edit",
  delete: "ss_delete",
};

export function actionEntry(screen: string, action: PermissionAction): string {
  return `${screen}:${action}`;
}

export function isActionEntry(entry: string): boolean {
  return entry.includes(":");
}

// Champ "สิทธิ์เฉพาะ" (กำหนดสิทธิ์การใช้งาน แท็บเฉพาะ): rights that are not a screen, stored and checked as the literal
// entry (backend report.PermissionPrint / PermissionExport). report:print = bcoption1 bCanPrintReport, report:export =
// BCOption13 bCanExportReportData (champ/champ/BC5Account.cpp:6177-6178,6213-6214)
export type SpecialPermission = "report:print" | "report:export";

export const SPECIAL_PERMISSION_LABELS: Record<SpecialPermission, { th: string; en: string }> = {
  "report:print": { th: "พิมพ์รายงานได้", en: "Can print reports" },
  "report:export": { th: "ส่งออกข้อมูลของรายงานได้", en: "Can export report data" },
};

// the labels above are the fallback — these keys read languages.tsv
export const specialPermissionTextKeys: Record<SpecialPermission, string> = {
  "report:print": "perm_special_report_print",
  "report:export": "perm_special_report_export",
};

export const SPECIAL_PERMISSIONS = Object.keys(SPECIAL_PERMISSION_LABELS) as SpecialPermission[];

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

// cancel is not a separate checkbox: like the backend (transaction/httpapi/http.go authorize), "<screen>:cancel" or the
// delete permission allows cancelling a document.
export type ScreenActions = Record<PermissionAction, boolean> & { cancel: boolean };

export const ALL_SCREEN_ACTIONS: ScreenActions = { create: true, update: true, delete: true, cancel: true };

export function screenActionsFor(permissions: readonly string[], screen: string): ScreenActions {
  const canDelete = hasScreenAction(permissions, screen, "delete");
  return {
    create: hasScreenAction(permissions, screen, "create"),
    update: hasScreenAction(permissions, screen, "update"),
    delete: canDelete,
    cancel: canDelete || permissions.includes(`${screen}:cancel`),
  };
}
