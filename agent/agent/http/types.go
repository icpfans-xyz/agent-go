package http

import (
	ainterface "github.com/dfinity/agent-go/agent/agent"
)

type Endpoint string

const (
	EndpointQuery     Endpoint = "read"
	EndpointReadState Endpoint = "read_state"
	EndpointCall      Endpoint = "call"
)

type HttpAgentBaseRequest interface {
	Endpoint() string
	RequestBody() interface{}
}

type HttpAgentRequestTransform interface {
	Priority() int64
}

type HttpAgentRequest struct {
	HttpRequest *HttpRequest
	Point       string
	Body        ainterface.Request
}

func (r *HttpAgentRequest) Endpoint() string {
	return r.Point
}

func (r *HttpAgentRequest) RequestBody() interface{} {
	return r.Body
}

type HttpRequest struct {
	Method  string
	Body    []byte
	Headers map[string]string
}
