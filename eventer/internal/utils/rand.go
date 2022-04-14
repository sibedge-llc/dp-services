package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomId() string {
	idbytes := GenerateRandomIdBytesN(16)
	return hex.EncodeToString(idbytes)
}

func GenerateRandomIdN(n int) string {
	idbytes := GenerateRandomIdBytesN((n >> 1) + 1)
	return hex.EncodeToString(idbytes)[:n]
}

func GenerateRandomIdBytes() []byte {
	return GenerateRandomIdBytesN(16)
}

func GenerateRandomIdBytesN(n int) []byte {
	idbytes := make([]byte, n)
	_, err := rand.Read(idbytes)
	if err != nil {
		panic(err)
	}
	return idbytes
}
