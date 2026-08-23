package models

type UserInfo struct {
	Username          string `json:"username" `
	Name              string `json:"name"`
	HoldingCode       string `json:"holdingcode" `
	BusinessCode      string `json:"businesscode"`
	Role              uint8  `json:"role"`
	UID               string `json:"uid"`
	SessionUID        string `json:"-"`
	MembershipUID     string `json:"-"`
	HoldingUID        string `json:"-"`
	CompanyUID        string `json:"-"`
	BranchUID         string `json:"-"`
	PermissionVersion int64  `json:"-"`
}
