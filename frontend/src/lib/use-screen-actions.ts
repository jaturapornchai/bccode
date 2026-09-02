"use client";

import { useEffect, useState } from "react";
import { authFetch } from "@/lib/client-auth-session";
import { flattenMenuItems } from "@/lib/menu-data";
import { ALL_SCREEN_ACTIONS, screenActionsFor, type ScreenActions } from "@/lib/permission-actions";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";

type RolePermissionPayload = {
  data?: unknown;
};

/** First record from `{ data: {...} }` or `{ data: [...] }` (same shapes the main menu reads). */
function firstRecord(payload: unknown): Record<string, unknown> | null {
  const data = (payload as RolePermissionPayload | null)?.data;
  const record = Array.isArray(data) ? data[0] : data;
  return record && typeof record === "object" ? (record as Record<string, unknown>) : null;
}

/**
 * เพิ่ม/แก้ไข/ลบ ที่ผู้ใช้ปัจจุบันทำได้บนจอ `route` ตาม role permission ของ holding.
 * - จอที่ไม่อยู่ในเมนู (ไม่มีรหัสสิทธิ์) หรือโหลดสิทธิ์ไม่ได้ → ไม่ล็อกปุ่ม (การเข้าจอถูก
 *   คุมที่เมนูหลักอยู่แล้ว; การบังคับจริงเป็นหน้าที่ backend)
 * - มี record สิทธิ์แต่ไม่ได้ติ๊ก action → ปุ่มนั้นซ่อน
 */
export function useScreenActions(
  auth: AuthSession | null,
  workspace: WorkspaceSession | null,
  route: string,
): ScreenActions {
  const [actions, setActions] = useState<ScreenActions>(ALL_SCREEN_ACTIONS);
  const holdingcode = workspace?.shop?.holdingcode ?? "";
  const token = auth?.token ?? "";

  useEffect(() => {
    const screen = flattenMenuItems().find((item) => item.route === route)?.id;
    if (!screen || !token || !holdingcode) {
      setActions(ALL_SCREEN_ACTIONS);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const response = await authFetch(
          `/api/system-settings/permissiongroup/me?holdingcode=${encodeURIComponent(holdingcode)}`,
          { headers: { Authorization: `Bearer ${token}` }, cache: "no-store" },
        );
        if (!response.ok) return;
        const record = firstRecord((await response.json()) as unknown);
        if (cancelled || !record || record.isactive === false) return;
        const permissions = Array.isArray(record.permissions)
          ? record.permissions.filter((item): item is string => typeof item === "string")
          : [];
        setActions(screenActionsFor(permissions, screen));
      } catch {
        // keep the permissive default; entry is already gated by the main menu
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [holdingcode, route, token]);

  return actions;
}
