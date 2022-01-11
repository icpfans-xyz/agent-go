package utils

import (
	"errors"
	"strings"
)

// func Base32Encode(bytes []byte) string {
// 	return base32.StdEncoding.EncodeToString(bytes)
// }

// func Base32Decode(input string) ([]byte, error) {
// 	return base32.StdEncoding.DecodeString(input)
// }

var alphabet = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "2", "3", "4", "5", "6", "7"}

// Build a lookup table for decoding.
var lookupTable map[string]int = map[string]int{}

func init() {
	i := 0
	for i < len(alphabet) {
		lookupTable[alphabet[i]] = i
		i++
	}
	// Add aliases for rfc4648.
	lookupTable["0"] = lookupTable["o"]
	lookupTable["1"] = lookupTable["i"]
}

/**
 * @param input The input array to encode.
 * @returns A Base32 string encoding the input.
 */
func Base32Encode(input []byte) string {
	// How many bits will we skip from the first byte.
	skip := 0
	// 5 high bits, carry from one byte to the next.
	bits := 0

	// The output string in base32.
	output := ""

	encodeByte := func(ebyte int) int {
		if skip < 0 {
			// we have a carry from the previous byte
			bits |= ebyte >> -skip
		} else {
			// no carry
			bits = (ebyte << skip) & 248
		}

		if skip > 3 {
			// Not enough data to produce a character, get us another one
			skip -= 8
			return 1
		}

		if skip < 4 {
			// produce a character
			output += alphabet[bits>>3]
			skip += 5
		}
		return 0
	}

	i := 0
	for i < len(input) {
		i += encodeByte(int(input[i]))
	}

	if skip < 0 {
		return output + alphabet[bits>>3]
	}
	return output
}

/**
 * @param input The base32 encoded string to decode.
 */
func Base32Decode(input string) ([]byte, error) {
	// how many bits we have from the previous character.
	skip := 0
	// current byte we're producing.
	ebyte := 0

	output := make([]byte, ((len(input) * 4) / 3))
	o := 0

	decodeChar := func(char string) error {
		// Consume a character from the stream, store
		// the output in this.output. As before, better
		// to use update().
		val, ok := lookupTable[strings.ToLower(char)]
		if !ok {
			return errors.New("Invalid character:" + char)
		}

		// move to the high bits
		val <<= 3
		ebyte |= int(uint(val) >> skip)
		skip += 5

		if skip >= 8 {
			// We have enough bytes to produce an output
			output[o] = byte(ebyte)
			o++
			skip -= 8

			if skip > 0 {
				ebyte = (val << (5 - skip)) & 255
			} else {
				ebyte = 0
			}
		}
		return nil
	}

	for _, ch := range input {
		err := decodeChar(string(ch))
		if err != nil {
			return nil, err
		}
	}

	return output[0:o], nil
}
