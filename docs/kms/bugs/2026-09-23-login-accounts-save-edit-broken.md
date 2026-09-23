---
date: 2026-09-23
severity: high
component: [frontend, backend]
tags: [bc-account, postgres, settings, permission]
fixed: true
---

# Symptom

Screen: ตั้งค่าระบบและการเข้าถึง › บัญชีเข้าระบบ (`/user`, `PUT/POST/DELETE /holding/permission`). This is the only place to add or remove a group admin since 2026-09-23.

1. **Adding a new user fails.** Pressing save shows the toast "บันทึกไม่สำเร็จ: json: cannot unmarshal object into Go struct field .requestAlias.permissionsets of type []string". The BFF POST returns 400.
2. **Editing an existing user fails** (including the Demo user). Pressing "แก้ไข" shows the toast "find failed". `GET /holding/permission/<useruid>` returns 400.

## Root Cause

1. The `permissionsets` field (สิทธิ์การใช้งานเพิ่มเติม) is a `json` field. Its default value was `{}` because `permissionsets` was missing from the list of array-valued JSON keys in `frontend/src/app/system-settings/system-settings-screen.tsx`. That list appears in three places: the empty-form default, record loading, and `parseJsonField`. A new user with no permission sets selected sent `permissionsets: {}`, which the backend `UserRoleRequest.PermissionSets []string` could not unmarshal.
2. `ShopUserService.InfoShopByUser` looked up the member by username first. It only fell back to useruid when it got back an **empty struct with no error**, which is how the old MongoDB repo behaved. The PostgreSQL repo (`ShopUserPostgresRepository.findMember`) returns `ErrShopUserNotFound` instead. The screen addresses members by useruid, so the lookup always failed. `DeleteUserPermissionShop` was unaffected because it already used `resolveShopUser`, which handles the error correctly.

## Fix

- Frontend: `permissionsets` added to the array-JSON key list in all three places. Default is now `[]`, and a record with `null` loads as `[]`.
- Backend: `InfoShopByUser` now uses `resolveShopUser`, which tries username, then useruid, then lowercased useruid. Profile lookup uses the username of the resolved member.

## Lesson (applies to other code too)

When porting a repo from MongoDB to PostgreSQL, "not found" changes from "empty struct + nil error" to `ErrShopUserNotFound`. Callers that check for empty values instead of checking the error lose their fallback path.

A 2026-09-23 grep found `FindByHoldingCodeAndUsername` used in only 2 places: `resolveShopUser` (correct) and `requireHoldingManager` (denies on error, which is correct).

## Regression Test

- `TestInfoShopByUserResolvesUserUID` (`backend/internal/shop/shopuser_service_test.go`): fails against the old code, passes against the new code.
- Local UAT with the Demo button (rungrueng), checking PostgreSQL `bcai_projection.holding_members` after each step:
  1. Create `uat_admin_0923` as ADMIN → new row, `permission_sets=[]`.
  2. Edit to USER → same row id `7083f12f…` changes role.
  3. Delete → row removed; the demo/OWNER row is untouched (count 2→1).
  4. The test user's `users` row was deleted by id afterwards.
