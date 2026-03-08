package utils

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func RandNumber(n int) string {
	var result string
	for i := 0; i < n; i++ {
		result += fmt.Sprint(rand.Intn(9))
	}
	return result
}

func NewGUID() string {
	newid := ksuid.New()
	return newid.String()
}

func NewID() string {
	newid := xid.New()
	return newid.String()
}

func NewUUID() string {
	return uuid.NewString()
}

func RandNumberX(n int) string {
	var result string
	for i := 0; i < n; i++ {
		result += fmt.Sprint(rand.Intn(9))
	}
	return result
}
