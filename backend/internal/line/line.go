package line

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ILineAdapter interface {
	ValidateToken(token string) (*UserInfo, error)
}

type LineAdapter struct {
	clientId string
}

func NewLineAdapter(clientId string) ILineAdapter {
	return &LineAdapter{clientId: clientId}
}

func (l *LineAdapter) ValidateToken(token string) (*UserInfo, error) {
	// Verify the ID token
	claims, err := l.verifyToken(token)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		UserId:        claims.Sub,
		DisplayName:   claims.Name,
		PictureUrl:    claims.Picture,
		StatusMessage: "", // Not provided in ID token claims
	}, nil
}

type verifyResponse struct {
	Scope string `json:"scope"`
	ClientId string `json:"client_id"`
	ExpiresIn int    `json:"expires_in"`
}

type profileResponse struct {
	UserId string `json:"user_id"`
	DisplayName string `json:"display_name"`
	PictureUrl string `json:"picture_url"`
	StatusMessage string `json:"status_message"`
}

type idTokenClaims struct {
	Iss string   `json:"iss"`
	Sub string   `json:"sub"`
	Aud string   `json:"aud"`
	Exp int64    `json:"exp"`
	Iat int64    `json:"iat"`
	Amr []string `json:"amr"`
	Name string   `json:"name"`
	Picture string   `json:"picture"`
}

func (l *LineAdapter) verifyToken(token string) (*idTokenClaims, error) {
	data := url.Values{}
	data.Set("id_token", token)
	data.Set("client_id", l.clientId)

	req, err := http.NewRequest(http.MethodPost, "https://api.line.me/oauth2/v2.1/verify", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(data.Encode()))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("network failure: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid token or network error")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response: " + err.Error())
	}

	var claims idTokenClaims
	err = json.Unmarshal(body, &claims)
	if err != nil {
		return nil, errors.New("invalid JSON response: " + err.Error())
	}

	// Validate claims
	if claims.Iss != "https://access.line.me" {
		return nil, errors.New("invalid issuer")
	}
	if claims.Aud != l.clientId {
		return nil, errors.New("invalid audience")
	}
	if claims.Exp <= time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

func (l *LineAdapter) getProfile(token string) (*profileResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.line.me/v2/profile", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var profile profileResponse
	err = json.Unmarshal(body, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
