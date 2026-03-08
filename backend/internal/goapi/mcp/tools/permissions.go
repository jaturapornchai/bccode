package tools

// WriteTools คือรายการ tools ที่แก้ไขข้อมูล (create/update/delete)
// tool ที่ไม่อยู่ใน list นี้ = readonly โดยอัตโนมัติ
var WriteTools = map[string]bool{
	"create_unit":  true,
	"create_units": true,
	"update_unit":  true,
	"delete_unit":  true,
	"delete_units": true,
}

// Permission keywords สำหรับใช้ใน AllowedTools
const (
	PermAll       = "*"        // ทุก tool (readonly + write)
	PermReadonly  = "readonly" // เฉพาะ readonly tools
	PermReadWrite = "readwrite"
)

// IsWriteTool ตรวจว่า tool นี้เป็น write tool หรือไม่
func IsWriteTool(toolName string) bool {
	return WriteTools[toolName]
}

// IsToolAllowedByList ตรวจว่า tool ได้รับอนุญาตจาก AllowedTools หรือไม่
//
// Rules:
//   - nil หรือ [] → readonly ทั้งหมด
//   - มี "*" หรือ "readwrite" → ทุก tool
//   - มี "readonly" → readonly ทั้งหมด (ไม่รวม write)
//   - อื่นๆ → exact match ชื่อ tool
//
// สามารถผสมได้ เช่น ["readonly", "create_unit"] = readonly ทั้งหมด + create_unit
func IsToolAllowedByList(toolName string, allowedTools []string) bool {
	if len(allowedTools) == 0 {
		return !IsWriteTool(toolName)
	}

	for _, entry := range allowedTools {
		switch entry {
		case PermAll, PermReadWrite:
			return true
		case PermReadonly:
			if !IsWriteTool(toolName) {
				return true
			}
		default:
			if entry == toolName {
				return true
			}
		}
	}
	return false
}
