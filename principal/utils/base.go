package utils

import (
	"encoding/base32"
)

func Base32Encode(bytes []byte) string {
	return base32.StdEncoding.EncodeToString(bytes)
}

func Base32Decode(input string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(input)
}
