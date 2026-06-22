---
name: go-api-handler
description: Create Go API handlers following BC Account patterns.
---

- **Route Framework**: Use Gin. Route handlers must accept `*gin.Context`.
- **Response Format**:
  - Success: `{"success": true, "data": ...}`
  - Error: `{"success": false, "error": {"code": "...", "message": "..."}}`
- **Tenancy Scope**: Authorize `tenant_id` from JWT/session/workspace and map to storage physical key.
- **Validation**: Validate input structs with tags. Handle exceptions with custom types. Write unit tests.
- **Image Upload = PNG/JPG only (set 2026-06-21)**: The image upload handler (`internal/goapi/handlers/image_r2.go ImageUploadHandler`, `POST /goapi/image/upload`) is the authoritative guard that S3/MinIO only stores PNG/JPG. Keep its extension allowlist `{.jpg,.jpeg,.png}` AND the `http.DetectContentType` check (`image/jpeg`/`image/png` only) — the content-type check rejects a webp/gif/bmp renamed `.jpg`. Do not re-add `.webp`/`.gif`/`.bmp`. The pipeline stores the uploaded bytes as-is (no server-side re-encode); the client sends PNG/JPG and a small PNG/JPG thumbnail.
- **Mongo not-found semantics (`FindOne` swallows)**: `PersisterMongo.FindOne` returns a **nil error + zero-valued struct** on no-match (it swallows `mongo: no documents in result`), unlike `FindByID` which returns `mongo.ErrNoDocuments`. Any existence / fallback / duplicate / "not found" decision built on a `FindOne`-based repo method MUST test the decoded result for emptiness (`ID == primitive.NilObjectID`, `Code == ""`, …), never `err != nil` alone — see go-expert for the canonical fix (select-holding `holdingcode invalid` 2026-06-21).
