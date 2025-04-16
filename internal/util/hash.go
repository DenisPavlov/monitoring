package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func GetHexSHA256(key string, value []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(value)
	return fmt.Sprintf("%x", h.Sum(nil))
}
