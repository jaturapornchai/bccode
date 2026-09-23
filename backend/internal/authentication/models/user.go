package models

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"smlcloudplatform/internal/models"
	timezone "smlcloudplatform/internal/models/timezone"
)

type UserDetail struct {
	UID         string `json:"uid"`
	Name        string `json:"name,omitempty"`
	Avatar      string `json:"avatar"`
	AvatarThumb string `json:"avatarthumb"`
	timezone.Timezone
	YearType     string   `json:"yeartype" validate:"max=21"`
	DedeZoom     DedeZoom `json:"dedezoom"`
	RegisterType string   `json:"registertype"`
}

type DedeZoom struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phonenumber"`
	Address     string `json:"address"`
}

type UsernameField struct {
	Username string `json:"username,omitempty" validate:"required,gte=3,max=64"`
}

type PhoneNumberField struct {
	CountryCode string `json:"countrycode" validate:"required,max=20"`
	PhoneNumber string `json:"phonenumber" validate:"required,max=100"`
}

type EmailField struct {
	Email string `json:"email,omitempty" validate:"required,email,max=233"`
}

type UserPassword struct {
	Password string `json:"password,omitempty" validate:"required,gte=15,max=64"`
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
	ID        string `json:"-"`
	GuidFixed string `json:"guidfixed"`
	UsernameField
	EmailField
	PhoneNumberField
	UserPassword
	UserDetail

	// === ข้อมูล LINE (ระดับ user — ใช้ร่วมทุก shop) ===
	LineUserID      string `json:"lineuserid"`      // LINE User ID
	LineDisplayName string `json:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl"`  // LINE Profile Picture URL

	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
	DisabledAt time.Time `json:"disabledat,omitempty"`
	IsDeleted  bool      `json:"isdeleted"`
	Version    int64     `json:"-"`
}

type GoogleIdentity struct {
	ID            string     `json:"-"`
	IdentityUID   string     `json:"identityuid"`
	UserUID       string     `json:"useruid"`
	Issuer        string     `json:"issuer"`
	Subject       string     `json:"subject"`
	VerifiedEmail string     `json:"verifiedemail"`
	IsActive      bool       `json:"isactive"`
	LinkedAt      time.Time  `json:"linkedat"`
	RevokedAt     *time.Time `json:"revokedat,omitempty"`
	RevokedBy     string     `json:"revokedby,omitempty"`
}

type AuthAudit struct {
	ID         string                 `json:"-"`
	AuditUID   string                 `json:"audituid"`
	UserUID    string                 `json:"useruid,omitempty"`
	Action     string                 `json:"action"`
	Outcome    string                 `json:"outcome"`
	ReasonCode string                 `json:"reasoncode,omitempty"`
	SessionUID string                 `json:"sessionuid,omitempty"`
	OccurredAt time.Time              `json:"occurredat"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
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

// RegisterUsernameRequest — สำหรับพนักงานสมัครด้วยรหัสพนักงาน + รหัสผ่าน (ไม่ต้องมี email)
type RegisterUsernameRequest struct {
	UsernameField
	UserPassword
	UserDetail
}

type UserRequest struct {
	UsernameField
	UserPassword
	UserDetail
}

type UserLoginRequest struct {
	UsernameField
	UserPassword
	HoldingCode string `json:"holdingcode,omitempty"`
}

type PosLoginRequest struct {
	UsernameField
	HoldingCode string `json:"holdingcode,omitempty"`
}

type UserLoginPhoneNumberRequest struct {
	PhoneNumberField
	UserPassword
	HoldingCode string `json:"holdingcode,omitempty"`
}

type UserProfile struct {
	UsernameField
	Email string `json:"email,omitempty"`
	UserDetail
	UserPassword

	// === ข้อมูล LINE (ระดับ user) ===
	LineUserID      string `json:"lineuserid"`
	LineDisplayName string `json:"linedisplayname"`
	LinePictureURL  string `json:"linepictureurl"`

	CreatedAt time.Time `json:"-"`
}

type UserProfileRequest struct {
	UserDetail
}

type UserPasswordRequest struct {
	CurrentPassword string `json:"currentpassword" validate:"required,gte=15,max=64"`
	NewPassword     string `json:"newpassword" validate:"required,gte=15,max=64"`
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
	MembershipUID string   `json:"membershipuid"`
	Username      string   `json:"username"`
	UserUID       string   `json:"useruid"`
	HoldingUID    string   `json:"holdinguid"`
	HoldingCode   string   `json:"holdingcode"`
	Role          UserRole `json:"role"`
}

// DocumentApproval - ข้อมูลการอนุมัติแยกตามประเภทเอกสาร
type DocumentApproval struct {
	ApprovalRole      int     `json:"approvalrole"`      // 0=ไม่มีสิทธิ์, 1-4=ระดับผู้อนุมัติ
	MaxApprovalAmount float64 `json:"maxapprovalamount"` // วงเงินอนุมัติสูงสุด (บาท)
}

type AccessScope struct {
	ScopeType    string `json:"scopetype"`              // company, branch
	CompanyUID   string `json:"companyuid,omitempty"`   // immutable company id
	BranchUID    string `json:"branchuid,omitempty"`    // immutable branch id
	BusinessCode string `json:"businesscode,omitempty"` // legacy display code
	BranchCode   string `json:"branchcode,omitempty"`   // legacy display code
	AllBranches  bool   `json:"allbranches,omitempty"`
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
	ID      string `json:"id"`
	Version int64  `json:"-"`
	ShopUserBase
	PermissionVersion int64     `json:"permissionversion"`
	IsDeleted         bool      `json:"isdeleted"`
	CreatedAt         time.Time `json:"createdat"`
	CreatedBy         string    `json:"createdby"`
	IsFavorite        bool      `json:"isfavorite"`
	LastAccessedAt    time.Time `json:"lastaccessedat"`
	IsCreator         bool      `json:"iscreator,omitempty"`
	IsAccessDisabled  bool      `json:"isaccessdisabled"`
	AccessDisabledAt  time.Time `json:"accessdisabledat,omitempty"`
	AccessDisabledBy  string    `json:"accessdisabledby,omitempty"`
	AccessEnabledAt   time.Time `json:"accessenabledat,omitempty"`
	AccessEnabledBy   string    `json:"accessenabledby,omitempty"`
	// AccessExpiryDate auto-blocks access once the date is reached (offboarding /
	// last working day). Zero = no expiry. The shop creator is always exempt.
	AccessExpiryDate time.Time `json:"accessexpirydate,omitempty"`

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position"`   // ตำแหน่งงาน
	Department string `json:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"lineuserid"`      // LINE User ID
	LineDisplayName string `json:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl"`  // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"poapproval,omitempty"`        // อนุมัติใบสั่งซื้อ
	QuotationApproval *DocumentApproval `json:"quotationapproval,omitempty"` // อนุมัติใบเสนอราคา
	AccessScopes      []AccessScope     `json:"accessscopes,omitempty"`
	PermissionSets    []string          `json:"permissionsets"` // ชุดสิทธิ์ (role_permission.rolecode) เลือกได้หลายชุด
}

type ShopUserInfo struct {
	HoldingUID  string `json:"holdinguid"`
	HoldingCode string `json:"holdingcode"`
	Name        string `json:"name"`
	// Name1          string         `json:"name1"`
	MainHoldingCode     string           `json:"mainholdingcode"`
	Names               []models.NameX   `json:"names"`
	BranchCode          string           `json:"branchcode"`
	Language            string           `json:"language"`
	LanguageConfigs     []LanguageConfig `json:"languageconfigs"`
	BaseCurrency        string           `json:"basecurrency"`
	Currencies          []string         `json:"currencies"`
	Timezone            string           `json:"timezone"`
	TimezoneLabel       string           `json:"timezonelabel"`
	TimezoneOffset      string           `json:"timezoneoffset"`
	DateFormat          string           `json:"dateformat"`
	UseBuddhistCalendar bool             `json:"usebuddhistcalendar"`
	Role                UserRole         `json:"role"`
	IsFavorite          bool             `json:"isfavorite"`
	LastAccessedAt      time.Time        `json:"lastaccessedat"`
	CreatedBy           string           `json:"createdby"`
	IsCreator           bool             `json:"iscreator,omitempty"`
	IsAccessDisabled    bool             `json:"isaccessdisabled"`
	AccessDisabledAt    time.Time        `json:"accessdisabledat,omitempty"`
	AccessDisabledBy    string           `json:"accessdisabledby,omitempty"`
	AccessEnabledAt     time.Time        `json:"accessenabledat,omitempty"`
	AccessEnabledBy     string           `json:"accessenabledby,omitempty"`
}

type LanguageConfig struct {
	Code           string `json:"code"`
	CodeTranslator string `json:"codetranslator"`
	Name           string `json:"name"`
	IsUse          bool   `json:"isuse"`
	IsDefault      bool   `json:"isdefault"`
}

type UserRoleRequest struct {
	HoldingCode     string `json:"holdingcode"`
	EditUsername    string `json:"editusername"`
	Username        string `json:"username"`
	UserUID         string `json:"useruid,omitempty"`
	UserProfileName string `json:"userprofilename"`
	Email           string `json:"email,omitempty"`
	// Avatar/AvatarThumb are pointers so LINE-sync/auto-unlink callers that omit them do not wipe stored values.
	Avatar           *string       `json:"avatar,omitempty"`
	AvatarThumb      *string       `json:"avatarthumb,omitempty"`
	Role             UserRole      `json:"role"`
	IsAccessDisabled bool          `json:"isaccessdisabled"`
	AccessDisabledAt time.Time     `json:"accessdisabledat,omitempty"`
	AccessDisabledBy string        `json:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time     `json:"accessenabledat,omitempty"`
	AccessEnabledBy  string        `json:"accessenabledby,omitempty"`
	AccessExpiryDate time.Time     `json:"accessexpirydate,omitempty"`
	AccessScopes     []AccessScope `json:"accessscopes,omitempty"`
	PermissionSets   []string      `json:"permissionsets"` // ชุดสิทธิ์ (role_permission.rolecode) เลือกได้หลายชุด

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position"`   // ตำแหน่งงาน
	Department string `json:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"lineuserid"`      // LINE User ID
	LineDisplayName string `json:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl"`  // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"poapproval,omitempty"`
	QuotationApproval *DocumentApproval `json:"quotationapproval,omitempty"`
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
	ID             string    `json:"id"`
	HoldingCode    string    `json:"holdingcode"`
	BusinessCode   string    `json:"businesscode,omitempty"`
	Username       string    `json:"username"`
	Ip             string    `json:"ip"`
	LastAccessedAt time.Time `json:"lastaccessedat"`
}

type ShopUserProfile struct {
	ShopUserBase
	UID              string        `json:"uid,omitempty"`
	Email            string        `json:"email,omitempty"`
	UserProfileName  string        `json:"userprofilename"`
	Avatar           string        `json:"avatar"`
	AvatarThumb      string        `json:"avatarthumb"`
	IsCreator        bool          `json:"iscreator,omitempty"`
	IsAccessDisabled bool          `json:"isaccessdisabled"`
	AccessDisabledAt time.Time     `json:"accessdisabledat,omitempty"`
	AccessDisabledBy string        `json:"accessdisabledby,omitempty"`
	AccessEnabledAt  time.Time     `json:"accessenabledat,omitempty"`
	AccessEnabledBy  string        `json:"accessenabledby,omitempty"`
	AccessScopes     []AccessScope `json:"accessscopes,omitempty"`
	PermissionSets   []string      `json:"permissionsets"` // ชุดสิทธิ์ (role_permission.rolecode) เลือกได้หลายชุด

	// === ข้อมูลพนักงาน ===
	Position   string `json:"position"`   // ตำแหน่งงาน
	Department string `json:"department"` // แผนก

	// === ข้อมูล LINE OA ===
	LineUserID      string `json:"lineuserid"`      // LINE User ID
	LineDisplayName string `json:"linedisplayname"` // LINE Display Name
	LinePictureURL  string `json:"linepictureurl"`  // LINE Profile Picture URL

	// === ข้อมูลการอนุมัติแยกตามประเภทเอกสาร ===
	POApproval        *DocumentApproval `json:"poapproval,omitempty"`
	QuotationApproval *DocumentApproval `json:"quotationapproval,omitempty"`
}
