package utils

import (
	"math/big"
)

var (
	x00 = big.NewInt(0x00)
	x7F = big.NewInt(0x7F)
	x80 = big.NewInt(0x80)
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

func LebEncode(v *big.Int) []byte {
	if v.Cmp(big.NewInt(0)) < 0 {
		panic("Cannot leb encode negative values.")
	}
	var bs []byte
	for {
		i := new(big.Int).And(v, x7F)
		v = v.Div(v, x80)
		if v.Cmp(x00) == 0 {
			b := i.Bytes()
			if len(b) == 0 {
				return []byte{0}
			}
			return append(bs, b...)
		} else {
			b := new(big.Int).Or(i, x80)
			bs = append(bs, b.Bytes()...)
		}
	}
}

func LebDecode(bytes []byte) *big.Int {
	var (
		weight = big.NewInt(1)
		value  = new(big.Int)
	)
	for _, b := range bytes {
		value = value.Add(
			value,
			new(big.Int).Mul(big.NewInt(int64(b&0x7f)), weight),
		)
		weight = weight.Mul(weight, x80)
		if b < 0x80 {
			break
		}
	}
	return value
}
