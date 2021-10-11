package utils

import "crypto/sha256"

func Sha224(data []byte) []byte {
	return sha256.New224().Sum(data)
}
