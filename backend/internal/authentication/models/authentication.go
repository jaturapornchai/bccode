package models

import "time"

type AuthenticationContext struct {
	Ip string
}

type ShopFavoriteRequest struct {
	HoldingCode string `json:"holdingcode" bson:"holdingcode"`
	IsFavorite  bool   `json:"isfavorite" bson:"isfavorite"`
}

type TokenLoginRequest struct {
	Token string `json:"token" validate:"required"`
}

type LineLoginRequest struct {
	Token string `json:"token" validate:"required"`
}

// LineUserLoginRequest — สำหรับ QR code / LIFF login flow
// Flutter sends lineuserid from LIFF server without LINE access token.
type LineUserLoginRequest struct {
	LineUserID  string `json:"lineuserid" validate:"required"`
	DisplayName string `json:"displayname"`
	PictureUrl  string `json:"pictureurl"`
	Email       string `json:"email"`
}

// GoogleLoginRequest — สำหรับ Google OAuth mobile (Android/iOS)
// Flutter ส่ง google_user_id + email หลังจาก Google OAuth สำเร็จ
type GoogleLoginRequest struct {
	GoogleUserID string `json:"googleuserid"`
	DisplayName  string `json:"displayname"`
	PictureUrl   string `json:"pictureurl"`
	Email        string `json:"email" validate:"required"`
}

type TokenLoginResponse struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}

type PhoneNumberLoginReponse struct {
	RefCode string    `json:"refcode"`
	Expire  time.Time `json:"expire"`
}

type PhoneNumberLoginRequest struct {
	PhoneNumber string `json:"phonenumber" bson:"phonenumber" validate:"required,max=233"`
}

type PhoneNumberOTPRequest struct {
	PhoneNumber string `json:"phonenumber" bson:"phonenumber" validate:"required,max=233"`
	RefCode     string `json:"refcode"`
	OTP         string `json:"otp" bson:"otp" validate:"required,max=20"`
}

type PhoneOTP struct {
	PhoneNumber string `json:"phonenumber"`
	OTP         string `json:"otp" `
}

// LinkLineRequest — สำหรับเชื่อมต่อ LINE กับ user profile
type LinkLineRequest struct {
	LineUserID      string `json:"lineuserid" validate:"required"`
	LineDisplayName string `json:"linedisplayname"`
	LinePictureURL  string `json:"linepictureurl"`
}

type UserDisableLoginError struct{}

func (e *UserDisableLoginError) Error() string {
	return "user is disabled"
}
