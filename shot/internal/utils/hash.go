package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
)

func GenerateSignature(token, expire, privateKey string) string {
	h := hmac.New(sha1.New, []byte(privateKey))
	h.Write([]byte(token + expire))
	return hex.EncodeToString(h.Sum(nil))
}