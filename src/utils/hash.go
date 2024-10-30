package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GetPasswordHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

func GenerateSessionToken(length int) string {
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
