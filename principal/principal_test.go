package principal_test

import (
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/icpfans-xyz/agent-go/principal"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {

	t.Run("encode", func(t *testing.T) {
		canisterId, err := principal.FromString("rrkah-fqaaa-aaaaa-aaaaq-cai")
		assert.Nil(t, err)
		bytes, err := cbor.Marshal(canisterId.Bytes)
		expected := []byte{0x4a, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x1}
		assert.Nil(t, err)
		assert.Equal(t, expected, bytes)
	})

}

func TestFromHex(t *testing.T) {

	t.Run("encode", func(t *testing.T) {
		canisterId, err := principal.FromHex("abcd01")
		assert.Nil(t, err)
		// bytes, err := cbor.Marshal(canisterId.Bytes)
		expected := "em77e-bvlzu-aq"
		assert.Nil(t, err)
		assert.Equal(t, expected, canisterId.ToString())
	})

}
