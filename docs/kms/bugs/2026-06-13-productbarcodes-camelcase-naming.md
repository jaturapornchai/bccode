# Collection name camelCase ละเมิด naming rule

#bc-account #mongodb #naming

## Symptom
[product browser] mongodb/product-schema.html อ้าง collection ชื่อ `productBarcodes` (camelCase B ตัวใหญ่) ขณะที่ naming rule บังคับ identifier ทุกตัวเป็นตัวเล็กติดกัน และไฟล์อื่น (pgsql/productbarcode-schema.html) ใช้ `productbarcodes` lowercase ถูกต้อง → ขัดแย้งกันเอง

## Root cause
typo ตอนเขียน doc — พิมพ์ camelCase ตามนิสัย JS แทนที่จะตามกฎ db identifier ของโครงการ หลุด review เพราะปรากฏจุดเดียวใน 35 ไฟล์

## Detection
เจอด้วย multi-agent consistency audit (workflow `product-consistency-audit`): fan-out 10 มิติ + adversarial verify. มิติ naming-convention จับได้ confidence 0.95 — verifier re-read AGENTS.md:18 (กฎ) + product-schema.html:128 (violation) ยืนยันจริง

## Fix
[product browser] mongodb/product-schema.html:128 `productBarcodes` → `productbarcodes`

## Regression test
tests/product-browser.test.mjs — "collection names stay lowercase (camelCase regression guard)": assert ไม่มี `>productBarcodes<` ใน mongo/pgsql/clickhouse schema + ต้องมี `>productbarcodes<`

## หมายเหตุ audit
อีก 3 findings ถูก verifier ตัดทิ้งเป็น intentional design (ไม่ใช่ bug): package dimension เป็น logistics ไม่ใช่ accounting (float ได้), tax effective-dating อยู่คนละ layer (mongo source vs pgsql current-only), clickhouse productfact เป็น future-model ที่ doc ระบุว่ายังไม่ implement
