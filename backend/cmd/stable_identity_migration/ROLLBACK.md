# stableidentitymigration rollback notes

This migration is additive and safe to rerun. It backfills stable identity fields and creates MongoDB indexes.

## Task 1: `users.uid`

- Forward: backfill missing or empty `users.uid`, then create unique index `uxusersuid`.
- Rollback index only:
  ```js
  db.users.dropIndex("uxusersuid")
  ```
- Data rollback is not recommended. `uid` is the permanent user identity. Removing generated `uid` values can orphan `shopusers`, approval history, downstream sync, and active tokens.

## Task 2: `shopusers.useruid`

- Forward: backfill `shopusers.useruid` from `shopusers.username -> users.username -> users.uid`, then create index `ixshopusersholdingcodeuseruid`.
- Rollback index only:
  ```js
  db.shopusers.dropIndex("ixshopusersholdingcodeuseruid")
  ```
- If an older deployment expects the legacy index name, recreate the same keys with the old name:
  ```js
  db.shopusers.createIndex(
    { holdingcode: 1, useruid: 1 },
    { name: "ixshopusersholdingcodeuseruid" }
  )
  ```
- Data rollback is not recommended. During transition the app still dual-reads username, but deleting `useruid` removes rename-safe joins.

## Task 3: PO approval user identity

- Forward: backfill `poapprovalsettings.rules[].approvers[].approver_useruid` and `poapprovalstatus.history[].approver_useruid` from username/user code to `users.uid`.
- Rollback data only if a deployment must return to a schema that cannot tolerate extra fields:
  ```js
  db.poapprovalsettings.updateMany(
    {},
    { $unset: { "rules.$[].approvers.$[].approver_useruid": "" } }
  )
  db.poapprovalstatus.updateMany(
    {},
    { $unset: { "history.$[].approver_useruid": "" } }
  )
  ```
- Prefer leaving these fields in place because old code ignores unknown fields and the values are needed for stable joins.

## Verification

After rollback or forward migration, run dry-run/audit again:

```powershell
cd D:\bccode\backend
go run .\cmd\stableidentitymigration -config .\bootstrap.json
```
