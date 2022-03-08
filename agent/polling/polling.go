package polling

import (
	"errors"
	"fmt"

	"github.com/icpfans-xyz/agent-go/agent"
	"github.com/icpfans-xyz/agent-go/agent/http"
	"github.com/icpfans-xyz/agent-go/principal"
	"github.com/icpfans-xyz/agent-go/principal/utils"
)

type PollStrategy = func(*principal.Principal, agent.RequestId, http.RequestStatusResponseStatus) error

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
	paths := [][]byte{[]byte("request_status"), requestId[:]}
	options := &agent.ReadStateOptions{
		Paths: [][][]byte{paths},
	}
	state, err := agentimpl.ReadState(canisterId, options)
	if err != nil {
		return nil, err
	}
	cert, err := agent.NewCertificate(*state, agentimpl)
	if err != nil {
		return nil, err
	}
	verified := cert.Verify()
	if !verified {
		return nil, errors.New("Fail to verify certificate")
	}
	statusBytes, err := cert.Lookup(paths)
	if err != nil {
		return nil, err
	}
	status := http.RequestStatusResponseStatus(statusBytes)

	switch status {
	case http.StatusReplied:
		return cert.Lookup(append(paths, []byte("reply")))
	case http.StatusReceived:
		fallthrough
	case http.StatusUnknown:
		fallthrough
	case http.StatusProcessing:
		strategy(canisterId, requestId, status)
		return PollForResponse(agentimpl, canisterId, requestId, strategy)
	case http.StatusRejected:
		code, err := cert.Lookup(append(paths, []byte("reject_code")))
		if err != nil {
			return nil, err
		}
		msg, err := cert.Lookup(append(paths, []byte("reject_message")))
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Call was rejected: RequestId:%s, Reject Code:%d, Reject Msg:%s", utils.Hex(requestId[:]), utils.Uint32(code), string(msg))
	case http.StatusDone:
		// This is _technically_ not an error, but we still didn't see the `Replied` status so
		// we don't know the result and cannot decode it.
		return nil, fmt.Errorf("Call was marked as done but we never saw the reply: RequestId:%s", utils.Hex(requestId[:]))
	}

	return nil, errors.New("unreachable")
}
