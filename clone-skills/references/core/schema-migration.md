# Schema Migration — Rules for DB Schema Changes

## Core Rule

When backend changes database schema (add/remove/rename columns, change types/nullable, add/remove tables/collections), **frontend dev must always be notified**.

## Why?

- Dart model may parse incorrectly (missing field -> null crash)
- `fromJson` without null handling will throw in production
- Frontend may send fields backend no longer accepts

## Migration Impact Report

Every schema change requires this summary:

```markdown
## Migration Impact Report

### Changes
- [Added/Removed/Changed] column `{name}` in table `{table}`
- Type: `{old_type}` -> `{new_type}`
- Nullable: `{old}` -> `{new}`

### Frontend Impact
- [ ] Dart model `{ModelName}` needs field `{name}` updated
- [ ] `fromJson()` must handle null/missing field
- [ ] Repository response parsing needs adjustment

### Backward Compatibility
- [ ] Does old frontend still work?
- [ ] Deploy frontend before or after backend?
```

## 3-Step Process

1. **Create Impact Report** -> notify frontend dev
2. **Deploy backend** -> verify API response still correct
3. **Deploy frontend** -> update Dart models to match

## Risk Assessment

| Action | Risk | Mitigation |
|--------|------|------------|
| Add column | Low | Old frontend still works (new field ignored) |
| Remove column | High | Frontend crashes if field missing -> fix frontend first |
| Rename column | High | Same as remove + add -> fix both sides |
| Change type | High | Parse error -> fix both sides |
| Change nullable | Medium | `fromJson` may throw -> add null checks |
