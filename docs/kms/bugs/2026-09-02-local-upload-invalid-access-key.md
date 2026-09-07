---
date: 2026-09-02
severity: medium
component: [backend, infra]
tags: [bc-account, minio, upload, bootstrap]
fixed: true
---

# Symptom

local: upload PNG (โลโก้สาขา) → "Failed to upload image to storage"; รูปโปรไฟล์พนักงาน thumbnail ไม่ขึ้น; `docker logs mainapi` เต็มไปด้วย `S3 … 403 InvalidAccessKeyId` — ทั้งที่ `docker exec mainapi env` แสดง `S3_ACCESS_KEY_ID` ตรงกับ `storage.local.env` และ `mc` ด้วย creds เดียวกันใช้ได้

## Root Cause
`internal/goapi/setupconfig/loader.go` `applyBootstrapSection` ทำ `os.Setenv` จาก `bootstrap.json` (local mount `bootstrap.local.json`) ตอน start → `s3accesskeyid: minioadmin`, `s3bucketname: app-images` ทับ env ของ compose; `GetR2Client` cache client ตัวนี้ตลอดอายุ process. `/proc/1/environ` เป็นค่าตอน spawn จึงดูเหมือนถูก

## Fix
แก้ `bootstrap.local.json` (gitignored) ให้ s3 key/secret/bucket ตรง `storage.local.env` แล้ว `docker restart mainapi` — log ต้องขึ้น `Object storage client initialized (bucket: bcai-account …)`; ไม่แตะ code

## Regression Test
1. `docker logs mainapi | grep "Object storage client initialized"` → bucket ตรง `S3_BUCKET_NAME`
2. หน้าเว็บ login แล้ว POST `/api/upload/image` (category=branch, PNG) → 200 + key `bc001/branch/…png`
3. `mc ls t/bcai-account/bc001/branch/` เห็น `.png` + `.png.thumb.webp`; GET `/goapi/s3/file/<key>.thumb.webp` = 200 `image/webp`
