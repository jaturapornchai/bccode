package models

type ProcessMongoEmployeeModel struct {
	Code        string `json:"code" bson:"code"`
	Name        string `json:"name" bson:"name"`
	Password    string `json:"password" bson:"password"`
	HoldingCode string `json:"holdingcode" bson:"holdingcode"`

	// === ฟิลด์สำหรับระบบอนุมัติ ===
	// ตำแหน่งงาน เช่น "หัวหน้าแผนก", "ผู้จัดการฝ่าย", "ผู้อำนวยการ"
	Position string `json:"position" bson:"position"`
	// แผนก เช่น "IT", "บัญชี", "จัดซื้อ"
	Department string `json:"department" bson:"department"`
	// บทบาทการอนุมัติ: 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	ApprovalRole int `json:"approvalrole" bson:"approvalrole"`
	// วงเงินอนุมัติสูงสุด (บาท) - 0 = ไม่จำกัด
	MaxApprovalAmount float64 `json:"maxapprovalamount" bson:"maxapprovalamount"`
}
