package microservice

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"
	"testing"
)

const (
	jwtKey = "946796991b2ece76900bfbc65612debc2e54554ef692cef6ec52181abe063d4d"
)

type TestCacherConfig struct{}

func (cfg *TestCacherConfig) Endpoint() string {
	return "127.0.0.1:6379"
}

func (cfg *TestCacherConfig) Password() string {
	return ""
}

func (cfg *TestCacherConfig) DB() int {
	return 0
}

func (cfg *TestCacherConfig) ConnectionSettings() config.ICacherConnectionSettings {
	return config.NewDefaultCacherConnectionSettings()
}

func (cfg *TestCacherConfig) UserName() string {
	return ""
}

func (cfg *TestCacherConfig) TLS() bool {
	return false
}

func TestGenerateToken(t *testing.T) {

	cacher := NewCacher(&TestCacherConfig{})

	jwtService := NewJwtService(cacher, jwtKey, 60*24*10)

	token, err := jwtService.GenerateToken(models.UserInfo{Username: "u001", Name: "My Name"})

	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(token) == 0 {
		t.Error("token is empty")
		return
	}

	t.Log(token)

}

func TestParseToken(t *testing.T) {

	cacher := NewCacher(&TestCacherConfig{})

	jwtService := NewJwtService(cacher, jwtKey, 60*24*10)

	tokenString, err := jwtService.GenerateToken(models.UserInfo{Username: "u001", Name: "My Name"})

	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(tokenString) == 0 {
		t.Error("token is empty")
		return
	}

	token, err := jwtService.ParseToken(tokenString)

	if err != nil {
		t.Error(err.Error())
		return
	}

	claims := token.Claims.(*CustomClaims)

	if claims.UserInfo.Username != "u001" {
		t.Error("username in token invalid")
		return
	}

	t.Log(claims.UserInfo.Name)
}

func TestParseTokenReal(t *testing.T) {

	cacher := NewCacher(&TestCacherConfig{})
	jwtService := NewJwtService(cacher, jwtKey, 60*24*10)

	tokenGen, err := jwtService.GenerateToken(models.UserInfo{Username: "u001", Name: "My Name"})

	if err != nil {
		t.Error(err.Error())
		return
	}

	token, err := jwtService.ParseToken(tokenGen)

	if err != nil {
		t.Error(err.Error())
		return
	}

	claims := token.Claims.(*CustomClaims)

	t.Log(claims.UserInfo.Name)
}

/*
func getKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := io.ReadFile("./../../private.key")

	if err != nil {
		return nil, nil, err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)

	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := io.ReadFile("./../../public.key")

	if err != nil {
		return nil, nil, err
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)

	if err != nil {
		return nil, nil, err
	}

	return signKey, verifyKey, nil
}

func TestGenerateToken(t *testing.T) {

	signKey, verifyKey, err := getKey()

	if err != nil {
		t.Error(err.Error())
		return
	}

	jwtService := NewJwtService(signKey, verifyKey, 60*24*10)

	token, err := jwtService.GenerateToken(models.UserInfo{Username: "u001", Name: "My Name"}, time.Duration(60*24*10)*time.Minute)

	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(token) == 0 {
		t.Error("token is empty")
		return
	}

	t.Log(token)

}

func TestParseToken(t *testing.T) {
	signKey, verifyKey, err := getKey()

	if err != nil {
		t.Error(err.Error())
		return
	}

	jwtService := NewJwtService(signKey, verifyKey, 60*24*10)

	tokenString, err := jwtService.GenerateToken(models.UserInfo{Username: "u001", Name: "My Name"}, time.Duration(60*24*10)*time.Minute)

	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(tokenString) == 0 {
		t.Error("token is empty")
		return
	}

	token, err := jwtService.ParseToken(tokenString)

	if err != nil {
		t.Error(err.Error())
		return
	}

	claims := token.Claims.(*CustomClaims)

	if claims.UserInfo.Username != "u001" {
		t.Error("username in token invalid")
		return
	}

	t.Log(claims.UserInfo.Name)
}

func TestParseTokenReal(t *testing.T) {
	signKey, verifyKey, err := getKey()

	if err != nil {
		t.Error(err.Error())
		return
	}

	jwtService := NewJwtService(signKey, verifyKey, 60*24*10)

	tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDYwMjIwMzAsInVzZXJuYW1lIjoiZGV2MDEiLCJuYW1lIjoiZGV2IGRldiJ9.FW613xlE7PJm_PADLFE9r65toXuuvl7q6cV7Pi_Lj7dEAfnQ4-fSJDlicS79k-s6rqtV835xV9nGmM1UEgl-VPAO_4syy-LpXjzJL2eRNFmMg3Cy1jp_RYY-mlHvshfVyCRqh6ZDbe_9lUKZ1avYjqjF9fjFWPB4IRF53kiYz3hIVbu3hAsbEYoQhH32sSdYtgF5aaa78XLH3c9BG3KHKWUpE4LU2bJp5NcDL2JT9NBRXZ8slMjHW0JGGh5oSnt2yYFJDSyWY1K8Z6tYnG4pY9of_qSfRzAM_KHBLoVTReYOLXIHfRpUe8VPHiwLvjG7Tn1VocLkCqpaZYxYv6iJtlHMkmx5RXs8QYCDWmrgUeqEhrGiIxU-VDKI5wa6YG1f1QcHVjEC461lpjtaytSfxPn_Jq_XfqEUBBERnQFESuOUCPmQUHgHLQYPS-Xoqxu5zcKweXfoSUvNfX0NqSTviaif4lv7J44iWijDcL3JqmDCdRx7xK02BSdI7TMTqw1u0-h5HegrPeKp2I7k3BGiEB76TkUW0O4nR1CigEKz0onD0_yeQkQv1zi_esi7zy1UlWmIlNfKzbZj6DmLNbgjbKZ0A_FaOi80zUdaVPEgPuzRj4Y7LuqpLszq7uHRQ9_MNdnT6jttY8GJy5Wli0j5gKqdeQdFWcqxogDs8UEQbCQ"

	token, err := jwtService.ParseToken(tokenString)

	if err != nil {
		t.Error(err.Error())
		return
	}

	claims := token.Claims.(*CustomClaims)

	t.Log(claims.UserInfo.Name)
}
*/
