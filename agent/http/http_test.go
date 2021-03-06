package http_test

import (
	"encoding/hex"
	"testing"

	"github.com/icpfans-xyz/agent-go/agent"
	"github.com/icpfans-xyz/agent-go/identity"

	"github.com/icpfans-xyz/agent-go/agent/http"
	"github.com/icpfans-xyz/agent-go/candid/idl"
	"github.com/icpfans-xyz/agent-go/principal"
	"github.com/stretchr/testify/assert"
)

func setupAgent(t *testing.T) *http.HttpAgent {
	options := &http.HttpAgentOptions{
		Source: nil,
		Host:   "http://localhost:8000",
	}
	agent, err := http.NewHttpAgent(*options)
	assert.Nil(t, err)
	return agent
}

func setupRemoteAgent(t *testing.T) *http.HttpAgent {
	pbBytes, _ := hex.DecodeString("833fe62409237b9d62ec77587520911e9a759cec1d19755b7da901b96dca3d42")
	options := &http.HttpAgentOptions{
		Source:   nil,
		Host:     "https://ic0.app",
		Identity: agent.NewSignIdentity(identity.NewEd25519Identity(pbBytes), nil),
	}
	httpAgent, err := http.NewHttpAgent(*options)
	assert.Nil(t, err)
	return httpAgent
}

func TestAgentStatus(t *testing.T) {
	agent := setupRemoteAgent(t)

	resp, err := agent.Status()
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestAgentCall(t *testing.T) {
	httpAgent := setupRemoteAgent(t)

	canisterId, err := principal.FromString("rrkah-fqaaa-aaaaa-aaaaq-cai")
	assert.Nil(t, err)
	// principal := principal.Anonymous()

	t.Run("call", func(t *testing.T) {
		options := &agent.CallOptions{
			MethodName: "greet",
			Arg:        []byte("DIDL\x00\xFD*"),
		}
		resp, err := httpAgent.Call(canisterId, options)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})
}

func TestAgentQuery(t *testing.T) {

	pbBytes, _ := hex.DecodeString("833fe62409237b9d62ec77587520911e9a759cec1d19755b7da901b96dca3d42")
	options := &http.HttpAgentOptions{
		Source:   nil,
		Host:     "https://ic0.app",
		Identity: agent.NewSignIdentity(identity.NewEd25519Identity(pbBytes), nil),
	}
	httpAgent, err := http.NewHttpAgent(*options)
	assert.Nil(t, err)

	canisterID, _ := principal.FromString("bzsui-sqaaa-aaaah-qce2a-cai")
	methodName := "supply"
	arg, err := idl.Encode([]idl.Type{new(idl.Text)}, []interface{}{"Motoko"})
	assert.Nil(t, err)
	opts := &agent.QueryFields{
		MethodName: methodName,
		Arg:        arg,
	}
	resp, err := httpAgent.Query(canisterID, opts)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

}
