package utils

import "crypto/sha256"

func Sha224(data []byte) [28]byte {
	return sha256.Sum224(data)
}
