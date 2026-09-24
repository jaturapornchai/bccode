---
date: 2026-09-24
severity: high
component: [backend]
tags: [bc-account, go, postgres, auth, security]
fixed: true
---

# Symptom

ADMIN ที่ถูกจำกัดขอบเขตไว้ที่บริษัท A สร้าง API/MCP token ให้บริษัท B ของ Holding เดียวกันได้ แล้วใช้ token นั้นอ่าน/เขียน GL ของบริษัท B ได้ทั้งหมด (manager ได้ `Permissions["*"]`) — รวมถึง token ที่ออกไว้ก่อน แม้ภายหลัง OWNER ลดขอบเขต ADMIN คนนั้นแล้ว ก็ยังใช้กับบริษัทเดิมได้จนหมดอายุ (สูงสุด 1 ปี)

## Root Cause

- `authorizeHolding` (`backend/internal/mcptoken/token.go`) ตรวจแค่ role OWNER/ADMIN ไม่อ่าน `access_scopes`
- `validateCompanies` รับทุกบริษัทที่ active ใน Holding โดยไม่เทียบกับขอบเขตของผู้ออก
- GL `resolveScope` (`backend/internal/generalledger/httpapi/scope.go`) ข้าม `sessionScopeAllowed` สำหรับคำขอ token โดยเชื่อ allow-list ของ token — แต่ allow-list นั้นไม่เคยถูกจำกัดด้วยขอบเขตผู้ออก

## Fix

- `authorizeHolding` ใช้ `access.FindActiveHoldingManager` (ขยายขอบเขตทั้งกลุ่มกิจการเป็นทุกบริษัทที่ active ตอนอ่าน) แล้วคืน membership
- `withinScope` = `ScopesAllowCompanySelection` (token ใหม่ทั้งบริษัท) / `ScopesAllowBranchSelection` (token เดิมที่ผูกสาขา) — กติกาเดียวกับการเลือกบริษัท/สาขาของ browser session
- ตอนออก: `create` ปฏิเสธบริษัทนอกขอบเขต → 400 key `mcp_err_company_outside_scope` (ไม่ใช่ 403 เพราะจอ token ตีความ 403 ว่าหมดสิทธิ์ผู้ดูแลแล้วล้างฟอร์ม); `GET /mcp-tokens/companies` แสดงเฉพาะบริษัทที่ผู้ออกเลือกได้
- ตอนใช้ (ตัดสินใจให้ตรวจซ้ำ เพราะ role ของผู้ออกก็ตรวจทุกครั้งอยู่แล้ว และ OWNER ลดขอบเขต ADMIN ได้ภายหลัง): `authenticateAudience` ตัดบริษัทที่ผู้ออกเข้าไม่ได้ ณ ตอนนั้นทิ้ง ไม่เหลือเลย = ปฏิเสธ token; ไม่แก้ grant ที่เก็บไว้ (ขยายขอบเขตคืนแล้ว token กลับมาใช้ได้)
- GL `resolveScope` เลิกข้าม `sessionScopeAllowed` สำหรับ token — ตรวจขอบเขตผู้ออกทุกคำขอเหมือน session
- ADR `docs/kms/decisions/2026-09-24-holding-wide-access-scope.md` (Consequences)

## Regression Test

- unit: `TestWithinScope`, `TestCompanyOutsideScopeMessageFollowsLanguage` (`backend/internal/mcptoken/scope_test.go`); `TestTokenScopeBoundedByIssuer` (`backend/internal/generalledger/httpapi/company_scope_test.go`)
- integration (`-tags integration`, `MCP_TOKEN_TEST_DSN` / `GL_AUTH_TEST_DSN` → postgres:18-alpine ชั่วคราว): `TestManagementHTTPPostgres` (ADMIN จำกัดบริษัท/สาขา/ทั้งกลุ่ม), `TestPostgresAuthenticationBoundedByIssuerScope` (ลดขอบเขตผู้ออกหลังออก token), `TestPostgresScopeRevocation` (คำขอ token ถูกขอบเขตผู้ออกปฏิเสธ)
- mutation check: ให้ `withinScope` คืน true เสมอ → integration test ทั้งสองตัว fail ตามคาด
