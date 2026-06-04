# Data Transfer

MongoDB source and destination are configurable. Prefer explicit values for this tool:

- `MONGODB_SOURCE_URI`, `MONGODB_SOURCE_DB`
- `MONGODB_DESTINATION_URI`, `MONGODB_DESTINATION_DB`

If source is not set, it falls back to `MONGODB_UAT_URI` / `MONGODB_UAT_DB`.
If destination is not set, it falls back to the current environment selected by `MODE` / `BC_ENV`.

```
go run cmd/datatransfer/main.go --holding_code=32mCs7T8ZUyVeo6rwoncMkXdxrv --toholding_code=33UYr4vEDECjXsql4x3Lb4FdWfX --confirm
```
