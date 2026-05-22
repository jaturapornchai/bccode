package models

import (
	"time"

	"smlcloudplatform/internal/models"
	timezone "smlcloudplatform/internal/models/timezone"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const userCollectionName = "users"

type UserDetail struct {
	UID               string `json:"uid" bson:"uid"`
	Name              string `json:"name,omitempty"`
	Avatar            string `json:"avatar"`
	timezone.Timezone `bson:"inline"`
	YearType          string   `json:"year_type" bson:"year_type" validate:"max=21"`
	DedeZoom          DedeZoom `json:"dede_zoom" bson:"dede_zoom"`
	RegisterType      string   `json:"register_type" bson:"register_type"`
}

type DedeZoom struct {
	Email       string `json:"email" bson:"email"`
	PhoneNumber string `json:"phone_number" bson:"phone_number"`
	Address     string `json:"address" bson:"address"`
}

type UsernameField struct {
	Username string `json:"username,omitempty" bson:"username" validate:"required,gte=3,max=233"` // validate:"required,alphanum,gte=3,max=233"
}

type PhoneNumberField struct {
	CountryCode string `json:"country_code" bson:"country_code" validate:"required,max=20"`
	PhoneNumber string `json:"phone_number" bson:"phone_number" validate:"required,max=100"`
}

type EmailField struct {
	Email string `json:"email,omitempty" bson:"email" validate:"required,email,max=233"`
}

type UserPassword struct {
	Password string `json:"password,omitempty" bson:"password" validate:"required,gte=5,max=233"`
}

type UserDoc struct {
	ID               primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UsernameField    `bson:"inline"`
	EmailField       `bson:"inline"`
	PhoneNumberField `bson:"inline"`
	UserPassword     `bson:"inline"`
	UserDetail       `bson:"inline"`

	// === ข้อมูล LINE (ระดับ user — ใช้ร่วมทุก shop) ===
	LineUserID      string `json:"line_user_id" bson:"line_user_id"`           // LINE User ID
	LineDisplayName string `json:"line_display_name" bson:"line_display_name"` // LINE Display Name
	LinePictureURL  string `json:"line_picture_url" bson:"line_picture_url"`   // LINE Profile Picture URL

	CreatedAt  time.Time `json:"-" bson:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"-" bson:"updated_at,omitempty"`
	DisabledAt time.Time `json:"disabled_at,omitempty" bson:"disabled_at,omitempty"`
}

func (*UserDoc) CollectionName() string {
	return userCollectionName
}

type RegisterEmailRequest struct {
	EmailField
	UserPassword
	UserDetail
}

type RegisterPhoneNumberRequest struct {
	PhoneNumberField
	UsernameField
	UserPassword
	UserDetail
	OTPVerifyRequest
}

type ForgotPasswordPhoneNumberRequest struct {
	PhoneNumberField
	UserPassword
	OTPVerifyRequest
}

func (RegisterEmailRequest) CollectionName() string {
	return userCollectionName
}

func (RegisterPhoneNumberRequest) CollectionName() string {
	return userCollectionName
}

// RegisterUsernameRequest — สำหรับพนักงานสมัครด้วยรหัสพนักงาน + รหัสผ่าน (ไม่ต้องมี email)
type RegisterUsernameRequest struct {
	UsernameField `bson:"inline"`
	UserPassword  `bson:"inline"`
	UserDetail    `bson:"inline"`
}

func (RegisterUsernameRequest) CollectionName() string {
	return userCollectionName
}

type UserRequest struct {
	UsernameField `bson:"inline"`
	UserPassword  `bson:"inline"`
	UserDetail    `bson:"inline"`
}

func (*UserRequest) CollectionName() string {
	return userCollectionName
}

type UserLoginRequest struct {
	UsernameField `bson:"inline"`
	UserPassword  `bson:"inline"`
	ShopID        string `json:"shopid,omitempty"`
}

type PosLoginRequest struct {
	UsernameField `bson:"inline"`
	ShopID        string `json:"shopid,omitempty"`
}

type UserLoginPhoneNumberRequest struct {
	PhoneNumberField `bson:"inline"`
	UserPassword     `bson:"inline"`
	ShopID           string `json:"shopid,omitempty"`
}

type UserProfile struct {
	UsernameField     `bson:"inline"`
	Email             string `json:"email,omitempty" bson:"email,omitempty"`
	UserDetail        `bson:"inline"`
	UserPassword      `bson:"inline"`
	IsDefaultPassword bool `json:"isdefaultpassword" bson:"-"`

	// === ข้อมูล LINE (ระดับ user) ===
	LineUserID      string `json:"line_user_id" bson:"line_user_id"`
	LineDisplayName string `json:"line_display_name" bson:"line_display_name"`
	LinePictureURL  string `json:"line_picture_url" bson:"line_picture_url"`

	CreatedAt time.Time `json:"-" bson:"created_at,omitempty"`
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
	ShopID string `json:"shopid" validate:"required"`
}

type UserRole = uint8

const (
	ROLE_USER  UserRole = iota // "USER"
	ROLE_ADMIN                 // "ADMIN"
	ROLE_OWNER                 // "OWNER"

	ROLE_SYSTEM = 255 // APP MANAGER
)

const DefaultUserPassword = "12345"

type ShopUserBase struct {
	Username string   `json:"username" bson:"username"`
	ShopID   string   `json:"shopid" bson:"shopid"`
	Role     UserRole `json:"role" bson:"role"`
}

// DocumentApproval - ข้อมูลการอนุมัติแยกตามประเภทเอกสาร
type DocumentApproval struct {
	ApprovalRole      int     `json:"approval_role" bson:"approval_role"`             // 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	MaxApprovalAmount float64 `json:"max_approval_amount" bson:"max_approval_amount"` // วงเงินอนุมัติสูงสุด (บาท)
}

type ShopUser struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopUserBase     `bson:"inline"`
	IsFavorite       bool      `json:"is_favorite" bson:"isfavorite"`
	LastAccessedAt   time.Time `json:"last_accessed_at" bson:"lastaccessedat"`
	IsCreator        bool      `json:"is_creator,omitempty" bson:"-"`
	IsAccessDisabled bool      `json:"is_access_disabled" bson:"isaccessdisabled"`
	AccessDisabledAt time.Time `json:"access_disabled_at,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy string    `json:"access_disabled_by,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time `json:"access_enabled_at,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy  string    `json:"access_enabled_by,omitempty" bson:"accessenabledby,omitempty"`

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
}

func (*ShopUser) CollectionName() string {
	return "shopUsers"
}

type ShopUserInfo struct {
	ShopID string `json:"shopid" bson:"shopid"`
	Name   string `json:"name" bson:"name1"`
	// Name1          string         `json:"name1" bson:"name1"`
	MainShopId       string         `json:"main_shop_id" bson:"main_shop_id"`
	Names            []models.NameX `json:"names" bson:"names"`
	BranchCode       string         `json:"branchcode" bson:"branchcode"`
	Role             UserRole       `json:"role" bson:"role"`
	IsFavorite       bool           `json:"is_favorite" bson:"isfavorite"`
	LastAccessedAt   time.Time      `json:"last_accessed_at" bson:"lastaccessedat"`
	CreatedBy        string         `json:"createdby" bson:"createdby"`
	IsCreator        bool           `json:"is_creator,omitempty" bson:"-"`
	IsAccessDisabled bool           `json:"is_access_disabled" bson:"isaccessdisabled"`
	AccessDisabledAt time.Time      `json:"access_disabled_at,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy string         `json:"access_disabled_by,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time      `json:"access_enabled_at,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy  string         `json:"access_enabled_by,omitempty" bson:"accessenabledby,omitempty"`
}

func (*ShopUserInfo) CollectionName() string {
	return "shopUsers"
}

type UserRoleRequest struct {
	ShopID           string    `json:"shopid" bson:"shopid"`
	EditUsername     string    `json:"editusername" bson:"editusername"`
	Username         string    `json:"username" bson:"username"`
	UserProfileName  string    `json:"user_profile_name" bson:"user_profile_name"`
	Email            string    `json:"email,omitempty" bson:"email,omitempty"`
	Role             UserRole  `json:"role" bson:"role"`
	IsAccessDisabled bool      `json:"is_access_disabled" bson:"is_access_disabled"`
	AccessDisabledAt time.Time `json:"access_disabled_at,omitempty" bson:"access_disabled_at,omitempty"`
	AccessDisabledBy string    `json:"access_disabled_by,omitempty" bson:"access_disabled_by,omitempty"`
	AccessEnabledAt  time.Time `json:"access_enabled_at,omitempty" bson:"access_enabled_at,omitempty"`
	AccessEnabledBy  string    `json:"access_enabled_by,omitempty" bson:"access_enabled_by,omitempty"`

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
}

type ShopUserAccessLog struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopID         string             `json:"shopid" bson:"shopid"`
	Username       string             `json:"username" bson:"username"`
	Ip             string             `json:"ip" bson:"ip"`
	LastAccessedAt time.Time          `json:"last_accessed_at" bson:"last_accessed_at"`
}

func (*ShopUserAccessLog) CollectionName() string {
	return "shop_user_access_logs"
}

type ShopUserProfile struct {
	ShopUserBase     `bson:"inline"`
	UID              string    `json:"uid,omitempty" bson:"uid,omitempty"`
	Email            string    `json:"email,omitempty" bson:"email,omitempty"`
	UserProfileName  string    `json:"user_profile_name" bson:"user_profile_name"`
	IsCreator        bool      `json:"is_creator,omitempty" bson:"-"`
	IsAccessDisabled bool      `json:"is_access_disabled" bson:"is_access_disabled"`
	AccessDisabledAt time.Time `json:"access_disabled_at,omitempty" bson:"access_disabled_at,omitempty"`
	AccessDisabledBy string    `json:"access_disabled_by,omitempty" bson:"access_disabled_by,omitempty"`
	AccessEnabledAt  time.Time `json:"access_enabled_at,omitempty" bson:"access_enabled_at,omitempty"`
	AccessEnabledBy  string    `json:"access_enabled_by,omitempty" bson:"access_enabled_by,omitempty"`

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
}
