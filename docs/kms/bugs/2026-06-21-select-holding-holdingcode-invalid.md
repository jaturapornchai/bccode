---
tags: [bc-account, go, mongodb, auth]
date: 2026-06-21
---

# select-holding wrongly returns "holdingcode invalid" for valid members

## Symptom
After creating a holding (`create-holding` → `{"success":true}`), selecting it
failed: `POST /select-holding {holdingcode}` → `{"success":false,"message":"holdingcode invalid"}`.
This cascaded: token never got a `holdingcode` → every `/goapi/*` call → `{"message":"Shop not selected"}`,
and all product-master CRUD broke. The holding-code format was valid (regex
`^[a-z][a-z0-9]{2,29}$` matched), so `utils/holding_code.go` was NOT the cause.

## Root cause
Two facts combine:
1. `PersisterMongo.FindOne` (`pkg/microservice/persister_mongo.go`) **swallows** the
   `mongo: no documents in result` error → returns `nil` error with a zero-valued struct.
2. `findShopUser` resolves membership by `useruid` first, then is supposed to fall back to
   the email `username`. But the fallback fired only on `err != nil`:
   ```go
   if strings.TrimSpace(userUID) == "" || err != nil { ... username lookup ... }
   ```
   When the token's `uid` ≠ the `useruid` stored on the membership, the uid query
   "succeeds empty" (nil err, empty struct) → fallback skipped → `AccessShop` sees
   `shopUser.ID == primitive.NilObjectID` → `"holdingcode invalid"`.

The uid drifts whenever a membership was created with a different/older uid than the
current token: a **by-email membership** (holding-admin add) created before that person
ever logged in keeps an empty/old `useruid`; a **stale token** minted before a data wipe
carries an old uid. `create-holding` stores `useruid` from the DB `users` record, so a
token whose uid no longer matches the DB user also triggers it.

## Fix
Add the empty-result case to the fallback condition in all 3 sibling call sites in
`backend/internal/authentication/services/authentication_service.go`:
```go
if strings.TrimSpace(userUID) == "" || err != nil || shopUser.ID == primitive.NilObjectID {
    shopUser, err = svc.shopUserRepo.FindByHoldingCodeAndUsername(ctx, holdingCode, username)
}
```
- `findShopUser` (~391) — used by `AccessShop`/select-holding
- `processUserLogin` (~323) — login-with-holdingcode in one shot
- `UpdateFavoriteShop` (~913)

Mirrors the same defensive `|| shopUser.Username == ""` guard already present in
`shop/shopuser_repository.go:FindByHoldingCodeAndUsername`.

## Regression test (live, .202)
Stale token (token uid `3EjpEvCK…` ≠ DB uid `3FRtbEfip…`, shopuser.useruid `3FRtbEfip…`):
- `create-holding bctestflow` → `{"success":true}`
- `select-holding bctestflow` → `{"success":true}` ✅ (was `holdingcode invalid`)
- token `holdingcode` = `bctestflow` → `/goapi/genpdf/history` returns a handler-level 400
  (param required), i.e. it passed the "Shop not selected" auth middleware.

## Prevention (skills + audit)
Encoded the root lesson into the project skills so it is not repeated:
- `.agents/skills/go-expert/SKILL.md` → "Mongo not-found semantics — `FindOne` SWALLOWS,
  `FindByID` does NOT": existence/fallback/duplicate/not-found decisions must test the decoded
  result for emptiness, never `err != nil` alone.
- `.agents/skills/bc-account-expert/SKILL.md` → "Membership / shop-access resolution": resolve
  shopuser by `useruid` then fall back to `username` on an EMPTY result (uid drift).

A 17-agent audit over all 232 `FindOne` call sites (map → adversarial-verify) found **2 more
instances of this exact class**, both in the report-execute service — `reportquerym/...:304`
(HIGH, `*findDoc.Fields` nil-ptr panic) and `reportqueryc/...:263` (LOW, empty SQL → ClickHouse
error). Both guarded with a `findDoc.Code == ""` not-found check. **Caveat:** those handlers are
not registered in the deployed mainapi (root `main.go`/8888) — only in `cmd/app`/8080 — so the
panic is not reachable in the running build; the guards are defense-in-depth.

## Related
- [[2026-06-21-create-holding-illegaloperation-transaction]] (replica-set fix that unblocked create-holding first)
