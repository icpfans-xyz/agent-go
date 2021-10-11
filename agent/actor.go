package agent

import (
	impl "github.com/dfinity/agent-go/agent/agent"
	"github.com/dfinity/agent-go/principal"
)

type CallConfig struct {
	Agent      impl.Agent
	CanisterId principal.Principal
}
