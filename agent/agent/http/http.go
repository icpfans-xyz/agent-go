package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dfinity/agent-go/agent"
	ainterface "github.com/dfinity/agent-go/agent/agent"
	"github.com/dfinity/agent-go/principal"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/xerrors"
)

type RequestStatusResponseStatus string

const (
	StatusReceived   = "received"
	StatusProcessing = "processing"
	StatusReplied    = "replied"
	StatusRejected   = "rejected"
	StatusUnknown    = "unknown"
	StatusDone       = "done"
)

// Default delta for ingress expiry is 5 minutes.
const DEFAULT_INGRESS_EXPIRY_DELTA_IN_MSECS = 5 * 60 * 1000

// Root public key for the IC, encoded as hex
const IC_ROOT_KEY = "308182301d060d2b0601040182dc7c0503010201060c2b0601040182dc7c05030201036100814" +
	"c0e6ec71fab583b08bd81373c255c3c371b2e84863c98a4f1e08b74235d14fb5d9c0cd546d968" +
	"5f913a0c0b2cc5341583bf4b4392e467db96d65b9bb4cb717112f8472e0d5a4d14505ffd7484" +
	"b01291091c5f87b98883463f98091a0baaae"

type Credentials struct {
	Name     string
	Password string
}

// HttpAgent options that can be used at construction.
type HttpAgentOptions struct {
	// Another HttpAgent to inherit configuration (pipeline and fetch) of. This
	// is only used at construction.
	Source *HttpAgent

	// The host to use for the client. By default, uses the same host as
	// the current page.
	Host string

	// The principal used to send messages. This cannot be empty at the request
	// time (will throw).
	Identity agent.Identity

	Credentials *Credentials
}

type HttpAgent struct {
	rootKey []byte

	pipeline []HttpAgentRequestTransform

	identity agent.Identity

	host string

	credentials string

	rootKeyFetched bool
}

func NewHttpAgent(options HttpAgentOptions) (*HttpAgent, error) {
	agent := &HttpAgent{
		pipeline: []HttpAgentRequestTransform{},
	}
	if options.Source != nil {
		agent.host = options.Source.host
		agent.identity = options.Source.identity
		agent.credentials = options.Source.credentials
		agent.pipeline = options.Source.pipeline
	}
	if len(options.Host) > 0 {
		agent.host = options.Host
	}
	if options.Identity != nil {
		agent.identity = options.Identity
	}
	if options.Credentials != nil {
		agent.credentials = options.Credentials.Name + ":" + options.Credentials.Password
	}
	return agent, nil
}

func (a *HttpAgent) AddTransform(transform HttpAgentRequestTransform) {
	if len(a.pipeline) > int(transform.Priority()) {
		a.pipeline[transform.Priority()] = transform
	} else {
		a.pipeline = append(a.pipeline, transform)
	}
}

func (a *HttpAgent) RootKey() []byte {
	return a.rootKey
}

func (a *HttpAgent) GetPrincipal() *principal.Principal {
	return a.identity.GetPrincipal()
}

func (a *HttpAgent) ReadState(canisterId *principal.Principal, options *ainterface.ReadStateOptions) (*ainterface.ReadStateResponse, error) {
	sender := a.identity.GetPrincipal()
	state := &ReadStateRequest{
		requestType: "read_state",
		paths:       options.Paths,
		sender:      sender,
		expiry:      NewExpiry(DEFAULT_INGRESS_EXPIRY_DELTA_IN_MSECS),
	}

	request := &HttpAgentRequest{
		request: &HttpRequest{
			Body:    nil,
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/cbor"},
		},
		endpoint: string(EndpointReadState),
		body:     state,
	}
	if len(a.credentials) > 0 {
		request.request.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))
	}
	transformRequest := a.identity.TransformRequest(request)
	body, err := cbor.Marshal(transformRequest.RequestBody())
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v2/canister/%s/read_state", canisterId.ToString())
	resp, err := a.fetch(path, request.request, body)
	if err != nil {
		return nil, err
	}
	var response ainterface.ReadStateResponse
	err = cbor.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (a *HttpAgent) fetch(path string, request *HttpRequest, body []byte) ([]byte, error) {
	client := &http.Client{}
	url := a.host + path
	req, err := http.NewRequest(request.Method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (a *HttpAgent) Call(canisterId *principal.Principal, options *ainterface.CallOptions) (*ainterface.SubmitResponse, error) {
	ecid := canisterId
	if options.EffectiveCanisterId != nil {
		ecid = options.EffectiveCanisterId
	}
	sender := a.identity.GetPrincipal()
	submit := &CallRequest{
		requestType: "call",
		canisterId:  canisterId,
		method:      options.MethodName,
		arg:         options.Arg,
		sender:      sender,
		expiry:      NewExpiry(DEFAULT_INGRESS_EXPIRY_DELTA_IN_MSECS),
	}

	request := &HttpAgentRequest{
		request: &HttpRequest{
			Body:    nil,
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/cbor"},
		},
		endpoint: string(EndpointCall),
		body:     submit,
	}
	if len(a.credentials) > 0 {
		request.request.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))
	}
	transformRequest := a.identity.TransformRequest(request)
	body, err := cbor.Marshal(transformRequest.RequestBody())
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v2/canister/%s/call", ecid.ToString())
	resp, err := a.fetch(path, request.request, body)
	if err != nil {
		return nil, err
	}
	var response ainterface.Response
	err = cbor.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, xerrors.Errorf("Server returned an error:", response.Status, response.StatusText)
	}
	requestId := ainterface.RequestIdOf(submit)

	return &ainterface.SubmitResponse{
		RequestId: requestId,
		Response:  response,
	}, nil
}

func (a *HttpAgent) Status() ([]byte, error) {
	request := &HttpRequest{
		Method: "GET",
		Body:   nil,
	}
	if len(a.credentials) > 0 {
		request.Headers = map[string]string{"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))}
	}
	resp, err := a.fetch("/api/v2/status", request, nil)
	if err != nil {
		return nil, err
	}
	var response ainterface.StatusResponse
	err = json.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	return response.RootKey, nil
}

func (a *HttpAgent) Query(canisterId *principal.Principal, options *ainterface.QueryFields) (*ainterface.QueryResponse, error) {
	sender := a.identity.GetPrincipal()
	query := &QueryRequest{
		requestType: "query",
		canisterId:  canisterId,
		method:      options.MethodName,
		arg:         options.Arg,
		sender:      sender,
		expiry:      NewExpiry(DEFAULT_INGRESS_EXPIRY_DELTA_IN_MSECS),
	}
	request := &HttpAgentRequest{
		request: &HttpRequest{
			Body:    nil,
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/cbor"},
		},
		endpoint: string(EndpointQuery),
		body:     query,
	}
	if len(a.credentials) > 0 {
		request.request.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))
	}
	transformRequest := a.identity.TransformRequest(request)
	body, err := cbor.Marshal(transformRequest.RequestBody())
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v2/canister/%s/query", canisterId.ToString())
	resp, err := a.fetch(path, request.request, body)
	if err != nil {
		return nil, err
	}
	var response ainterface.QueryResponse
	err = cbor.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, xerrors.Errorf("Server returned an error:", response.Status, response.StatusText)
	}

	return &response, nil
}

func (a *HttpAgent) FetchRootKey() ([]byte, error) {
	if !a.rootKeyFetched {
		bytes, err := a.Status()
		if err != nil {
			return nil, err
		}
		a.rootKey = bytes
	}
	return a.rootKey, nil
}
