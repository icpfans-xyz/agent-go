package utils

import "hash/crc32"

func Crc32(bytes []byte) uint32 {
	return crc32.ChecksumIEEE(bytes)
}
