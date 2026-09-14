---
date: 2026-09-12
status: deployed
tags: [bc-account, deployment, ui, baseline-search, debounce, clean-icon, general-ledger]
---

# Deploy แถบค้นหาหลัก Baseline Toolbar (Auto Search 2s & Clean Icon) สู่ Production

## วัตถุประสงค์และผลการปล่อยระบบ

ปล่อยรีลีส `r20260912-search-auto-1` สู่ Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) วันที่ 12 กันยายน 2026 ตามคำสั่งลุงจืด โดยครอบคลุม:
1. การค้นหาอัตโนมัติ (Auto Search) หน่วงเวลา 2 วินาที (2000ms debounce) หลังผู้ใช้หยุดพิมพ์
2. การยกเลิกตัวนับเวลาและสั่งค้นหาทันทีเมื่อกดปุ่ม "ค้นหา" หรือเคาะ Enter
3. ปุ่มไอคอน Clean (`[ ✕ ]`) ทางขวาของกล่องข้อความ ปรากฏเมื่อมีข้อความ กดคลิกหรือ Esc เพื่อล้างคำค้นทันทีพร้อมคืนโฟกัส
4. ประยุกต์ใช้กับหน้าจอข้อมูลหลัก GL, สมุดรายวัน และระบบออกแบบงบการเงิน

## Production Artifacts & Release Details

| บริการ | Image |
|---|---|
| mainapi / worker | `bcai-account-mainapi:r20260912-search-auto-1` |
| frontend | `bcai-account-frontend:r20260912-search-auto-1` |

- **Release Directory**: `/opt/bcai-account/releases/r20260912-search-auto-1`
- **Images Archive**: `images.tar` (597,849,600 bytes)
- **Preflight Backups**:
  - `mongo.archive.gz`: 72,133 bytes (SHA-256: `abbba1cf1cf033f5692fa2285983a666d610dc0530f2bc291598eb1cd8dd5daa`)
  - `postgres-all.sql`: 723,955 bytes (SHA-256: `75e628f55ea55c7d44480d0723f614bb4e569d5c7ee1c4f8cc858ce964a9cf4a`)
  - `runtime-config.tar.gz`: 7,264 bytes (SHA-256: `8645b118c7c40feaa1f5893629f2f7be8be09f677f81e94e636d66b59c1c5e0d`)
  - `release.env.before`: บันทึกสถานะก่อน deploy ไว้อย่างปลอดภัย
- **Rollback Procedure**: คืนค่า `/etc/bcai-account/release.env` จาก `release.env.before` แล้วสั่ง `docker compose up -d --no-deps mainapi worker frontend`

## ผลการทดสอบบน Production จริง (Live Verification)

- **HTTP Status**:
  - `https://account.bcaicloud.com/`: 200 OK
  - `https://account.bcaicloud.com/api/gl/accounts`: 401 Unauthorized (Auth guard ทำงานถูกต้อง)
  - `http://127.0.0.1:8888/healthz`: 200 OK
- **Live Chunk Verification**:
  - ตรวจพบโค้ด `useDebouncedSearch` และ `ล้างคำค้นหา` ใน Production Chunk `/_next/static/chunks/2-oz0lfgtc97p.js`
- **Smoke Tests**:
  - Demo Login: สำเร็จ (`success: true`)
  - Select Holding: สำเร็จ (`holdingcode: 'demo'`)
  - Query Accounts: สำเร็จ (`total: 38`)
  - Query Journals: สำเร็จ (`total: 18`)
  - Query Trial Balance: สำเร็จ (`rows: 31`)
