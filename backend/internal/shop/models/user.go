package models

import (
	"time"

	"smlcloudplatform/internal/models"
	timezone "smlcloudplatform/internal/models/timezone"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const userCollectionName = "users"

type UserDetail struct {
	Name              string `json:"name,omitempty"  validate:"required"`
	Avatar            string `json:"avatar"`
	timezone.Timezone `bson:"inline"`
	YearType          string   `json:"year_type" bson:"year_type" validate:"max=21"`
	DedeZoom          DedeZoom `json:"dede_zoom" bson:"dede_zoom"`
}

type DedeZoom struct {
	Email       string `json:"email" bson:"email"`
	PhoneNumber string `json:"phone_number" bson:"phone_number"`
	Address     string `json:"address" bson:"address"`
}

type UsernameCode struct {
	Username string `json:"username,omitempty" bson:"username" validate:"required,alphanum,gte=5,max=233"`
}

type UserPassword struct {
	Password string `json:"password,omitempty" bson:"password" validate:"required,gte=5,max=233"`
}

type UserDoc struct {
	ID           primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UsernameCode `bson:"inline"`
	UserPassword `bson:"inline"`
	UserDetail   `bson:"inline"`
	CreatedAt    time.Time `json:"-" bson:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"-" bson:"updated_at,omitempty"`
}

func (*UserDoc) CollectionName() string {
	return userCollectionName
}

type UserRequest struct {
	UsernameCode `bson:"inline"`
	UserPassword `bson:"inline"`
	UserDetail   `bson:"inline"`
}

func (*UserRequest) CollectionName() string {
	return userCollectionName
}

type UserLoginRequest struct {
	UsernameCode `bson:"inline"`
	UserPassword `bson:"inline"`
	HoldingCode  string `json:"holding_code,omitempty"`
}

type UserProfile struct {
	UsernameCode `bson:"inline"`
	UserDetail   `bson:"inline"`
	CreatedAt    time.Time `json:"-" bson:"created_at,omitempty"`
}

func (UserProfile) CollectionName() string {
	return userCollectionName
}

type UserProfileRequest struct {
	UserDetail `bson:"inline"`
}

type UserPasswordRequest struct {
	CurrentPassword string `json:"currentpassword" bson:"currentpassword" validate:"required,gte=5"`
	NewPassword     string `json:"newpassword" bson:"newpassword" validate:"required,gte=5"`
}

type UserProfileReponse struct {
	Success bool        `json:"success"`
	Data    UserProfile `json:"data"`
}

type ShopSelectRequest struct {
	HoldingCode string `json:"holding_code"`
}

type UserRole = uint8

const (
	ROLE_USER  UserRole = iota // "USER"
	ROLE_ADMIN                 // "ADMIN"
	ROLE_OWNER                 // "OWNER"

	ROLE_SYSTEM = 255 // APP MANAGER
)

type ShopUserBase struct {
	Username    string   `json:"username" bson:"username"`
	UserUID     string   `json:"user_uid" bson:"user_uid"`
	HoldingCode string   `json:"holding_code" bson:"holding_code"`
	Role        UserRole `json:"role" bson:"role"`
}

// DocumentApproval - ข้อมูลการอนุมัติแยกตามประเภทเอกสาร
type DocumentApproval struct {
	ApprovalRole      int     `json:"approval_role" bson:"approval_role"`             // 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	MaxApprovalAmount float64 `json:"max_approval_amount" bson:"max_approval_amount"` // วงเงินอนุมัติสูงสุด (บาท)
}

type AccessScope struct {
	ScopeType    string `json:"scope_type" bson:"scope_type"`                           // holding, company, branch
	BusinessCode string `json:"business_code,omitempty" bson:"business_code,omitempty"` // company code
	BranchCode   string `json:"branch_code,omitempty" bson:"branch_code,omitempty"`     // Thai tax branch code
	AllBranches  bool   `json:"all_branches,omitempty" bson:"all_branches,omitempty"`
}

type ShopUser struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopUserBase   `bson:"inline"`
	IsFavorite     bool      `json:"is_favorite" bson:"is_favorite"`
	LastAccessedAt time.Time `json:"last_accessed_at" bson:"last_accessed_at"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position" bson:"position"`     // ตำแหน่งงาน
	Department string `json:"department" bson:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"line_user_id" bson:"line_user_id"`           // LINE User ID
	LineDisplayName string `json:"line_display_name" bson:"line_display_name"` // LINE Display Name
	LinePictureURL  string `json:"line_picture_url" bson:"line_picture_url"`   // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"po_approval,omitempty" bson:"po_approval,omitempty"`               // อนุมัติใบสั่งซื้อ
	QuotationApproval *DocumentApproval `json:"quotation_approval,omitempty" bson:"quotation_approval,omitempty"` // อนุมัติใบเสนอราคา
	AccessScopes      []AccessScope     `json:"access_scopes,omitempty" bson:"access_scopes,omitempty"`
}

func (*ShopUser) CollectionName() string {
	return "shop_users"
}

type ShopUserInfo struct {
	HoldingCode    string         `json:"holding_code" bson:"holding_code"`
	Name           string         `json:"name" bson:"name"`
	Names          []models.NameX `json:"names" bson:"names"`
	BranchCode     string         `json:"branchcode" bson:"branchcode"`
	Role           UserRole       `json:"role" bson:"role"`
	IsFavorite     bool           `json:"is_favorite" bson:"is_favorite"`
	LastAccessedAt time.Time      `json:"last_accessed_at" bson:"last_accessed_at"`
	CreatedBy      string         `json:"createdby" bson:"createdby"`
}

func (*ShopUserInfo) CollectionName() string {
	return "shop_users"
}

type UserRoleRequest struct {
	HoldingCode  string   `json:"holding_code" bson:"holding_code"`
	EditUsername string   `json:"editusername" bson:"editusername"`
	Username     string   `json:"username" bson:"username"`
	UserUID      string   `json:"user_uid,omitempty" bson:"user_uid,omitempty"`
	Role         UserRole `json:"role" bson:"role"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position" bson:"position"`     // ตำแหน่งงาน
	Department string `json:"department" bson:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"line_user_id" bson:"line_user_id"`           // LINE User ID
	LineDisplayName string `json:"line_display_name" bson:"line_display_name"` // LINE Display Name
	LinePictureURL  string `json:"line_picture_url" bson:"line_picture_url"`   // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"po_approval,omitempty" bson:"po_approval,omitempty"`
	QuotationApproval *DocumentApproval `json:"quotation_approval,omitempty" bson:"quotation_approval,omitempty"`
	AccessScopes      []AccessScope     `json:"access_scopes,omitempty" bson:"access_scopes,omitempty"`
}

type ShopUserAccessLog struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode    string             `json:"holding_code" bson:"holding_code"`
	Username       string             `json:"username" bson:"username"`
	Ip             string             `json:"ip" bson:"ip"`
	LastAccessedAt time.Time          `json:"last_accessed_at" bson:"last_accessed_at"`
}

func (*ShopUserAccessLog) CollectionName() string {
	return "shop_user_access_logs"
}

type ShopUserProfile struct {
	ShopUserBase    `bson:"inline"`
	UserProfileName string `json:"user_profile_name" bson:"user_profile_name"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position" bson:"position"`     // ตำแหน่งงาน
	Department string `json:"department" bson:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"line_user_id" bson:"line_user_id"`           // LINE User ID
	LineDisplayName string `json:"line_display_name" bson:"line_display_name"` // LINE Display Name
	LinePictureURL  string `json:"line_picture_url" bson:"line_picture_url"`   // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"po_approval,omitempty" bson:"po_approval,omitempty"`
	QuotationApproval *DocumentApproval `json:"quotation_approval,omitempty" bson:"quotation_approval,omitempty"`
	AccessScopes      []AccessScope     `json:"access_scopes,omitempty" bson:"access_scopes,omitempty"`
}

// func (u UserRole) EqualString(userRoleStr string)  bool {
// 	switch u {
// 		case
// 	}
// }
