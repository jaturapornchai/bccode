# API Request: {ชื่อ feature}

## สิ่งที่ต้องการ
{อธิบายสั้นๆ ว่าต้องการ API อะไร ทำไม}

## Endpoint ที่ต้องการ
- **Method:** GET/POST/PUT/DELETE
- **Path:** /api/{resource}
- **Auth:** Bearer token (JWT)

## Request Parameters
```json
{
  "field1": "string — คำอธิบาย",
  "field2": "number — คำอธิบาย"
}
```

## Expected Response
```json
{
  "success": true,
  "data": {
    "field1": "string — คำอธิบาย",
    "field2": "number — คำอธิบาย"
  }
}
```

## Use Case (Frontend)
- หน้าจอ: {ชื่อหน้าจอ}
- การใช้งาน: {อธิบายว่า frontend จะใช้ API นี้ทำอะไร}

## Database Tables ที่เกี่ยวข้อง
{ผลจาก MCP get_database_schema — ถ้ามี}

## MCP Tool ที่อยากได้ (optional)
- **Tool name:** {tool_name}
- **Description:** {คำอธิบาย}
- **Parameters:** {param1 (type, required/optional), param2 (type)}
