package lineoa

import "time"

// MongoDB Collections for Line OA
const (
	ConfigCollection    = "lineoa_configs"
	EmployeeCollection  = "lineoa_employees"
	LinkTokenCollection = "lineoa_link_tokens"
	UserProfileCollection = "user_line_profiles"
)

// ConfigDoc - MongoDB document for Line OA config
type ConfigDoc struct {
	GUID            string     `bson:"guid" json:"guid"`
	ShopID          string     `bson:"shop_id" json:"shop_id"`
	LineOAType      string     `bson:"lineoa_type" json:"lineoa_type"`
	ChannelID       string     `bson:"channel_id" json:"channel_id"`
	ChannelSecret   string     `bson:"channel_secret" json:"channel_secret"`
	AccessToken     string     `bson:"access_token" json:"access_token"`
	LiffID          string     `bson:"liff_id" json:"liff_id"`
	IsActive        bool       `bson:"is_active" json:"is_active"`
	LastTestAt      *time.Time `bson:"last_test_at,omitempty" json:"last_test_at"`
	LastTestSuccess bool       `bson:"last_test_success" json:"last_test_success"`
	LastTestMessage string     `bson:"last_test_message,omitempty" json:"last_test_message"`
	CreatedAt       time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `bson:"updated_at" json:"updated_at"`
}

// EmployeeDoc - MongoDB document for Line OA employee
type EmployeeDoc struct {
	GUID             string     `bson:"guid" json:"guid"`
	LineOAConfigGUID string     `bson:"lineoa_config_guid" json:"lineoa_config_guid"`
	ShopID           string     `bson:"shop_id" json:"shop_id"`
	EmployeeCode     string     `bson:"employee_code" json:"employee_code"`
	EmployeeName     string     `bson:"employee_name" json:"employee_name"`
	LineUserID       string     `bson:"line_user_id" json:"line_user_id"`
	DisplayName      string     `bson:"display_name,omitempty" json:"display_name"`
	PictureURL       string     `bson:"picture_url,omitempty" json:"picture_url"`
	IsActive         bool       `bson:"is_active" json:"is_active"`
	LinkedAt         *time.Time `bson:"linked_at,omitempty" json:"linked_at"`
	CreatedAt        time.Time  `bson:"created_at" json:"created_at"`
}

// LinkTokenDoc - MongoDB document for link token
type LinkTokenDoc struct {
	Token            string     `bson:"token" json:"token"`
	ShopID           string     `bson:"shop_id" json:"shop_id"`
	LineOAConfigGUID string     `bson:"lineoa_config_guid,omitempty" json:"lineoa_config_guid"`
	EmployeeCode     string     `bson:"employee_code,omitempty" json:"employee_code"`
	Username         string     `bson:"username,omitempty" json:"username"`
	TokenType        string     `bson:"token_type" json:"token_type"` // "employee" or "user"
	ExpiresAt        time.Time  `bson:"expires_at" json:"expires_at"`
	UsedAt           *time.Time `bson:"used_at,omitempty" json:"used_at"`
	CreatedAt        time.Time  `bson:"created_at" json:"created_at"`
}

// Request types
type (
	// ConfigRequest - Request for Line OA config operations
	ConfigRequest struct {
		ShopID     string `json:"shop_id"`
		LineOAType string `json:"lineoa_type"`
	}

	// SaveConfigRequest - Request for saving Line OA config
	SaveConfigRequest struct {
		ShopID        string `json:"shop_id"`
		GUID          string `json:"guid"`
		LineOAType    string `json:"lineoa_type"`
		ChannelID     string `json:"channel_id"`
		ChannelSecret string `json:"channel_secret"`
		AccessToken   string `json:"access_token"`
		LiffID        string `json:"liff_id"`
		IsActive      bool   `json:"is_active"`
	}

	// TestRequest - Request for testing Line OA connection
	TestRequest struct {
		ShopID        string `json:"shop_id"`
		ChannelID     string `json:"channel_id"`
		ChannelSecret string `json:"channel_secret"`
		AccessToken   string `json:"access_token"`
	}

	// EmployeeRequest - Request for employee operations
	EmployeeRequest struct {
		ShopID           string `json:"shop_id"`
		LineOAConfigGUID string `json:"lineoa_config_guid"`
	}

	// AddEmployeeRequest - Request for adding employee
	AddEmployeeRequest struct {
		ShopID           string `json:"shop_id"`
		LineOAConfigGUID string `json:"lineoa_config_guid"`
		EmployeeCode     string `json:"employee_code"`
		EmployeeName     string `json:"employee_name"`
	}

	// RemoveEmployeeRequest - Request for removing employee
	RemoveEmployeeRequest struct {
		ShopID string `json:"shop_id"`
		GUID   string `json:"guid"`
	}

	// GenerateLinkRequest - Request for generating employee link
	GenerateLinkRequest struct {
		ShopID           string `json:"shop_id"`
		LineOAConfigGUID string `json:"lineoa_config_guid"`
		EmployeeCode     string `json:"employee_code"`
	}

	// UserLinkRequest - Request for generating user Line OA link
	UserLinkRequest struct {
		ShopID   string `json:"shop_id"`
		Username string `json:"username"`
	}

	// CallbackRequest - Callback from LIFF after user links
	CallbackRequest struct {
		ShopID      string `json:"shop_id"`
		Token       string `json:"token"`
		LineUserID  string `json:"line_user_id"`
		DisplayName string `json:"display_name"`
		PictureURL  string `json:"picture_url"`
	}

	// UserProfileRequest - Request for user LINE profile
	UserProfileRequest struct {
		ShopID   string `json:"shop_id"`
		Username string `json:"username"`
	}
)
