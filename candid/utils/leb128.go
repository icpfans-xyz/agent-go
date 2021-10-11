package utils

import (
	"math/big"
)

func eob() {
	panic("unexpected end of buffer")
}

func SafeRead(pipe *PipeArrayBuffer, num int64) []byte {
	if pipe.ByteLength() < int(num) {
		eob()
	}
	return pipe.Read(num)
}

func SafeReadUint8(pipe *PipeArrayBuffer) uint8 {
	if pipe.ByteLength() < 1 {
		eob()
	}
	byte := pipe.readUint8()
	return byte
}

func LebEncode(value *big.Int) []byte {
	if value.Cmp(big.NewInt(0)) < 0 {
		panic("Cannot leb encode negative values.")
	}
	// byteLength := 0
	// if value.Cmp(big.NewInt(0)) > 0 {
	// 	// byteLength = Math.ceil(Math.log2(Number(value)))) + 1;
	// 	length := math.Log2(float64(value.Uint64()))
	// 	byteLength = int(math.Ceil(length)) + 1
	// }
	// pipe := NewPipeArrayBuffer(nil, 0)
	return value.Bytes()
}

func LebDecode(bytes []byte) *big.Int {
	return big.NewInt(0).SetBytes(bytes)
}
