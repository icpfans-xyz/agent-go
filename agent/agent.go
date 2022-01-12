package agent

import (
	"github.com/icpfans-xyz/agent-go/principal"
)

type GlobalInternetComputer interface {
	Agent() Agent

	Authentication() bool

	// Canister() Actor
}

/**
 * A General Identity object. This does not have to be a private key (for example,
 * the Anonymous identity), but it must be able to transform request.
 */
type Identity interface {
	/**
	 * Get the principal represented by this identity. Normally should be a
	 * `Principal.selfAuthenticating()`.
	 */
	GetPrincipal() *principal.Principal

	/**
	 * Transform a request into a signed version of the request. This is done last
	 * after the transforms on the body of a request. The returned object can be
	 * anything, but must be serializable to CBOR.
	 */
	TransformRequest(Request) (*TransformRequest, error)
}

var (
	window GlobalInternetComputer
	global GlobalInternetComputer
	self   GlobalInternetComputer
)

func GetDefaultAgent() Agent {
	if self != nil {
		return self.Agent()
	} else if global != nil {
		return global.Agent()
	} else if window != nil {
		return window.Agent()
	}
	panic("No Agent could be found.');")
}
