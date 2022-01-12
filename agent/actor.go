package agent

import (
	"github.com/icpfans-xyz/agent-go/principal"
)

type CallConfig struct {
	Agent      Agent
	CanisterId principal.Principal
}
