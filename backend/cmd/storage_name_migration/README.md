# Storage Name Migration

Migration history is recorded in `MIGRATION_HISTORY.md`.

## Objective

Rename existing MongoDB collections, PostgreSQL tables, and ClickHouse tables to lowercase `snake_case` names.

Example:

```text
shopUserAccessLogs -> shop_user_access_logs
productbarcode -> product_barcodes
docdetail -> doc_detail
```

## Workflow

1. Run dry-run first.
2. Review every `dry-run` and `conflict` line.
3. Backup the target database.
4. Run again with `--apply` only on a verified dev/staging copy.
5. Run application smoke tests after rename.

The tool skips a rename when the target name already exists. It does not merge data automatically.

## Config

```powershell
go run ./cmd/storage_name_migration `
  --target all `
  --mongo-uri "$env:MONGODB_DEV_URI" `
  --mongo-db "$env:MONGODB_DEV_DB" `
  --postgres-dsn "postgres://user:password@localhost:5432/bc_account?sslmode=disable" `
  --postgres-schema "public" `
  --clickhouse-url "http://localhost:8123" `
  --clickhouse-db "bc_account"
```

Run only one database:

```powershell
go run ./cmd/storage_name_migration --target mongo --mongo-uri "$env:MONGODB_DEV_URI" --mongo-db "$env:MONGODB_DEV_DB"
go run ./cmd/storage_name_migration --target postgres --postgres-dsn "postgres://user:password@localhost:5432/bc_account?sslmode=disable"
go run ./cmd/storage_name_migration --target clickhouse --clickhouse-url "http://localhost:8123" --clickhouse-db "bc_account"
```

Apply changes:

```powershell
go run ./cmd/storage_name_migration --target all --apply ...
```

## Dependencies

- MongoDB access with permission to run `renameCollection`.
- PostgreSQL access with permission to `ALTER TABLE ... RENAME TO`.
- ClickHouse HTTP endpoint with permission to `RENAME TABLE`.
- Go toolchain.

## Usage Example

Dry-run output:

```text
[mongo] dry-run shopUserAccessLogs -> shop_user_access_logs
[postgres] dry-run productbarcode -> product_barcodes
[clickhouse] conflict docdetail -> doc_detail (target table already exists)
```

## Limitation

- This tool renames object names only. It does not rename columns, indexes, views, foreign keys, materialized views, Kafka topics, backups, dashboards, or external SQL files.
- If both old and new names exist, the tool reports `conflict` and skips the rename. Merge or archive data manually first.
- Test against a copy of production data before running on real data.
