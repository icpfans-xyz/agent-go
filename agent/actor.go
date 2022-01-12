package agent

import (
	impl "github.com/icpfans-xyz/agent-go/agent/agent"
	"github.com/icpfans-xyz/agent-go/principal"
)

type CallConfig struct {
	Agent      impl.Agent
	CanisterId principal.Principal
}
