package encrypt

import (
	"crypto/sha256"
	"encoding/hex"
)

type IEncrypt interface {
}

type Encrypt struct{}

func (e *Encrypt) GenerateSHA256Hash(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

func NewEncrypt() *Encrypt {
	return &Encrypt{}
}
