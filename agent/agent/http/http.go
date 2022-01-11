package http

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
const DEFAULT_INGRESS_EXPIRY_DELTA = time.Minute * 5

// Root public key for the IC, encoded as hex
const IC_ROOT_KEY = "308182301d060d2b0601040182dc7c0503010201060c2b0601040182dc7c05030201036100814" +
	"c0e6ec71fab583b08bd81373c255c3c371b2e84863c98a4f1e08b74235d14fb5d9c0cd546d968" +
	"5f913a0c0b2cc5341583bf4b4392e467db96d65b9bb4cb717112f8472e0d5a4d14505ffd7484" +
	"b01291091c5f87b98883463f98091a0baaae"

type RequestType = string

const (
	RequestTypeCall      RequestType = "call"
	RequestTypeQuery     RequestType = "query"
	RequestTypeReadState RequestType = "read_state"
)

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
	hagent := &HttpAgent{
		pipeline: []HttpAgentRequestTransform{},
	}
	if options.Source != nil {
		hagent.host = options.Source.host
		hagent.identity = options.Source.identity
		hagent.credentials = options.Source.credentials
		hagent.pipeline = options.Source.pipeline
	}
	if len(options.Host) > 0 {
		hagent.host = options.Host
	}
	if options.Identity != nil {
		hagent.identity = options.Identity
	}
	if hagent.identity == nil {
		hagent.identity = agent.NewAnonymousIdentity()
	}
	if options.Credentials != nil {
		hagent.credentials = options.Credentials.Name + ":" + options.Credentials.Password
	}
	return hagent, nil
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

func (a *HttpAgent) ExpiryDate(expiry time.Duration) uint64 {
	if expiry <= 0 {
		expiry = DEFAULT_INGRESS_EXPIRY_DELTA
	}
	return uint64(time.Now().Add(time.Duration(expiry)).UnixNano())
}

func (a *HttpAgent) GetPrincipal() *principal.Principal {
	return a.identity.GetPrincipal()
}

func (a *HttpAgent) ReadState(canisterId *principal.Principal, options *ainterface.ReadStateOptions) (*ainterface.ReadStateResponse, error) {
	sender := a.identity.GetPrincipal()
	state := ainterface.Request{
		Type:          RequestTypeReadState,
		Paths:         options.Paths,
		Sender:        sender.ToBytes(),
		IngressExpiry: a.ExpiryDate(time.Second * 10),
	}

	request := &HttpAgentRequest{
		HttpRequest: &HttpRequest{
			Body:    nil,
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/cbor"},
		},
		Point: string(EndpointReadState),
		Body:  state,
	}
	if len(a.credentials) > 0 {
		request.HttpRequest.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))
	}
	transformRequest, err := a.identity.TransformRequest(request.Body)
	if err != nil {
		return nil, err
	}
	body, err := cbor.Marshal(transformRequest)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v2/canister/%s/read_state", canisterId.ToString())
	resp, err := a.fetch(path, request.HttpRequest, body)
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

	for key, val := range request.Headers {
		req.Header.Set(key, val)
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
	submit := ainterface.Request{
		Type:          RequestTypeCall,
		Sender:        sender.ToBytes(),
		CanisterID:    canisterId.ToBytes(),
		MethodName:    options.MethodName,
		Arguments:     options.Arg,
		IngressExpiry: a.ExpiryDate(0),
	}

	request := &HttpAgentRequest{
		HttpRequest: &HttpRequest{
			Body:    nil,
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/cbor"},
		},
		Point: string(EndpointCall),
		Body:  submit,
	}
	if len(a.credentials) > 0 {
		request.HttpRequest.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))
	}
	transformRequest, err := a.identity.TransformRequest(request.Body)
	if err != nil {
		return nil, err
	}
	body, err := cbor.Marshal(transformRequest.Body)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v2/canister/%s/call", ecid.ToString())
	resp, err := a.fetch(path, request.HttpRequest, body)
	if err != nil {
		return nil, err
	}
	var response ainterface.Response
	err = cbor.Unmarshal(resp, &response)
	if err != nil {
		return nil, fmt.Errorf("faild to parse response:%v, error:%v", string(resp), err)
	}
	if !response.OK {
		return nil, xerrors.Errorf("Server returned an error:%d,%s", response.Status, response.StatusText)
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
	err = cbor.Unmarshal(resp, &response)
	if err != nil {
		return nil, err
	}
	return response.RootKey, nil
}

func (a *HttpAgent) Query(canisterId *principal.Principal, options *ainterface.QueryFields) (*ainterface.QueryResponse, error) {
	sender := a.identity.GetPrincipal()
	query := ainterface.Request{
		Type:          RequestTypeQuery,
		Sender:        sender.ToBytes(),
		CanisterID:    canisterId.ToBytes(),
		MethodName:    options.MethodName,
		Arguments:     options.Arg,
		IngressExpiry: a.ExpiryDate(0),
	}
	request := &HttpAgentRequest{
		HttpRequest: &HttpRequest{
			Body:    nil,
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/cbor"},
		},
		Point: string(EndpointQuery),
		Body:  query,
	}
	if len(a.credentials) > 0 {
		request.HttpRequest.Headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(a.credentials))
	}
	transformRequest, err := a.identity.TransformRequest(request.Body)
	if err != nil {
		return nil, err
	}
	body, err := cbor.Marshal(transformRequest.Body)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v2/canister/%s/query", canisterId.ToString())
	resp, err := a.fetch(path, request.HttpRequest, body)
	if err != nil {
		return nil, err
	}
	var response ainterface.QueryResponse
	err = cbor.Unmarshal(resp, &response)
	if err != nil {
		return nil, fmt.Errorf("faild to parse response:%v, error:%v", string(resp), err)
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
