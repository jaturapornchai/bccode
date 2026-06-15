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
	YearType          string   `json:"yeartype" bson:"yeartype" validate:"max=21"`
	DedeZoom          DedeZoom `json:"dedezoom" bson:"dedezoom"`
}

type DedeZoom struct {
	Email       string `json:"email" bson:"email"`
	PhoneNumber string `json:"phonenumber" bson:"phonenumber"`
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
	CreatedAt    time.Time `json:"-" bson:"createdat,omitempty"`
	UpdatedAt    time.Time `json:"-" bson:"updatedat,omitempty"`
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
	HoldingCode  string `json:"holdingcode,omitempty"`
}

type UserProfile struct {
	UsernameCode `bson:"inline"`
	UserDetail   `bson:"inline"`
	CreatedAt    time.Time `json:"-" bson:"createdat,omitempty"`
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
	HoldingCode string `json:"holdingcode"`
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
	UserUID     string   `json:"useruid" bson:"useruid"`
	HoldingCode string   `json:"holdingcode" bson:"holdingcode"`
	Role        UserRole `json:"role" bson:"role"`
}

// DocumentApproval - ข้อมูลการอนุมัติแยกตามประเภทเอกสาร
type DocumentApproval struct {
	ApprovalRole      int     `json:"approvalrole" bson:"approvalrole"`           // 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	MaxApprovalAmount float64 `json:"maxapprovalamount" bson:"maxapprovalamount"` // วงเงินอนุมัติสูงสุด (บาท)
}

type AccessScope struct {
	ScopeType    string `json:"scopetype" bson:"scopetype"`                           // holding, company, branch
	BusinessCode string `json:"businesscode,omitempty" bson:"businesscode,omitempty"` // company code
	BranchCode   string `json:"branchcode,omitempty" bson:"branchcode,omitempty"`     // Thai tax branch code
	AllBranches  bool   `json:"allbranches,omitempty" bson:"allbranches,omitempty"`
}

type ShopUser struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopUserBase   `bson:"inline"`
	IsFavorite     bool      `json:"isfavorite" bson:"isfavorite"`
	LastAccessedAt time.Time `json:"lastaccessedat" bson:"lastaccessedat"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position" bson:"position"`     // ตำแหน่งงาน
	Department string `json:"department" bson:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"lineuserid" bson:"lineuserid"`           // LINE User ID
	LineDisplayName string `json:"linedisplayname" bson:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl" bson:"linepictureurl"`   // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"poapproval,omitempty" bson:"poapproval,omitempty"`               // อนุมัติใบสั่งซื้อ
	QuotationApproval *DocumentApproval `json:"quotationapproval,omitempty" bson:"quotationapproval,omitempty"` // อนุมัติใบเสนอราคา
	AccessScopes      []AccessScope     `json:"accessscopes,omitempty" bson:"accessscopes,omitempty"`
}

func (*ShopUser) CollectionName() string {
	return "shopusers"
}

type ShopUserInfo struct {
	HoldingCode    string         `json:"holdingcode" bson:"holdingcode"`
	Name           string         `json:"name" bson:"name"`
	Names          []models.NameX `json:"names" bson:"names"`
	BranchCode     string         `json:"branchcode" bson:"branchcode"`
	Role           UserRole       `json:"role" bson:"role"`
	IsFavorite     bool           `json:"isfavorite" bson:"isfavorite"`
	LastAccessedAt time.Time      `json:"lastaccessedat" bson:"lastaccessedat"`
	CreatedBy      string         `json:"createdby" bson:"createdby"`
}

func (*ShopUserInfo) CollectionName() string {
	return "shopusers"
}

type UserRoleRequest struct {
	HoldingCode  string   `json:"holdingcode" bson:"holdingcode"`
	EditUsername string   `json:"editusername" bson:"editusername"`
	Username     string   `json:"username" bson:"username"`
	UserUID      string   `json:"useruid,omitempty" bson:"useruid,omitempty"`
	Role         UserRole `json:"role" bson:"role"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position" bson:"position"`     // ตำแหน่งงาน
	Department string `json:"department" bson:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"lineuserid" bson:"lineuserid"`           // LINE User ID
	LineDisplayName string `json:"linedisplayname" bson:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl" bson:"linepictureurl"`   // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"poapproval,omitempty" bson:"poapproval,omitempty"`
	QuotationApproval *DocumentApproval `json:"quotationapproval,omitempty" bson:"quotationapproval,omitempty"`
	AccessScopes      []AccessScope     `json:"accessscopes,omitempty" bson:"accessscopes,omitempty"`
}

type ShopUserAccessLog struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode    string             `json:"holdingcode" bson:"holdingcode"`
	Username       string             `json:"username" bson:"username"`
	Ip             string             `json:"ip" bson:"ip"`
	LastAccessedAt time.Time          `json:"lastaccessedat" bson:"lastaccessedat"`
}

func (*ShopUserAccessLog) CollectionName() string {
	return "shopuseraccesslogs"
}

type ShopUserProfile struct {
	ShopUserBase    `bson:"inline"`
	UserProfileName string `json:"userprofilename" bson:"userprofilename"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position" bson:"position"`     // ตำแหน่งงาน
	Department string `json:"department" bson:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"lineuserid" bson:"lineuserid"`           // LINE User ID
	LineDisplayName string `json:"linedisplayname" bson:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl" bson:"linepictureurl"`   // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"poapproval,omitempty" bson:"poapproval,omitempty"`
	QuotationApproval *DocumentApproval `json:"quotationapproval,omitempty" bson:"quotationapproval,omitempty"`
	AccessScopes      []AccessScope     `json:"accessscopes,omitempty" bson:"accessscopes,omitempty"`
}

// func (u UserRole) EqualString(userRoleStr string)  bool {
// 	switch u {
// 		case
// 	}
// }
