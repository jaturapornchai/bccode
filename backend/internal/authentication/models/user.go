package models

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

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
	Username string `json:"username,omitempty" bson:"username,omitempty" validate:"required,gte=3,max=64"`
}

type PhoneNumberField struct {
	CountryCode string `json:"countrycode" bson:"countrycode" validate:"required,max=20"`
	PhoneNumber string `json:"phonenumber" bson:"phonenumber" validate:"required,max=100"`
}

type EmailField struct {
	Email string `json:"email,omitempty" bson:"email" validate:"required,email,max=233"`
}

type UserPassword struct {
	Password string `json:"password,omitempty" bson:"password,omitempty" validate:"required,gte=15,max=64"`
}

var usercodePattern = regexp.MustCompile(`^[a-z0-9._-]+$`)

func NormalizeUsercode(usercode string) string {
	return strings.ToLower(strings.TrimSpace(usercode))
}

func IsValidUsercode(usercode string) bool {
	usercode = NormalizeUsercode(usercode)
	length := utf8.RuneCountInString(usercode)
	return length >= 3 && length <= 64 && usercodePattern.MatchString(usercode)
}

func IsValidPasswordLength(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= 15 && length <= 64
}

var knownCompromisedPasswords = map[string]struct{}{
	"123456789012345":  {},
	"passwordpassword": {},
	"password123456":   {},
	"qwertyuiop12345":  {},
	"adminadmin12345":  {},
	"letmeinletmein":   {},
	"iloveyouiloveyou": {},
}

func IsKnownCompromisedPassword(password string) bool {
	_, found := knownCompromisedPasswords[strings.ToLower(strings.TrimSpace(password))]
	return found
}

type UserDoc struct {
	ID               primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	GuidFixed        string             `json:"guidfixed" bson:"guidfixed"`
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
	IsDeleted  bool      `json:"isdeleted" bson:"isdeleted"`
	Version    int64     `json:"-" bson:"__v"`
}

type GoogleIdentity struct {
	ID            primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	IdentityUID   string             `json:"identityuid" bson:"identityuid"`
	UserUID       string             `json:"useruid" bson:"useruid"`
	Issuer        string             `json:"issuer" bson:"issuer"`
	Subject       string             `json:"subject" bson:"subject"`
	VerifiedEmail string             `json:"verifiedemail" bson:"verifiedemail"`
	IsActive      bool               `json:"isactive" bson:"isactive"`
	LinkedAt      time.Time          `json:"linkedat" bson:"linkedat"`
	RevokedAt     *time.Time         `json:"revokedat,omitempty" bson:"revokedat,omitempty"`
	RevokedBy     string             `json:"revokedby,omitempty" bson:"revokedby,omitempty"`
}

func (*GoogleIdentity) CollectionName() string {
	return "googleidentities"
}

type AuthAudit struct {
	ID         primitive.ObjectID     `json:"-" bson:"_id,omitempty"`
	AuditUID   string                 `json:"audituid" bson:"audituid"`
	UserUID    string                 `json:"useruid,omitempty" bson:"useruid,omitempty"`
	Action     string                 `json:"action" bson:"action"`
	Outcome    string                 `json:"outcome" bson:"outcome"`
	ReasonCode string                 `json:"reasoncode,omitempty" bson:"reasoncode,omitempty"`
	SessionUID string                 `json:"sessionuid,omitempty" bson:"sessionuid,omitempty"`
	OccurredAt time.Time              `json:"occurredat" bson:"occurredat"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func (*AuthAudit) CollectionName() string {
	return "authaudits"
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
	UsernameField `bson:"inline"`
	Email         string `json:"email,omitempty" bson:"email,omitempty"`
	UserDetail    `bson:"inline"`
	UserPassword  `bson:"inline"`

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
	CurrentPassword string `json:"currentpassword" bson:"currentpassword" validate:"required,gte=15,max=64"`
	NewPassword     string `json:"newpassword" bson:"newpassword" validate:"required,gte=15,max=64"`
}

type UserProfileReponse struct {
	Success bool        `json:"success"`
	Data    UserProfile `json:"data"`
}

type ShopSelectRequest struct {
	HoldingCode  string `json:"holdingcode"`
	BusinessCode string `json:"businesscode,omitempty"`
	BranchUID    string `json:"branchuid,omitempty"`
}

type UserRole = uint8

const (
	ROLE_USER  UserRole = iota // "USER"
	ROLE_ADMIN                 // "ADMIN"
	ROLE_OWNER                 // "OWNER"

	ROLE_SYSTEM = 255 // APP MANAGER
)

type ShopUserBase struct {
	MembershipUID string   `json:"membershipuid" bson:"membershipuid"`
	Username      string   `json:"username" bson:"username"`
	UserUID       string   `json:"useruid" bson:"useruid"`
	HoldingUID    string   `json:"holdinguid" bson:"holdinguid"`
	HoldingCode   string   `json:"holdingcode" bson:"holdingcode"`
	Role          UserRole `json:"role" bson:"role"`
}

// DocumentApproval - ข้อมูลการอนุมัติแยกตามประเภทเอกสาร
type DocumentApproval struct {
	ApprovalRole      int     `json:"approvalrole" bson:"approvalrole"`           // 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	MaxApprovalAmount float64 `json:"maxapprovalamount" bson:"maxapprovalamount"` // วงเงินอนุมัติสูงสุด (บาท)
}

type AccessScope struct {
	ScopeType    string `json:"scopetype" bson:"scopetype"`                           // company, branch
	CompanyUID   string `json:"companyuid,omitempty" bson:"companyuid,omitempty"`     // immutable company id
	BranchUID    string `json:"branchuid,omitempty" bson:"branchuid,omitempty"`       // immutable branch id
	BusinessCode string `json:"businesscode,omitempty" bson:"businesscode,omitempty"` // legacy display code
	BranchCode   string `json:"branchcode,omitempty" bson:"branchcode,omitempty"`     // legacy display code
	AllBranches  bool   `json:"allbranches,omitempty" bson:"allbranches,omitempty"`
}

// ScopesAllow reports whether the given access scopes permit access to a company (businessCode)
// and optionally a branch. EMPTY scopes fail closed; a Holding role never grants
// transaction access by itself.
func ScopesAllow(scopes []AccessScope, businessCode string, branchCode string) bool {
	if len(scopes) == 0 {
		return false
	}
	bc := strings.ToUpper(strings.TrimSpace(businessCode))
	brc := strings.ToUpper(strings.TrimSpace(branchCode))
	for _, s := range scopes {
		st := strings.ToLower(strings.TrimSpace(s.ScopeType))
		if st != "company" && st != "branch" {
			continue
		}
		sbc := strings.ToUpper(strings.TrimSpace(s.BusinessCode))
		if sbc == "" || sbc != bc {
			continue
		}
		// Company-level (or "all branches") grants the whole company. A bare company check
		// (branchCode == "") also passes here.
		if st == "company" || s.AllBranches || brc == "" {
			return true
		}
		if st == "branch" && strings.ToUpper(strings.TrimSpace(s.BranchCode)) == brc {
			return true
		}
	}
	return false
}

// ScopesAllowCompanySelection reports whether a user may enter a company-wide
// session by immutable CompanyUID. Branch-only scopes and legacy business codes
// must not be promoted to company-wide access.
func ScopesAllowCompanySelection(scopes []AccessScope, companyUID string) bool {
	if len(scopes) == 0 {
		return false
	}
	companyUID = strings.TrimSpace(companyUID)
	if companyUID == "" {
		return false
	}
	for _, scope := range scopes {
		scopeType := strings.ToLower(strings.TrimSpace(scope.ScopeType))
		if scopeType != "company" {
			continue
		}
		if strings.TrimSpace(scope.CompanyUID) != companyUID {
			continue
		}
		if strings.TrimSpace(scope.BranchUID) == "" {
			return true
		}
	}
	return false
}

// ScopesAllowBranchSelection permits an exact Branch scope or a Company scope
// explicitly covering every Branch. It never promotes a Branch scope to a
// Company-wide workspace.
func ScopesAllowBranchSelection(scopes []AccessScope, companyUID, branchUID string) bool {
	companyUID = strings.TrimSpace(companyUID)
	branchUID = strings.TrimSpace(branchUID)
	if companyUID == "" || branchUID == "" {
		return false
	}
	for _, scope := range scopes {
		if strings.TrimSpace(scope.CompanyUID) != companyUID {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(scope.ScopeType)) {
		case "company":
			if scope.AllBranches {
				return true
			}
		case "branch":
			if strings.TrimSpace(scope.BranchUID) == branchUID {
				return true
			}
		}
	}
	return false
}

type ShopUser struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Version           int64              `json:"-" bson:"__v"`
	ShopUserBase      `bson:"inline"`
	PermissionVersion int64     `json:"permissionversion" bson:"permissionversion"`
	IsDeleted         bool      `json:"isdeleted" bson:"isdeleted"`
	CreatedAt         time.Time `json:"createdat" bson:"createdat"`
	CreatedBy         string    `json:"createdby" bson:"createdby"`
	IsFavorite        bool      `json:"isfavorite" bson:"isfavorite"`
	LastAccessedAt    time.Time `json:"lastaccessedat" bson:"lastaccessedat"`
	IsCreator         bool      `json:"iscreator,omitempty" bson:"-"`
	IsAccessDisabled  bool      `json:"isaccessdisabled" bson:"isaccessdisabled"`
	AccessDisabledAt  time.Time `json:"accessdisabledat,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy  string    `json:"accessdisabledby,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt   time.Time `json:"accessenabledat,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy   string    `json:"accessenabledby,omitempty" bson:"accessenabledby,omitempty"`
	// AccessExpiryDate auto-blocks access once the date is reached (offboarding /
	// last working day). Zero = no expiry. The shop creator is always exempt.
	AccessExpiryDate time.Time `json:"accessexpirydate,omitempty" bson:"accessexpirydate,omitempty"`

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
	HoldingUID  string `json:"holdinguid" bson:"holdinguid"`
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
	HoldingCode     string `json:"holdingcode" bson:"holdingcode"`
	EditUsername    string `json:"editusername" bson:"editusername"`
	Username        string `json:"username" bson:"username"`
	UserUID         string `json:"useruid,omitempty" bson:"useruid,omitempty"`
	UserProfileName string `json:"userprofilename" bson:"userprofilename"`
	Email           string `json:"email,omitempty" bson:"email,omitempty"`
	// Avatar/AvatarThumb are pointers so LINE-sync/auto-unlink callers that omit them do not wipe stored values.
	Avatar           *string       `json:"avatar,omitempty" bson:"-"`
	AvatarThumb      *string       `json:"avatarthumb,omitempty" bson:"-"`
	Role             UserRole      `json:"role" bson:"role"`
	IsAccessDisabled bool          `json:"isaccessdisabled" bson:"isaccessdisabled"`
	AccessDisabledAt time.Time     `json:"accessdisabledat,omitempty" bson:"accessdisabledat,omitempty"`
	AccessDisabledBy string        `json:"accessdisabledby,omitempty" bson:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time     `json:"accessenabledat,omitempty" bson:"accessenabledat,omitempty"`
	AccessEnabledBy  string        `json:"accessenabledby,omitempty" bson:"accessenabledby,omitempty"`
	AccessExpiryDate time.Time     `json:"accessexpirydate,omitempty" bson:"accessexpirydate,omitempty"`
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

// UnmarshalJSON accepts an empty expiry from optional date inputs as no expiry.
func (req *UserRoleRequest) UnmarshalJSON(data []byte) error {
	type requestAlias UserRoleRequest
	decoded := struct {
		AccessExpiryDate json.RawMessage `json:"accessexpirydate"`
		*requestAlias
	}{requestAlias: (*requestAlias)(req)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	rawExpiry := strings.TrimSpace(string(decoded.AccessExpiryDate))
	if rawExpiry == "" || rawExpiry == "null" || rawExpiry == `""` {
		req.AccessExpiryDate = time.Time{}
		return nil
	}
	return json.Unmarshal(decoded.AccessExpiryDate, &req.AccessExpiryDate)
}

type ShopUserAccessLog struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode    string             `json:"holdingcode" bson:"holdingcode"`
	BusinessCode   string             `json:"businesscode,omitempty" bson:"businesscode,omitempty"`
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
