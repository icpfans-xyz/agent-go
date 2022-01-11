package agent

type RequestType = string

// DOCS: https://smartcontracts.org/docs/interface-spec/index.html#http-call
type Request struct {
	Type RequestType `cbor:"request_type,omitempty"`
	// The user who issued the request.
	Sender []byte `cbor:"sender,omitempty"`
	// Arbitrary user-provided data, typically randomly generated. This can be
	// used to create distinct requests with otherwise identical fields.
	Nonce []byte `cbor:"nonce,omitempty"`
	// An upper limit on the validity of the request, expressed in nanoseconds
	// since 1970-01-01 (like ic0.time()).
	IngressExpiry uint64 `cbor:"ingress_expiry,omitempty"`
	// The principal of the canister to call.
	CanisterID []byte `cbor:"canister_id"`
	// Name of the canister method to call.
	MethodName string `cbor:"method_name,omitempty"`
	// Argument to pass to the canister method.
	Arguments []byte `cbor:"arg,omitempty"`
	// Paths (sequence of paths): A list of paths, where a path is itself a sequence of blobs.
	Paths [][][]byte `cbor:"paths,omitempty"`
}
