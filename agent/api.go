package agent

import (
	"github.com/icpfans-xyz/agent-go/principal"
)

type ReplicaRejectCode uint64

const (
	SysFatal ReplicaRejectCode = 1
	SysTransient
	DestinationInvalid
	CanisterReject
	CanisterError
)

type ReadStateOptions struct {
	Paths [][][]byte
}

// type Request interface {
// 	Endpoint() string
// 	RequestBody() interface{}
// }

type ReadStateResponse struct {
	Response
	Certificate []byte
}

type CallOptions struct {
	/**
	 * The method name to call.
	 */
	MethodName string

	/**
	 * A binary encoded argument. This is already encoded and will be sent as is.
	 */
	Arg []byte

	/**
	 * An effective canister ID, used for routing. This should only be mentioned if
	 * it's different from the canister ID.
	 */
	EffectiveCanisterId *principal.Principal
}

type Response struct {
	OK         bool
	Status     int
	StatusText string
}

type SubmitResponse struct {
	Response
	RequestId RequestId
}

type QueryFields struct {
	/**
	 * The method name to call.
	 */
	MethodName string

	/**
	 * A binary encoded argument. This is already encoded and will be sent as is.
	 */
	Arg []byte
}

const (
	QueryResponseStatusReplied  = "replied"
	QueryResponseStatusRejected = "rejected"
)

type StatusResponse struct {
	IcApiVersion        string                 `cbor:"ic_api_version"`
	ImplSource          string                 `cbor:"impl_source,omitempty"`
	ImplVersion         string                 `cbor:"impl_version,omitempty"`
	ImplRevision        string                 `cbor:"impl_revision,omitempty"`
	ReplicaHealthStatus string                 `cbor:"replica_health_status,omitempty"`
	RootKey             []byte                 `cbor:"root_key,omitempty"`
	Values              map[string]interface{} `cbor:"values,omitempty"`
}

type QueryResponse struct {
	// Response
	// Status string
	Status     string            `cbor:"status"`
	Reply      map[string][]byte `cbor:"reply"`
	RejectCode uint64            `cbor:"reject_code"`
	RejectMsg  string            `cbor:"reject_message"`
}

type QueryResponseReplied struct {
	QueryResponse
	Reply []byte
}

type QueryResponseRejected struct {
	QueryResponse

	RejectCode    ReplicaRejectCode
	RejectMessage string
}

type Agent interface {
	RootKey() []byte

	/**
	 * Returns the principal ID associated with this agent (by default). It only shows
	 * the principal of the default identity in the agent, which is the principal used
	 * when calls don't specify it.
	 */
	GetPrincipal() *principal.Principal

	/**
	 * Send a read state query to the replica. This includes a list of paths to return,
	 * and will return a Certificate. This will only reject on communication errors,
	 * but the certificate might contain less information than requested.
	 * @param effectiveCanisterId A Canister ID related to this call.
	 * @param options The options for this call.
	 */
	ReadState(effectiveCanisterId *principal.Principal, options *ReadStateOptions) (*ReadStateResponse, error)

	Call(canisterId *principal.Principal, options *CallOptions) (*SubmitResponse, error)

	/**
	 * Query the status endpoint of the replica. This normally has a few fields that
	 * corresponds to the version of the replica, its root public key, and any other
	 * information made public.
	 * @returns A JsonObject that is essentially a record of fields from the status
	 *     endpoint.
	 */
	Status() ([]byte, error)

	/**
	 * Send a query call to a canister. See
	 * {@link https://sdk.dfinity.org/docs/interface-spec/#http-query | the interface spec}.
	 * @param canisterId The Principal of the Canister to send the query to. Sending a query to
	 *     the management canister is not supported (as it has no meaning from an agent).
	 * @param options Options to use to create and send the query.
	 * @returns The response from the replica. The Promise will only reject when the communication
	 *     failed. If the query itself failed but no protocol errors happened, the response will
	 *     be of type QueryResponseRejected.
	 */
	Query(canisterId *principal.Principal, options *QueryFields) (*QueryResponse, error)

	/**
	 * By default, the agent is configured to talk to the main Internet Computer,
	 * and verifies responses using a hard-coded public key.
	 *
	 * This function will instruct the agent to ask the endpoint for its public
	 * key, and use that instead. This is required when talking to a local test
	 * instance, for example.
	 *
	 * Only use this when you are  _not_ talking to the main Internet Computer,
	 * otherwise you are prone to man-in-the-middle attacks! Do not call this
	 * function by default.
	 */
	FetchRootKey() ([]byte, error)
}
