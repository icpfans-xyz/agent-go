package identity

import "github.com/icpfans-xyz/agent-go/principal"

type AnonymousTransformBody struct {
	Content Request `cbor:"content,omitempty"`
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
