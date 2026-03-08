package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IAuthenticationSMSRepository interface {
	SendOTP(phoneNumber string, refCode string, otpCode string, expire time.Duration) error
	VerifyOTP(refCode string, otpCode string) (bool, error)

	SendOTPViaLink(fullPhoneNumber string) (models.OTPResponse, error)
	VerifyOTPViaLink(otpToken, optRefCode, otpPin string) (bool, error)
}

type AuthenticationSMSRepository struct {
	cache microservice.ICacher
}

func NewAuthenticationSMSRepository(cache microservice.ICacher) *AuthenticationSMSRepository {
	return &AuthenticationSMSRepository{
		cache: cache,
	}
}

func (repo AuthenticationSMSRepository) SendSMS(phoneNumber string, message string, expire time.Duration) error {
	return nil
}

func (repo AuthenticationSMSRepository) SendOTP(phoneNumber string, refCode string, otpCode string, expire time.Duration) error {
	messageOTP := fmt.Sprintf("OTP: %s , (ref code: %s) expire in 1 minute ", otpCode, refCode)

	err := repo.SendSMS(phoneNumber, messageOTP, expire)
	if err != nil {
		return err
	}

	tempPayload := models.PhoneOTP{
		PhoneNumber: phoneNumber,
		OTP:         otpCode,
	}

	cacheKey := repo.generateAuthCacheKey(refCode)
	err = repo.cache.Set(cacheKey, tempPayload, expire)
	if err != nil {
		return err
	}

	return nil
}

func (repo AuthenticationSMSRepository) VerifyOTP(refCode string, otpCode string) (bool, error) {
	cacheKey := repo.generateAuthCacheKey(refCode)
	tempPayloadRaw, err := repo.cache.Get(cacheKey)
	if err != nil {
		return false, err
	}

	tempPayload := models.PhoneOTP{}

	err = json.Unmarshal([]byte(tempPayloadRaw), &tempPayload)

	if err != nil {
		return false, err
	}

	if tempPayload.OTP == otpCode {
		return true, nil
	}

	return false, nil
}

func (repo AuthenticationSMSRepository) generateAuthCacheKey(refCode string) string {
	return fmt.Sprintf("auth-otp:%s", refCode)
}

func (repo AuthenticationSMSRepository) SendOTPViaLink(fullPhoneNumber string) (models.OTPResponse, error) {
	url := "https://smsapi.deecommerce.co.th:4300/service/v1/otp/request"
	payload := map[string]string{
		"accountId": "08992231310610",
		"secretKey": "U2FsdGVkX19gSK0SR/xX5DAa6B2Mn1wDyEo1es83LNQ=",
		"type":      "OTP",
		"lang":      "th",
		"to":        fullPhoneNumber,
		"sender":    "deeSMS.OTP",
		"isShowRef": "1",
	}

	resp, err := repo.sendPostRequest(url, payload)
	if err != nil {
		return models.OTPResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.OTPResponse{}, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return models.OTPResponse{}, err
	}

	if errVal, ok := result["error"]; ok && errVal != "0" {
		return models.OTPResponse{}, fmt.Errorf("error sending OTP")
	}

	otpResponse := models.OTPResponse{}

	tempResult := result["result"].(map[string]interface{})
	otpResponse.OTPToken = tempResult["token"].(string)
	otpResponse.OTPRefCode = tempResult["ref"].(string)

	return otpResponse, nil
}

func (repo AuthenticationSMSRepository) sendPostRequest(url string, payload map[string]string) (*http.Response, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to server: %w", err)
	}

	return resp, nil
}

func (repo AuthenticationSMSRepository) VerifyOTPViaLink(otpToken, optRefCode, otpPin string) (bool, error) {

	url := "https://smsapi.deecommerce.co.th:4300/service/v1/otp/verify"
	payload := map[string]string{
		"accountId": "08992231310610",
		"secretKey": "U2FsdGVkX19gSK0SR/xX5DAa6B2Mn1wDyEo1es83LNQ=",
		"token":     otpToken,
		"ref":       optRefCode,
		"pin":       otpPin,
	}

	resp, err := repo.sendPostRequest(url, payload)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %w", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return false, fmt.Errorf("error unmarshalling response JSON: %w", err)
	}

	if errVal, ok := result["error"]; ok && errVal == "0" {
		return true, nil
	}

	return false, nil
}
