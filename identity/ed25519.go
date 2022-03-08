package identity

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
)

type Ed25519PublicKey struct {
	pub crypto.PublicKey
}

type Ed25519Identity struct {
	PriKey ed25519.PrivateKey
	PubKey Ed25519PublicKey
}

func NewEd25519Identity(pkey []byte) *Ed25519Identity {
	priv := ed25519.NewKeyFromSeed(pkey)
	return &Ed25519Identity{
		PriKey: priv,
		PubKey: Ed25519PublicKey{pub: priv.Public()},
	}
}

func (t *Ed25519Identity) SecretKey() []byte {
	return t.PriKey.Seed()
}

func (t *Ed25519Identity) Sign(m []byte) ([]byte, error) {
	return ed25519.Sign(t.PriKey, m[:]), nil
}

func (t *Ed25519Identity) PublicKey() PublicKey {
	return &t.PubKey
}

func (p *Ed25519PublicKey) ToBytes() []byte {
	return p.pub.([]byte)
}

func (p *Ed25519PublicKey) ToDer() []byte {
	bytes, err := MarshalEd25519PublicKey(p.pub)
	if err != nil {
		panic(err)
	}
	return bytes
}

var errEd25519WrongID = errors.New("incorrect object identifier")
var errEd25519WrongKeyType = errors.New("incorrect key type")

// ed25519OID is the OID for the Ed25519 signature scheme: see
// https://datatracker.ietf.org/doc/draft-ietf-curdle-pkix-04.
var ed25519OID = asn1.ObjectIdentifier{1, 3, 101, 112}

// subjectPublicKeyInfo reflects the ASN.1 object defined in the X.509 standard.
//
// This is defined in crypto/x509 as "publicKeyInfo".
type subjectPublicKeyInfo struct {
	Algorithm pkix.AlgorithmIdentifier
	PublicKey asn1.BitString
}

// MarshalEd25519PublicKey creates a DER-encoded SubjectPublicKeyInfo for an
// ed25519 public key, as defined in
// https://tools.ietf.org/html/draft-ietf-curdle-pkix-04. This is analogous to
// MarshalPKIXPublicKey in crypto/x509, which doesn't currently support Ed25519.
func MarshalEd25519PublicKey(pk crypto.PublicKey) ([]byte, error) {
	pub, ok := pk.(ed25519.PublicKey)
	if !ok {
		return nil, errEd25519WrongKeyType
	}

	spki := subjectPublicKeyInfo{
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm: ed25519OID,
		},
		PublicKey: asn1.BitString{
			BitLength: len(pub) * 8,
			Bytes:     pub,
		},
	}

	return asn1.Marshal(spki)
}
