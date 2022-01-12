package agent

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"

	"github.com/icpfans-xyz/agent-go/principal"
)

var DomainSeparator = []byte("\x0Aic-request")

type SignTransformBody struct {
	Content      Request `cbor:"content,omitempty"`
	SenderPubkey []byte  `cbor:"sender_pubkey,omitempty"`
	SenderSig    []byte  `cbor:"sender_sig,omitempty"`
}

type AnonymousTransformBody struct {
	Content Request `cbor:"content,omitempty"`
}

type TransformRequest struct {
	Request
	Body interface{}
}

type Signature = []byte

/**
 * A Key Pair, containing a secret and public key.
 */
type KeyPair interface {
	SecretKey() []byte
	Sign([]byte) ([]byte, error)
	PublicKey() crypto.PublicKey
}

type IdentityKey struct {
	PriKey ed25519.PrivateKey
}

func NewIdentityKey(pkey []byte) *IdentityKey {
	return &IdentityKey{
		PriKey: ed25519.NewKeyFromSeed(pkey),
	}
}

func (t *IdentityKey) SecretKey() []byte {
	return t.PriKey.Seed()
}

func (t *IdentityKey) Sign(m []byte) ([]byte, error) {
	return ed25519.Sign(t.PriKey, m[:]), nil
}

func (t *IdentityKey) PublicKey() crypto.PublicKey {
	return t.PriKey.Public()
}

type SignIdentity struct {
	key       KeyPair
	principal *principal.Principal
}

func NewSignIdentity(key KeyPair, principal *principal.Principal) *SignIdentity {
	return &SignIdentity{
		key:       key,
		principal: principal,
	}
}

func (s *SignIdentity) GetPublicKey() crypto.PublicKey {
	return s.key.PublicKey()
}

func (s *SignIdentity) Sign(blob []byte) (Signature, error) {
	return s.key.Sign(blob)
}

func (s *SignIdentity) GetPrincipal() *principal.Principal {
	if s.principal == nil {
		pubBytes, _ := MarshalEd25519PublicKey(s.GetPublicKey())
		principal, err := principal.SelfAuthenticating(pubBytes)
		if err != nil {
			panic(err)
		}
		s.principal = principal
	}
	return s.principal
}

func (s *SignIdentity) TransformRequest(request Request) (*TransformRequest, error) {
	requestId := RequestIdOf(request)
	sign, err := s.Sign(append(DomainSeparator, requestId[:]...))
	if err != nil {
		return nil, err
	}
	pubBytes, _ := MarshalEd25519PublicKey(s.GetPublicKey())
	return &TransformRequest{
		Body: SignTransformBody{
			Content:      request,
			SenderPubkey: pubBytes,
			SenderSig:    sign,
		},
	}, nil
}

type AnonymousIdentity struct {
	principal *principal.Principal
}

func NewAnonymousIdentity() *AnonymousIdentity {
	return &AnonymousIdentity{
		principal: principal.Anonymous(),
	}
}

func (a *AnonymousIdentity) GetPrincipal() *principal.Principal {
	return a.principal
}

func (s *AnonymousIdentity) TransformRequest(request Request) (*TransformRequest, error) {
	return &TransformRequest{
		Body: AnonymousTransformBody{
			Content: request,
		},
	}, nil
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
