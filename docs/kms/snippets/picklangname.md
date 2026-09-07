# picklangname — function กลางเลือกชื่อ NameX ตามภาษา

#bc-account #postgres #clickhouse #schema

> ⚠️ **PostgreSQL: DEPRECATED 2026-06-11** — ลบ `picklangname` ออกจาก pgsql แล้ว เปลี่ยนเป็นตาราง `productlanguage` join (LEFT JOIN ×2 + COALESCE) ตามคำสั่งลุงจืด ห้ามสร้าง function เลือกภาษาใน pgsql อีก — ดู [[2026-06-11-productlanguage-join-table]]. ส่วน ClickHouse (lambda UDF) ด้านล่างยังใช้อยู่.

ใช้กับ field `[]NameX` (`[{"code":"th","name":"..."}]`) ทุกตัวใน read model — ห้าม copy subquery/expression inline เอง.
ลำดับ: langcode → fallbacklangcode → ชื่อแรกที่มีค่า → defaultname. ที่มา: [[2026-06-10-product-readmodel-parity-taxtype-rename]]

## PostgreSQL — ❌ เลิกใช้แล้ว (2026-06-11) เก็บไว้อ้างอิงเท่านั้น
```sql
CREATE OR REPLACE FUNCTION picklangname(
    names jsonb, langcode text, fallbacklangcode text, defaultname text
) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE AS $$
    SELECT COALESCE(NULLIF((
        SELECT value->>'name'
        FROM jsonb_array_elements(names) WITH ORDINALITY AS t(value, ord)
        WHERE COALESCE(value->>'name', '') <> ''
        ORDER BY CASE WHEN value->>'code' = langcode THEN 0
                      WHEN value->>'code' = fallbacklangcode THEN 1
                      ELSE 2 END, ord
        LIMIT 1
    ), ''), defaultname)
$$;
-- ใช้: picklangname(p.names, i.langcode, i.fallbacklangcode, p.code) AS name
```

## ClickHouse (lambda UDF ระดับ server — สร้างตอน deploy/migration)
```sql
CREATE FUNCTION IF NOT EXISTS picklangname AS (names, langcode, fallbacklangcode, defaultname) ->
    coalesce(
        nullIf(JSONExtractString(arrayFirst(x -> JSONExtractString(x, 'code') = langcode, JSONExtractArrayRaw(names)), 'name'), ''),
        nullIf(JSONExtractString(arrayFirst(x -> JSONExtractString(x, 'code') = fallbacklangcode, JSONExtractArrayRaw(names)), 'name'), ''),
        defaultname
    );
```

## ข้อระวัง
- pgsql: ห้ามเปลี่ยนเป็น plpgsql (เสีย inline, ช้าลง per-row)
- clickhouse UDF เป็น object ระดับ server ไม่ใช่ per-database
- function ไม่ทำให้เร็วขึ้นกว่า inline — ประโยชน์คือ consistency/ลด bug
- clickhouse lambda ไม่มี ordinality fallback "ชื่อแรกที่มีค่า" เหมือน pgsql (ได้แค่ lang → fallback → default)
