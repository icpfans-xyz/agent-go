package http_test

import (
	"encoding/hex"
	"testing"

	"github.com/icpfans-xyz/agent-go/agent"
	ainterface "github.com/icpfans-xyz/agent-go/agent/agent"
	"github.com/icpfans-xyz/agent-go/agent/agent/http"
	"github.com/icpfans-xyz/agent-go/principal"
	"github.com/mix-labs/IC-Go/utils/idl"
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

func TestAgentStatus(t *testing.T) {
	agent := setupAgent(t)

	resp, err := agent.Status()
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestAgent(t *testing.T) {
	agent := setupAgent(t)

	canisterId, err := principal.FromString("rrkah-fqaaa-aaaaa-aaaaq-cai")
	assert.Nil(t, err)
	// principal := principal.Anonymous()

	// t.Run("call", func(t *testing.T) {
	// 	options := &ainterface.CallOptions{
	// 		MethodName: "greet",
	// 		Arg:        []byte("DIDL\x00\xFD*"),
	// 	}
	// 	resp, err := agent.Call(canisterId, options)
	// 	assert.Nil(t, err)
	// 	assert.NotNil(t, resp)
	// })

	t.Run("query", func(t *testing.T) {
		options := &ainterface.QueryFields{
			MethodName: "greet",
			Arg:        []byte("hello"),
		}
		resp, err := agent.Query(canisterId, options)
		assert.Nil(t, err)
		assert.NotNil(t, resp)
	})

}

func TestAgent2(t *testing.T) {

	pbBytes, _ := hex.DecodeString("833fe62409237b9d62ec77587520911e9a759cec1d19755b7da901b96dca3d42")
	options := &http.HttpAgentOptions{
		Source:   nil,
		Host:     "https://ic0.app",
		Identity: agent.NewSignIdentity(agent.NewIdentityKey(pbBytes), nil),
	}
	agent, err := http.NewHttpAgent(*options)
	assert.Nil(t, err)

	canisterID, _ := principal.FromString("bzsui-sqaaa-aaaah-qce2a-cai")
	methodName := "supply"
	arg, err := idl.Encode([]idl.Type{new(idl.Text)}, []interface{}{"Motoko"})
	assert.Nil(t, err)
	opts := &ainterface.QueryFields{
		MethodName: methodName,
		Arg:        arg,
	}
	resp, err := agent.Query(canisterID, opts)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

}
