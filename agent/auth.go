package agent

import (
	"github.com/icpfans-xyz/agent-go/identity"
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

type SignIdentity struct {
	key       identity.KeyPair
	principal *principal.Principal
}

func NewSignIdentity(key identity.KeyPair, principal *principal.Principal) *SignIdentity {
	return &SignIdentity{
		key:       key,
		principal: principal,
	}
}

func (s *SignIdentity) GetPublicKey() identity.PublicKey {
	return s.key.PublicKey()
}

func (s *SignIdentity) Sign(blob []byte) (Signature, error) {
	return s.key.Sign(blob)
}

func (s *SignIdentity) GetPrincipal() *principal.Principal {
	if s.principal == nil {
		principal, err := principal.SelfAuthenticating(s.GetPublicKey().ToDer())
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
	pubDer := s.GetPublicKey().ToDer()
	return &TransformRequest{
		Body: SignTransformBody{
			Content:      request,
			SenderPubkey: pubDer,
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
