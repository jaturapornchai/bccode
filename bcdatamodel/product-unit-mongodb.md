# MongoDB JSON Structure: Product Unit / หน่วยนับสินค้า

เอกสารนี้อธิบายโครงสร้างข้อมูล MongoDB สำหรับหน้าจอ `หน่วยนับสินค้า` (`/productunit`) โดยอ้างอิงจาก source ปัจจุบันของระบบ

## คุณสมบัติ

- `units` รองรับหลายบริษัท
- สามารถกำหนดได้ว่า หน่วยอะไร ใช้กับบริษัทอะไรได้บ้าง ได้หลายบริษัท
- เมื่อทำรายการอื่นที่เกี่ยวข้อง เช่น หน้าจอซื้อ หน้าจอขาย จะเห็นเฉพาะหน่วยของบริษัทนั้น

## สรุปกติกาสำคัญ

| เรื่อง | กติกา |
|---|---|
| แหล่งข้อมูลจริง | CRUD หน่วยนับสินค้าใช้ MongoDB collection `units` |
| ตัวตนถาวรของ record | ใช้ `guid_fixed` |
| รหัสที่ผู้ใช้เห็น | ใช้ `unitcode` |
| การอ้างอิงจากสินค้า/บาร์โค้ด/เอกสาร | อ้างด้วย `unitcode` ไม่ใช่ `guid_fixed` |
| ชื่อหลายภาษา | เก็บใน `names[]` |
| สิทธิ์บริษัท | เก็บใน `company_guids[]`; ว่าง = ใช้ได้ทุกบริษัท |
| การลบ | ใช้ soft delete ด้วย `deleted_at` / `deleted_by` |
| Sync downstream | MongoDB -> Kafka -> PostgreSQL -> ClickHouse |

## JSON พร้อมคำอธิบายในตัว

```json
{
  "collection": "units",
  "collection_description_th": "Collection หลักสำหรับเก็บหน่วยนับสินค้าใน MongoDB",
  "source_of_truth": {
    "store": "MongoDB",
    "description_th": "MongoDB เป็นแหล่งข้อมูลจริงของ CRUD หน่วยนับสินค้า ส่วน Kafka, PostgreSQL และ ClickHouse เป็นระบบ sync/projection/reporting ที่สร้างตามหลัง"
  },
  "crud_api": {
    "list": "GET /unit/list",
    "detail": "GET /unit/{guid_fixed}",
    "create": "POST /unit",
    "update": "PUT /unit/{guid_fixed}",
    "delete": "DELETE /unit/{guid_fixed}",
    "bulk_upsert": "POST /unit/bulk",
    "description_th": "Frontend เรียกผ่าน Next.js proxy /api/system-settings/productunit แล้วส่งต่อไป MongoDB-backed backend API"
  },
  "identity_rule": {
    "record_identity": "guid_fixed",
    "business_reference": "unitcode",
    "description_th": "ตัว record ใช้ guid_fixed เป็นรหัสถาวรสำหรับแก้ไข/ลบ แต่ข้อมูลธุรกิจอื่น เช่น สินค้า บาร์โค้ด และเอกสาร ต้องอ้างหน่วยนับด้วย unitcode ไม่ใช่ guid_fixed"
  },
  "active_document_example": {
    "_id": {
      "$oid": "665c00000000000000000001",
      "description_th": "MongoDB ObjectId ใช้ภายใน MongoDB เท่านั้น ไม่ใช้เป็นรหัสธุรกิจ"
    },
    "shopid": {
      "value": "3EL6B3jlbAcZTxiMLkGMwGNzzUo",
      "description_th": "รหัสกิจการ / tenant / workspace ใช้แบ่งข้อมูลแต่ละกิจการ ทุก query ต้องอยู่ภายใต้ shopid ที่ผู้ใช้มีสิทธิ์"
    },
    "guid_fixed": {
      "value": "2uXfJ5l2Wf0pQ8nR4YcLk9mA1Bb",
      "description_th": "GUID ถาวรของหน่วยนับ สร้างครั้งแรกแล้วห้ามเปลี่ยน ใช้เป็น record identity สำหรับ detail/update/delete/sync"
    },
    "parid": {
      "value": "",
      "description_th": "Partition/legacy field จาก base model ถ้าไม่ใช้ปล่อยว่างได้"
    },
    "unitcode": {
      "value": "PCS",
      "description_th": "รหัสหน่วยนับที่ผู้ใช้เห็นและใช้เป็นรหัสอ้างอิงทางธุรกิจ เช่น PCS, BOX, KG, JOB"
    },
    "names": {
      "value": [
        {
          "code": "th",
          "name": "ชิ้น",
          "isauto": false,
          "isdelete": false
        },
        {
          "code": "en",
          "name": "Piece",
          "isauto": false,
          "isdelete": false
        },
        {
          "code": "lo",
          "name": "",
          "isauto": false,
          "isdelete": false
        }
      ],
      "description_th": "ชื่อหน่วยนับหลายภาษา ใช้โครงสร้าง NameX โดย code คือรหัสภาษา name คือชื่อที่แสดงใน UI"
    },
    "company_guids": {
      "value": [
        "COMPANY_GUID_1",
        "COMPANY_GUID_2"
      ],
      "description_th": "รายชื่อบริษัทที่มีสิทธิ์ใช้หน่วยนับนี้ในกิจการเดียวกัน ถ้าไม่มี field นี้หรือเป็น array ว่าง แปลว่าใช้ได้ทุกบริษัท"
    },
    "createdby": {
      "value": "user@example.com",
      "description_th": "ผู้สร้างข้อมูล"
    },
    "created_at": {
      "value": "2026-06-01T00:00:00Z",
      "description_th": "วันที่และเวลาที่สร้างข้อมูล"
    },
    "updatedby": {
      "value": "user@example.com",
      "description_th": "ผู้แก้ไขล่าสุด ถ้ายังไม่เคยแก้อาจไม่มี field นี้"
    },
    "updated_at": {
      "value": "2026-06-01T00:00:00Z",
      "description_th": "วันที่และเวลาที่แก้ไขล่าสุด ถ้ายังไม่เคยแก้อาจไม่มี field นี้"
    }
  },
  "soft_delete_fields": {
    "deleted_by": {
      "type": "string",
      "description_th": "ผู้ลบข้อมูล จะถูกใส่เมื่อ soft delete"
    },
    "deleted_at": {
      "type": "datetime",
      "description_th": "วันที่และเวลาที่ลบข้อมูล ถ้าเอกสารยังใช้งานอยู่ต้องไม่มี field deleted_at"
    }
  },
  "validation_rules": {
    "shopid": {
      "required": true,
      "description_th": "ต้องมีเสมอและต้องตรงกับกิจการที่ผู้ใช้เลือก"
    },
    "guid_fixed": {
      "required": true,
      "immutable": true,
      "description_th": "ต้องสร้างตอน create และห้ามเปลี่ยนตลอดชีวิต record"
    },
    "unitcode": {
      "required": true,
      "max_length": 100,
      "unique_scope": "shopid + unitcode + active document",
      "description_th": "ต้องไม่ซ้ำในกิจการเดียวกันสำหรับเอกสารที่ยังไม่ถูกลบ"
    },
    "names": {
      "required": true,
      "min_items": 1,
      "unique_by": "code",
      "description_th": "ต้องมีอย่างน้อย 1 ภาษา และ code ภาษาไม่ควรซ้ำกันใน array เดียวกัน"
    },
    "company_guids": {
      "required": false,
      "description_th": "ถ้าว่างคือทุกบริษัทใช้ได้ ถ้ามีค่าให้กรองสิทธิ์ตามบริษัทที่เลือก"
    }
  },
  "recommended_mongodb_indexes": [
    {
      "name": "ux_units_shopid_unitcode_active",
      "keys": {
        "shopid": 1,
        "unitcode": 1
      },
      "unique": true,
      "partialFilterExpression": {
        "deleted_at": {
          "$exists": false
        }
      },
      "description_th": "กันรหัสหน่วยนับซ้ำในกิจการเดียวกัน เฉพาะ record ที่ยังไม่ถูกลบ"
    },
    {
      "name": "ix_units_shopid_guid_fixed_active",
      "keys": {
        "shopid": 1,
        "guid_fixed": 1
      },
      "unique": true,
      "partialFilterExpression": {
        "deleted_at": {
          "$exists": false
        }
      },
      "description_th": "ใช้ค้น detail/update/delete ด้วย record identity อย่างรวดเร็ว"
    },
    {
      "name": "ix_units_shopid_names_search",
      "keys": {
        "shopid": 1,
        "names.name": 1
      },
      "description_th": "ช่วยค้นหาจากชื่อหน่วยนับหลายภาษา"
    },
    {
      "name": "ix_units_shopid_company_guids",
      "keys": {
        "shopid": 1,
        "company_guids": 1
      },
      "description_th": "ช่วยกรองรายการหน่วยนับตามสิทธิ์บริษัท"
    }
  ]
}
```

## MongoDB Index Script ที่แนะนำ

```js
db.units.createIndex(
  { shopid: 1, unitcode: 1 },
  {
    name: "ux_units_shopid_unitcode_active",
    unique: true,
    partialFilterExpression: {
      deleted_at: { $exists: false }
    }
  }
);

db.units.createIndex(
  { shopid: 1, guid_fixed: 1 },
  {
    name: "ix_units_shopid_guid_fixed_active",
    unique: true,
    partialFilterExpression: {
      deleted_at: { $exists: false }
    }
  }
);

db.units.createIndex(
  { shopid: 1, "names.name": 1 },
  { name: "ix_units_shopid_names_search" }
);

db.units.createIndex(
  { shopid: 1, company_guids: 1 },
  { name: "ix_units_shopid_company_guids" }
);
```
