package lineoa

import "time"

// MongoDB Collections for Line OA
const (
	ConfigCollection      = "lineoa_configs"
	EmployeeCollection    = "lineoa_employees"
	LinkTokenCollection   = "lineoa_link_tokens"
	UserProfileCollection = "user_line_profiles"
)

// ConfigDoc - MongoDB document for Line OA config
type ConfigDoc struct {
	GUID            string     `bson:"guid" json:"guid"`
	HoldingCode     string     `bson:"holdingcode" json:"holdingcode"`
	LineOAType      string     `bson:"lineoatype" json:"lineoatype"`
	ChannelID       string     `bson:"channelid" json:"channelid"`
	ChannelSecret   string     `bson:"channelsecret" json:"channelsecret"`
	AccessToken     string     `bson:"accesstoken" json:"accesstoken"`
	LiffID          string     `bson:"liffid" json:"liffid"`
	IsActive        bool       `bson:"isactive" json:"isactive"`
	LastTestAt      *time.Time `bson:"lasttestat,omitempty" json:"lasttestat"`
	LastTestSuccess bool       `bson:"lasttestsuccess" json:"lasttestsuccess"`
	LastTestMessage string     `bson:"lasttestmessage,omitempty" json:"lasttestmessage"`
	CreatedAt       time.Time  `bson:"createdat" json:"createdat"`
	UpdatedAt       time.Time  `bson:"updatedat" json:"updatedat"`
}

// EmployeeDoc - MongoDB document for Line OA employee
type EmployeeDoc struct {
	GUID             string     `bson:"guid" json:"guid"`
	LineOAConfigGUID string     `bson:"lineoaconfigguid" json:"lineoaconfigguid"`
	HoldingCode      string     `bson:"holdingcode" json:"holdingcode"`
	EmployeeCode     string     `bson:"employeecode" json:"employeecode"`
	EmployeeName     string     `bson:"employeename" json:"employeename"`
	LineUserID       string     `bson:"lineuserid" json:"lineuserid"`
	DisplayName      string     `bson:"displayname,omitempty" json:"displayname"`
	PictureURL       string     `bson:"pictureurl,omitempty" json:"pictureurl"`
	IsActive         bool       `bson:"isactive" json:"isactive"`
	LinkedAt         *time.Time `bson:"linkedat,omitempty" json:"linkedat"`
	CreatedAt        time.Time  `bson:"createdat" json:"createdat"`
}

// LinkTokenDoc - MongoDB document for link token
type LinkTokenDoc struct {
	Token            string     `bson:"token" json:"token"`
	HoldingCode      string     `bson:"holdingcode" json:"holdingcode"`
	LineOAConfigGUID string     `bson:"lineoaconfigguid,omitempty" json:"lineoaconfigguid"`
	EmployeeCode     string     `bson:"employeecode,omitempty" json:"employeecode"`
	Username         string     `bson:"username,omitempty" json:"username"`
	TokenType        string     `bson:"tokentype" json:"tokentype"` // "employee" or "user"
	ExpiresAt        time.Time  `bson:"expiresat" json:"expiresat"`
	UsedAt           *time.Time `bson:"usedat,omitempty" json:"usedat"`
	CreatedAt        time.Time  `bson:"createdat" json:"createdat"`
}

// Request types
type (
	// ConfigRequest - Request for Line OA config operations
	ConfigRequest struct {
		HoldingCode string `json:"holdingcode"`
		LineOAType  string `json:"lineoatype"`
	}

	// SaveConfigRequest - Request for saving Line OA config
	SaveConfigRequest struct {
		HoldingCode   string `json:"holdingcode"`
		GUID          string `json:"guid"`
		LineOAType    string `json:"lineoatype"`
		ChannelID     string `json:"channelid"`
		ChannelSecret string `json:"channelsecret"`
		AccessToken   string `json:"accesstoken"`
		LiffID        string `json:"liffid"`
		IsActive      bool   `json:"isactive"`
	}

	// TestRequest - Request for testing Line OA connection
	TestRequest struct {
		HoldingCode   string `json:"holdingcode"`
		ChannelID     string `json:"channelid"`
		ChannelSecret string `json:"channelsecret"`
		AccessToken   string `json:"accesstoken"`
	}

	// EmployeeRequest - Request for employee operations
	EmployeeRequest struct {
		HoldingCode      string `json:"holdingcode"`
		LineOAConfigGUID string `json:"lineoaconfigguid"`
	}

	// AddEmployeeRequest - Request for adding employee
	AddEmployeeRequest struct {
		HoldingCode      string `json:"holdingcode"`
		LineOAConfigGUID string `json:"lineoaconfigguid"`
		EmployeeCode     string `json:"employeecode"`
		EmployeeName     string `json:"employeename"`
	}

	// RemoveEmployeeRequest - Request for removing employee
	RemoveEmployeeRequest struct {
		HoldingCode string `json:"holdingcode"`
		GUID        string `json:"guid"`
	}

	// GenerateLinkRequest - Request for generating employee link
	GenerateLinkRequest struct {
		HoldingCode      string `json:"holdingcode"`
		LineOAConfigGUID string `json:"lineoaconfigguid"`
		EmployeeCode     string `json:"employeecode"`
	}

	// UserLinkRequest - Request for generating user Line OA link
	UserLinkRequest struct {
		HoldingCode string `json:"holdingcode"`
		Username    string `json:"username"`
	}

	// CallbackRequest - Callback from LIFF after user links
	CallbackRequest struct {
		HoldingCode string `json:"holdingcode"`
		Token       string `json:"token"`
		LineUserID  string `json:"lineuserid"`
		DisplayName string `json:"displayname"`
		PictureURL  string `json:"pictureurl"`
	}

	// UserProfileRequest - Request for user LINE profile
	UserProfileRequest struct {
		HoldingCode string `json:"holdingcode"`
		Username    string `json:"username"`
	}
)
