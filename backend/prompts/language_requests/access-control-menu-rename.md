# Access Control Menu Rename — STATUS: KEYS LANDED IN TSV

UX refactor for the "การเข้าถึง" sub-menu (system settings group `company-system`). Goal: remove repeated word "สิทธิ์", reorder by setup workflow, use a folder name that covers approval workflow too, and keep menu labels aligned with the actual data model (no "Role" entity exists — system uses Permission Codes + Approval Codes + employee_permissions binding).

## Keys (already in `assets/language/languages.tsv` lines 2258-2262)

| key | Thai | English | replaces (old key still in TSV for legacy callers) |
|---|---|---|---|
| `menu_users_permissions` | ผู้ใช้และสิทธิ์ | Users & Permissions | folder (was `access_control` "การเข้าถึง" / "Access") |
| `menu_permission_define` | นิยามสิทธิ์ | Permission Definition | item (was `permission_definition` "กำหนดสิทธิ์") |
| `menu_approval` | การอนุมัติ | Approvals | item (was `approval_setting` "สิทธิ์การอนุมัติ") |
| `menu_user` | ผู้ใช้งาน | Users | item (was `user` "User" — switched to plural EN) |
| `menu_assign_permission` | มอบหมายสิทธิ์ | Assign Permissions | item (was `permission_link` "ผูกสิทธิ์") |

Note: An earlier draft used key `menu_role_permission` "บทบาทและสิทธิ์" — REVERTED because the backend has no Role entity. Replaced with `menu_permission_define` "นิยามสิทธิ์" to match the `permission_definitions` collection.

## screen/module
- folder `access-control` + items `permission-definition`, `approval-setting`, `user`, `permission-link` under section `settings` > group `company-system`.

## reason / context
- Original labels used "สิทธิ์" three times in four items → high cognitive load, hard to scan.
- Original folder name "การเข้าถึง" was too narrow (the group also contains approval workflow + user management).
- Original order placed `ผู้ใช้งาน` first even though setup workflow is: define roles → define approvals → create users → assign permissions to users. New order matches that workflow.
- Old folder/item icons reused the `UserPlus` (user-add) glyph for both heading and first child, making parent vs child visually identical. Folder now uses `ShieldCheck`; `/user` now uses `UserRound`.

## Existing keys still in TSV (do NOT delete — may be reused elsewhere)
- `access_control` — still used by other code paths; only frontend folder migrated away from it.
- `permission_definition`, `approval_setting`, `permission_link`, `user`, `approval_permission` — still in TSV as business terms (permission API, audit log, etc.).

## caller file path
- `frontend/src/lib/menu-data.ts` lines 305-308 (item declarations)
- `frontend/src/app/menu/main-menu-screen.tsx` line 126 (folder definition)

## verification
1. UI loads with TH/EN fallback while keys are missing from backend.
2. After batch translate, switching language reflects the new labels.
3. Existing `permission_definition` / `approval_setting` / `permission_link` keys remain intact for other callers.
