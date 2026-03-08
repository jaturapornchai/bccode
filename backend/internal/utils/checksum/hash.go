package checksum

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

func Sum(val interface{}) (string, error) {

	data, err := json.Marshal(val)
	if err != nil {
		return "", err
	}

	h := sha1.New()

	hashData := h.Sum(data)

	hashString := hex.EncodeToString(hashData)
	return hashString, nil
}

func CheckSum(hash string, val interface{}) (bool, string, error) {
	newHash, err := Sum(val)
	if err != nil {
		return false, "", err
	}

	return newHash == hash, newHash, nil
}
