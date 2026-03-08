package models

type ProcessMongoEmployeeModel struct {
	Code     string `json:"code" bson:"code"`
	Name     string `json:"name" bson:"name"`
	Password string `json:"password" bson:"password"`
	ShopId   string `json:"shopid" bson:"shopid"`

	// === ฟิลด์สำหรับระบบอนุมัติ ===
	// ตำแหน่งงาน เช่น "หัวหน้าแผนก", "ผู้จัดการฝ่าย", "ผู้อำนวยการ"
	Position string `json:"position" bson:"position"`
	// แผนก เช่น "IT", "บัญชี", "จัดซื้อ"
	Department string `json:"department" bson:"department"`
	// บทบาทการอนุมัติ: 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	ApprovalRole int `json:"approval_role" bson:"approval_role"`
	// วงเงินอนุมัติสูงสุด (บาท) - 0 = ไม่จำกัด
	MaxApprovalAmount float64 `json:"max_approval_amount" bson:"max_approval_amount"`
}
