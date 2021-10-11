package http

import (
	"github.com/dfinity/agent-go/agent/agent"
	"github.com/dfinity/agent-go/principal"
)

type Endpoint string

const (
	EndpointQuery     Endpoint = "read"
	EndpointReadState Endpoint = "read_state"
	EndpointCall      Endpoint = "call"
)

type HttpAgentBaseRequest interface {
	Endpoint() string
	RequestBody() agent.RequestBody
}

type HttpAgentRequestTransform interface {
	Priority() int64
}

// CallRequest call request
type CallRequest struct {
	requestType string
	canisterId  *principal.Principal
	method      string
	arg         []byte
	sender      *principal.Principal
	expiry      *Expiry
}

func (c *CallRequest) Type() string {
	return c.requestType
}

func (c *CallRequest) Method() string {
	return c.method
}

func (c *CallRequest) Sender() *principal.Principal {
	return c.sender
}

func (c *CallRequest) IngressExpiry() *Expiry {
	return c.expiry
}

type QueryRequest struct {
	requestType string
	canisterId  *principal.Principal
	method      string
	arg         []byte
	sender      *principal.Principal
	expiry      *Expiry
}

func (c *QueryRequest) Type() string {
	return c.requestType
}

func (c *QueryRequest) Method() string {
	return c.method
}

func (c *QueryRequest) Sender() *principal.Principal {
	return c.sender
}

func (c *QueryRequest) IngressExpiry() *Expiry {
	return c.expiry
}

type ReadStateRequest struct {
	requestType string
	paths       [][]byte
	sender      *principal.Principal
	expiry      *Expiry
}

func (c *ReadStateRequest) Type() string {
	return c.requestType
}

func (c *ReadStateRequest) Method() string {
	return "read_state"
}

func (c *ReadStateRequest) Sender() *principal.Principal {
	return c.sender
}

func (c *ReadStateRequest) IngressExpiry() *Expiry {
	return c.expiry
}

type HttpAgentRequest struct {
	request  *HttpRequest
	endpoint string
	body     agent.RequestBody
}

func (r *HttpAgentRequest) Endpoint() string {
	return r.endpoint
}

func (r *HttpAgentRequest) RequestBody() agent.RequestBody {
	return r.body
}

type HttpRequest struct {
	Method  string
	Body    []byte
	Headers map[string]string
}
