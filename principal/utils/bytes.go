package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// Hex encodes []byte to Hex.
func Hex(data []byte) string {
	return hex.EncodeToString(data)
}

// FromHex decodes string from Hex.
func FromHex(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// Uint32 encodes []byte.
func Uint32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// FromUint32 decodes uint32.
func FromUint32(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// Sha256 returns the SHA-256 digest of the data.
func Sha256(args ...[]byte) []byte {
	hasher := sha256.New()
	for _, bytes := range args {
		hasher.Write(bytes)
	}
	return hasher.Sum(nil)
}
