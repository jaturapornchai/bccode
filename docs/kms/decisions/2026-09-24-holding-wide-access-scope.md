---
date: 2026-09-24
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, go, auth, tenancy, security]
---

# ขอบเขต "ใช้ได้ทั้งกลุ่มกิจการ" ครอบคลุมบริษัท/สาขาที่เพิ่มภายหลัง — ขยายตอนอ่าน ไม่ขยายตอนบันทึก

## Context

- จอ ตั้งค่าระบบ › ผู้ใช้งาน มีช่อง "ใช้ได้ทั้งกลุ่มกิจการ" (`scopetype: holding`) แต่ backend เดิม (`hydrateAccessScopes`) แตกกฎนี้ตอนบันทึกเป็นกฎรายบริษัทเฉพาะบริษัทที่มีอยู่ขณะนั้น → บริษัทที่สร้างทีหลังผู้ใช้เข้าไม่ได้ และผู้อ่านทุกตัว fail closed กับกฎ holding ที่เก็บไว้
- OWNER/ADMIN ที่ไม่มี scope ถูกแสดงบนจอเป็นรายชื่อบริษัท ณ วันที่เปิด (snapshot) — กดบันทึกซ้ำก็กลายเป็นรายการตายตัว เสียความหมาย "ทั้งกลุ่ม"
- normalizer ฝั่ง frontend (BFF + editor) แปลงกฎที่ไม่มี `businesscode` หรือ `scopetype` ไม่รู้จักเป็น holding — เมื่อ holding กลายเป็นสิทธิ์กว้างที่ขยายเองได้ นี่คือช่องยกระดับสิทธิ์ (fail open)

## Decision

ลุงจืดตัดสินใจ: สมาชิกที่มี `[{scopetype:"holding"}]` ใช้ได้ **ทุกบริษัทและสาขาที่ active** ของ holding รวมที่สร้างภายหลัง

1. **เก็บ**: `hydrateAccessScopes` (`backend/internal/shop/scopes_postgres.go`) ตรวจทุกกฎ — ต้องเป็น holding/company/branch และบริษัท/สาขาต้องอยู่ใน holding นี้ ไม่งั้น 400 `VALIDATION_FAILED`; ถ้ามีกฎ holding เก็บเหลือกฎเดียว `[{"scopetype":"holding"}]`
2. **อ่าน**: `FindActiveMembership` (`backend/internal/organization/access/membership.go`) ขยายกฎ holding (และ OWNER/ADMIN ที่ไม่มี scope) เป็นทุกบริษัทที่ active ณ เวลาอ่าน; `HasHoldingScope` + `ScopesAllowCompanySelection/BranchSelection` (`backend/internal/authentication/models/user.go`) ให้ `AccessShop` และ GL `sessionScopeAllowed` ผ่าน โดยผู้เรียกต้องตรวจว่าบริษัท/สาขา active ก่อน (`ResolveCompanyUID` เพิ่ม `is_active = true`)
3. **ให้สิทธิ์**: `checkScopeGrant` (`backend/internal/shop/shopuser_service.go`) — OWNER ให้อะไรก็ได้; ADMIN ให้ holding ได้เมื่อตัวเองเป็นทั้งกลุ่มกิจการ (เก็บ holding หรือไม่มี scope) เท่านั้น; ADMIN ที่จำกัดบริษัทให้ได้เฉพาะบริษัท/สาขาในขอบเขตตัวเอง รวมตอนแก้ตัวเอง และสร้าง ADMIN ไม่มี scope (= ทั้งกลุ่ม) ไม่ได้ → 403 `FORBIDDEN`; **และ** `checkTargetWithinGrantor` — สมาชิกที่มีอยู่แล้วจะแก้หรือลบ (`DeleteUserPermissionShop`) ได้ก็ต่อเมื่อบทบาท+ขอบเขต **ปัจจุบัน** ของสมาชิกนั้นอยู่ในขอบเขตของผู้แก้ด้วย (ตรวจแค่ขอบเขตที่ขอไม่พอ); สมาชิกเป้าหมายหาจาก editusername → username → useruid (`findSaveTarget`) เพราะ repo upsert ตาม useruid แล้วตาม username — editusername ปลอมจึงข้ามการตรวจไม่ได้ และลบด้วย useruid ของสมาชิกที่ตรวจแล้ว; ข้อความผ่าน key `ss_err_access_scope_invalid` / `ss_err_access_scope_exceeds_grantor` ใน `languages.tsv` แปลตาม `lang`/`Accept-Language`
4. **แสดง**: `findMember` คืน OWNER/ADMIN ที่ไม่มี scope เป็น `[{"scopetype":"holding"}]` แทน snapshot รายบริษัท — ช่องติ๊กขึ้นถูก และบันทึกซ้ำยังเป็นทั้งกลุ่ม
5. **frontend**: BFF `normalizeScopeRule` + editor `normalizeHoldingScopeRule` (จอหลักและ `holding-scope-editor.tsx`) — เฉพาะ `scopetype: "holding"` ตรง ๆ เท่านั้นที่เป็น holding; กฎที่มีแค่ `companyuid` ยังเป็นกฎบริษัท; BFF **ปฏิเสธทั้งคำขอ** (400 `VALIDATION_FAILED` + key `ss_err_access_scope_invalid`, `hasInvalidAccessScopeRules`) เมื่อมีกฎใดไม่รู้จัก/ไม่มีบริษัท/กฎสาขาไม่มีสาขา — ไม่ทิ้งเงียบ ๆ เพราะ OWNER/ADMIN ที่เหลือ `[]` = ทั้งกลุ่ม; ติ๊ก "ใช้ได้ทั้งกลุ่มกิจการ" ส่ง `[{scopetype:"holding"}]` เพียงกฎเดียว และเอาติ๊กออกได้บริษัท/สาขาเดิมคืน (`toggleHoldingScopeRules`)
6. **backend รายละเอียดเสริม**: กฎ `company` ถูกล้าง `branchuid/branchcode` ตอน hydrate (ค่าที่ไม่ได้ตรวจจะเปลี่ยนความหมายกฎ); รายการสมาชิก (`FindByUserInShopPageWithProfileMatches`) กับรายละเอียดรายงาน OWNER/ADMIN ที่ไม่มี scope เป็น `[{"scopetype":"holding"}]` เหมือนกัน (`scanMember`)

## Alternatives

- **ขยายตอนบันทึก (แบบเดิม)** — ไม่เลือก: บริษัทใหม่ต้องให้ผู้ดูแลกลับไปแก้สมาชิกทุกคน และค่าที่เก็บไม่สะท้อนเจตนา "ทั้งกลุ่ม"
- **ให้ helper ใน `organization/access` (`AllowsCompany`/`AllowedCompanyUIDs`) รู้จัก holding เอง** — ไม่เลือก: helper ไม่มีฐานข้อมูล จึงไม่รู้ว่ามีบริษัทไหน active; ขยายที่ `FindActiveMembership` จุดเดียวแล้ว helper ทำงานกับกฎรายบริษัทเหมือนเดิม
- **ตัวกรองแบบ fail-open ใน frontend เดิม** — ไม่เลือก: กฎเสียกลายเป็นสิทธิ์กว้างสุด

## Consequences

- ✅ บริษัท/สาขาใหม่ใช้ได้ทันทีสำหรับสมาชิกทั้งกลุ่ม; ปิดบริษัท/สาขาแล้วหลุดทันทีเพราะผู้อ่านเปิดเฉพาะ active
- ✅ ADMIN ที่จำกัดบริษัทให้ขอบเขตเกินตัวเองไม่ได้ **และ** แก้/ลด/ถอดบทบาท/ลบสมาชิกที่ขอบเขตปัจจุบันเกินตัวเองไม่ได้ (เช่น ADMIN ทั้งกลุ่ม หรือพนักงานบริษัทอื่น) — แก้ได้เฉพาะสมาชิกที่อยู่ในขอบเขตตัวเองทั้งหมด; **API/MCP token ก็เช่นกัน**: ออก token ได้เฉพาะบริษัทที่ตัวเองเข้าได้ทั้งบริษัท (`create` ใน `backend/internal/mcptoken/http.go` → 400 key `mcp_err_company_outside_scope`; รายชื่อบริษัทในฟอร์มกรองตามขอบเขตเดียวกัน) และทุกครั้งที่ใช้ token ระบบตัดบริษัทที่ผู้ออก **ณ ตอนนั้น** เข้าไม่ได้ทิ้ง (`authenticateAudience` ใน `token.go`) + GL `resolveScope` ตรวจขอบเขตผู้ออกทุกคำขอ token เหมือน session — ลดขอบเขตผู้ออกแล้ว token เดิมแคบลงทันทีโดยไม่ต้องเพิกถอน; ขอบเขตไม่ให้สิทธิ์หน้าจอ/GL (`role_permissions` แยก)
- ⚠️ **ความปลอดภัย**: สิทธิ์ทั้งกลุ่มเปิดบริษัทที่สร้างในอนาคตให้อัตโนมัติ — ให้เฉพาะคนที่ไว้ใจทั้งกลุ่ม; มีได้เฉพาะ OWNER/ADMIN ทั้งกลุ่มเป็นผู้ให้
- ⚠️ ADMIN ที่จำกัดบริษัทจะแก้ข้อมูลใด ๆ ของผู้ใช้ที่มีขอบเขตกว้างกว่าหรือนอกขอบเขตตัวเองไม่ได้เลย (แม้แค่ตำแหน่ง/ชื่อ) — ต้องให้ OWNER หรือ ADMIN ทั้งกลุ่มแก้
- ⚠️ ผู้ใช้ที่ถูกปิด (`users.is_active = false`) ไม่ถูกอ่านเป็นสมาชิก จึงไม่ถูกตรวจเป็น "สมาชิกเดิม" — การบันทึกชื่อผู้ใช้นั้นใหม่จะ upsert ทับ membership เดิมโดยตรวจแค่ขอบเขตที่ขอ
- ⚠️ `AccessShop` ยังไม่ตรวจว่าสาขาที่ขอมีจริง/active (เหมือนเดิมกับกฎบริษัท allbranches) — live authorization ปฏิเสธทุกคำขอถัดไปอยู่แล้ว
- ข้อมูลเก่าที่ถูกแตกเป็นรายบริษัทไว้แล้วไม่ถูกแปลงกลับ (ช่วง dev เดินหน้าอย่างเดียว) — บันทึกผู้ใช้นั้นใหม่พร้อมติ๊ก "ใช้ได้ทั้งกลุ่มกิจการ"
