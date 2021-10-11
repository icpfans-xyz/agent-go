package http

import (
	"math/big"
	"time"

	cbor "github.com/fxamacker/cbor/v2"
)

var NANOSECONDS_PER_MILLISECONDS = big.NewInt(1000000)

var REPLICA_PERMITTED_DRIFT_MILLISECONDS = big.NewInt(60 * 1000)

type Expiry struct {
	value big.Int
}

func NewExpiry(deltaInMSec int64) *Expiry {
	value := big.NewInt(time.Now().UnixNano())
	value = big.NewInt(0).Add(value, big.NewInt(deltaInMSec))
	value = big.NewInt(0).Sub(value, REPLICA_PERMITTED_DRIFT_MILLISECONDS)
	return &Expiry{value: *big.NewInt(0).Mul(value, NANOSECONDS_PER_MILLISECONDS)}
}

func (e *Expiry) ToCBOR() ([]byte, error) {
	return cbor.Marshal(e)
}

func (e *Expiry) ToHash() []byte {
	return e.value.Bytes()
}
