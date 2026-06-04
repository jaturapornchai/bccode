# stable_identity_migration rollback notes

This migration is additive and safe to rerun. It backfills stable identity fields and creates MongoDB indexes.

## Task 1: `users.uid`

- Forward: backfill missing or empty `users.uid`, then create unique index `ux_users_uid`.
- Rollback index only:
  ```js
  db.users.dropIndex("ux_users_uid")
  ```
- Data rollback is not recommended. `uid` is the permanent user identity. Removing generated `uid` values can orphan `shopUsers`, approval history, downstream sync, and active tokens.

## Task 2: `shopUsers.user_uid`

- Forward: backfill `shopUsers.user_uid` from `shopUsers.username -> users.username -> users.uid`, then create index `ix_shopUsers_holding_code_user_uid`.
- Rollback index only:
  ```js
  db.shopUsers.dropIndex("ix_shopUsers_holding_code_user_uid")
  ```
- If an older deployment expects the legacy index name, recreate the same keys with the old name:
  ```js
  db.shopUsers.createIndex(
    { holding_code: 1, user_uid: 1 },
    { name: "ix_shop_users_holding_code_user_uid" }
  )
  ```
- Data rollback is not recommended. During transition the app still dual-reads username, but deleting `user_uid` removes rename-safe joins.

## Task 3: PO approval user identity

- Forward: backfill `po_approval_settings.rules[].approvers[].approver_user_uid` and `po_approval_status.history[].approver_user_uid` from username/user code to `users.uid`.
- Rollback data only if a deployment must return to a schema that cannot tolerate extra fields:
  ```js
  db.po_approval_settings.updateMany(
    {},
    { $unset: { "rules.$[].approvers.$[].approver_user_uid": "" } }
  )
  db.po_approval_status.updateMany(
    {},
    { $unset: { "history.$[].approver_user_uid": "" } }
  )
  ```
- Prefer leaving these fields in place because old code ignores unknown fields and the values are needed for stable joins.

## Verification

After rollback or forward migration, run dry-run/audit again:

```powershell
cd D:\bccode\backend
go run .\cmd\stable_identity_migration -config .\bootstrap.json
```
