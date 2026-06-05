package models

type OTPRequest struct {
	PhoneNumberField
}

type OTPResponse struct {
	OTPToken   string `json:"otptoken"`
	OTPRefCode string `json:"otprefcode"`
}

type OTPVerifyRequest struct {
	OTPToken   string `json:"otptoken"`
	OTPRefCode string `json:"otprefcode"`
	OTPPin     string `json:"otppin"`
}
