package polling

import (
	"github.com/icpfans-xyz/agent-go/agent"
	"github.com/icpfans-xyz/agent-go/agent/http"
	"github.com/icpfans-xyz/agent-go/principal"
)

type PollStrategy = func(*principal.Principal, *agent.RequestId, http.RequestStatusResponseStatus)

type PollStrategyFactory = func() PollStrategy

/**
 * Polls the IC to check the status of the given request then
 * returns the response bytes once the request has been processed.
 * @param agent The agent to use to poll read_state.
 * @param canisterId The effective canister ID.
 * @param requestId The Request ID to poll status for.
 * @param strategy A polling strategy.
 */
func PollForResponse(agentimpl agent.Agent, canisterId *principal.Principal, requestId agent.RequestId, strategy PollStrategy) ([]byte, error) {
	options := &agent.ReadStateOptions{
		Paths: [][][]byte{[][]byte{[]byte("request_status"), requestId[:]}},
	}
	_, err := agentimpl.ReadState(canisterId, options)
	if err != nil {
		return nil, err
	}
	// cert :=
	return nil, nil
}
