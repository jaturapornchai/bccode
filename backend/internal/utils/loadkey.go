package utils

import (
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt"
)

func LoadFile(filePath string) ([]byte, error) {

	fileBytes, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func LoadKey(signPath string, verifyPath string) (*rsa.PrivateKey, *rsa.PublicKey, error) {

	signBytes, err := LoadFile(signPath)

	if err != nil {
		return nil, nil, err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)

	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := LoadFile(verifyPath)

	if err != nil {
		return nil, nil, err
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)

	if err != nil {
		return nil, nil, err
	}

	return signKey, verifyKey, nil
}
