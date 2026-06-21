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
	AvatarThumb       string `json:"avatarthumb"`
	timezone.Timezone `bson:"inline"`
	YearType          string   `json:"yeartype" bson:"yeartype" validate:"max=21"`
	DedeZoom          DedeZoom `json:"dedezoom" bson:"dedezoom"`
	RegisterType      string   `json:"registertype" bson:"registertype"`
}

type DedeZoom struct {
	Email       string `json:"email" bson:"email"`
	PhoneNumber string `json:"phonenumber" bson:"phonenumber"`
	Address     string `json:"address" bson:"address"`
}

type UsernameField struct {
	Username string `json:"username,omitempty" bson:"username" validate:"required,gte=3,max=233"` // validate:"required,alphanum,gte=3,max=233"
}

type PhoneNumberField struct {
	CountryCode string `json:"countrycode" bson:"countrycode" validate:"required,max=20"`
	PhoneNumber string `json:"phonenumber" bson:"phonenumber" validate:"required,max=100"`
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
	LineUserID      string `json:"lineuserid" bson:"lineuserid"`           // LINE User ID
	LineDisplayName string `json:"linedisplayname" bson:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl" bson:"linepictureurl"`   // LINE Profile Picture URL

	CreatedAt  time.Time `json:"-" bson:"createdat,omitempty"`
	UpdatedAt  time.Time `json:"-" bson:"updatedat,omitempty"`
	DisabledAt time.Time `json:"disabledat,omitempty" bson:"disabledat,omitempty"`
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
	HoldingCode   string `json:"holdingcode,omitempty"`
}

type PosLoginRequest struct {
	UsernameField `bson:"inline"`
	HoldingCode   string `json:"holdingcode,omitempty"`
}

type UserLoginPhoneNumberRequest struct {
	PhoneNumberField `bson:"inline"`
	UserPassword     `bson:"inline"`
	HoldingCode      string `json:"holdingcode,omitempty"`
}

type UserProfile struct {
	UsernameField     `bson:"inline"`
	Email             string `json:"email,omitempty" bson:"email,omitempty"`
	UserDetail        `bson:"inline"`
	UserPassword      `bson:"inline"`
	IsDefaultPassword bool `json:"isdefaultpassword" bson:"-"`

	// === ข้อมูล LINE (ระดับ user) ===
	LineUserID      string `json:"lineuserid" bson:"lineuserid"`
	LineDisplayName string `json:"linedisplayname" bson:"linedisplayname"`
	LinePictureURL  string `json:"linepictureurl" bson:"linepictureurl"`

	CreatedAt time.Time `json:"-" bson:"createdat,omitempty"`
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

const DefaultUserPassword = "12345"

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
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopUserBase     `bson:"inline"`
	IsFavorite       bool      `json:"isfavorite" bson:"isfavorite"`
	LastAccessedAt   time.Time `json:"lastaccessedat" bson:"lastaccessedat"`
	IsCreator        bool      `json:"iscreator,omitempty" bson:"-"`
	IsAccessDisabled bool      `json:"isaccessdisabled" bson:"isaccessdisabled"`
	AccessDisabledAt time.Time `json:"accessdisabledat,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy string    `json:"accessdisabledby,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time `json:"accessenabledat,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy  string    `json:"accessenabledby,omitempty" bson:"accessenabledby,omitempty"`

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
	HoldingCode string `json:"holdingcode" bson:"holdingcode"`
	Name        string `json:"name" bson:"name1"`
	// Name1          string         `json:"name1" bson:"name1"`
	MainHoldingCode     string           `json:"mainholdingcode" bson:"mainholdingcode"`
	Names               []models.NameX   `json:"names" bson:"names"`
	BranchCode          string           `json:"branchcode" bson:"branchcode"`
	Language            string           `json:"language" bson:"language"`
	LanguageConfigs     []LanguageConfig `json:"languageconfigs" bson:"languageconfigs"`
	BaseCurrency        string           `json:"basecurrency" bson:"basecurrency"`
	Currencies          []string         `json:"currencies" bson:"currencies"`
	Timezone            string           `json:"timezone" bson:"timezone"`
	TimezoneLabel       string           `json:"timezonelabel" bson:"timezonelabel"`
	TimezoneOffset      string           `json:"timezoneoffset" bson:"timezoneoffset"`
	DateFormat          string           `json:"dateformat" bson:"dateformat"`
	UseBuddhistCalendar bool             `json:"usebuddhistcalendar" bson:"usebuddhistcalendar"`
	Role                UserRole         `json:"role" bson:"role"`
	IsFavorite          bool             `json:"isfavorite" bson:"isfavorite"`
	LastAccessedAt      time.Time        `json:"lastaccessedat" bson:"lastaccessedat"`
	CreatedBy           string           `json:"createdby" bson:"createdby"`
	IsCreator           bool             `json:"iscreator,omitempty" bson:"-"`
	IsAccessDisabled    bool             `json:"isaccessdisabled" bson:"isaccessdisabled"`
	AccessDisabledAt    time.Time        `json:"accessdisabledat,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy    string           `json:"accessdisabledby,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt     time.Time        `json:"accessenabledat,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy     string           `json:"accessenabledby,omitempty" bson:"accessenabledby,omitempty"`
}

type LanguageConfig struct {
	Code           string `json:"code" bson:"code"`
	CodeTranslator string `json:"codetranslator" bson:"codetranslator"`
	Name           string `json:"name" bson:"name"`
	IsUse          bool   `json:"isuse" bson:"isuse"`
	IsDefault      bool   `json:"isdefault" bson:"isdefault"`
}

func (*ShopUserInfo) CollectionName() string {
	return "shopusers"
}

type UserRoleRequest struct {
	HoldingCode      string        `json:"holdingcode" bson:"holdingcode"`
	EditUsername     string        `json:"editusername" bson:"editusername"`
	Username         string        `json:"username" bson:"username"`
	UserUID          string        `json:"useruid,omitempty" bson:"useruid,omitempty"`
	UserProfileName  string        `json:"userprofilename" bson:"userprofilename"`
	Email            string        `json:"email,omitempty" bson:"email,omitempty"`
	// Avatar/AvatarThumb are pointers so LINE-sync/auto-unlink callers that omit them do not wipe stored values.
	Avatar           *string       `json:"avatar,omitempty" bson:"-"`
	AvatarThumb      *string       `json:"avatarthumb,omitempty" bson:"-"`
	Role             UserRole      `json:"role" bson:"role"`
	IsAccessDisabled bool          `json:"isaccessdisabled" bson:"isaccessdisabled"`
	AccessDisabledAt time.Time     `json:"accessdisabledat,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy string        `json:"accessdisabledby,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time     `json:"accessenabledat,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy  string        `json:"accessenabledby,omitempty" bson:"accessenabledby,omitempty"`
	AccessScopes     []AccessScope `json:"accessscopes,omitempty" bson:"accessscopes,omitempty"`

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
	ShopUserBase     `bson:"inline"`
	UID              string        `json:"uid,omitempty" bson:"uid,omitempty"`
	Email            string        `json:"email,omitempty" bson:"email,omitempty"`
	UserProfileName  string        `json:"userprofilename" bson:"userprofilename"`
	Avatar           string        `json:"avatar" bson:"avatar"`
	AvatarThumb      string        `json:"avatarthumb" bson:"avatarthumb"`
	IsCreator        bool          `json:"iscreator,omitempty" bson:"-"`
	IsAccessDisabled bool          `json:"isaccessdisabled" bson:"isaccessdisabled"`
	AccessDisabledAt time.Time     `json:"accessdisabledat,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy string        `json:"accessdisabledby,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time     `json:"accessenabledat,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy  string        `json:"accessenabledby,omitempty" bson:"accessenabledby,omitempty"`
	AccessScopes     []AccessScope `json:"accessscopes,omitempty" bson:"accessscopes,omitempty"`

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
}
