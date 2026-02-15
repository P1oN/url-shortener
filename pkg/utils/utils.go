package utils

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length int) string {
	if length <= 0 {
		return ""
	}

	b := make([]byte, length)
	random := make([]byte, length)
	if _, err := cryptoRand.Read(random); err == nil {
		for i := range b {
			b[i] = charset[int(random[i])%len(charset)]
		}
		return string(b)
	}

	// Fallback for rare entropy source failures.
	var seed [8]byte
	if _, err := cryptoRand.Read(seed[:]); err != nil {
		binary.LittleEndian.PutUint64(seed[:], 1)
	}
	r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
