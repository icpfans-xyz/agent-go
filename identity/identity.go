package identity

type PublicKey interface {
	// Get the public key bytes
	ToBytes() []byte

	// Get the public key bytes encoded with DER.
	ToDer() []byte
}

/**
 * A Key Pair, containing a secret and public key.
 */
type KeyPair interface {
	SecretKey() []byte
	Sign([]byte) ([]byte, error)
	PublicKey() PublicKey
}
