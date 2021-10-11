package agent

import (
	"errors"

	"github.com/dfinity/agent-go/agent/agent"
	"github.com/dfinity/agent-go/principal"
)

var DomainSeparator = []byte("\x0Aic-request")

type Signature = []byte

/**
 * A Key Pair, containing a secret and public key.
 */
type KeyPair interface {
	SecretKey() []byte
	PublicKey() PublicKey
}

/**
 * A Public Key implementation.
 */
type PublicKey interface {
	// Get the public key bytes encoded with DER.
	ToDer() []byte
}

type SignIdentity struct {
	key       KeyPair
	principal *principal.Principal
}

func (s *SignIdentity) GetPublicKey() PublicKey {
	return s.key.PublicKey()
}

func (s *SignIdentity) Sign(blob []byte) (Signature, error) {
	return nil, errors.New("Sign need implemented")
}

func (s *SignIdentity) GetPrincipal() (*principal.Principal, error) {
	if s.principal == nil {
		principal, err := principal.SelfAuthenticating(s.GetPublicKey().ToDer())
		if err != nil {
			return nil, err
		}
		s.principal = principal
	}
	return s.principal, nil
}

func (s *SignIdentity) TransformRequest(request agent.Request) (interface{}, error) {
	return nil, errors.New("TransformRequest need implemented")
}
