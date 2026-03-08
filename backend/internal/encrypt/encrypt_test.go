package encrypt_test

import (
	"fmt"
	"smlcloudplatform/internal/encrypt"
	"testing"
)

func TestEncrypt(t *testing.T) {

	encrypt := &encrypt.Encrypt{}

	input := "2OJMfbQIUGfZMAPmmst9QDlyMfO"
	hash := encrypt.GenerateSHA256Hash(input)
	fmt.Printf("Input: %s\nSHA256 Hash: %s\n", input, hash)
}
