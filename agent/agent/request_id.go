package agent

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"sort"

	"github.com/dfinity/agent-go/candid/utils"
)

var (
	typeKey          = sha256.Sum256([]byte("request_type"))
	canisterIDKey    = sha256.Sum256([]byte("canister_id"))
	nonceKey         = sha256.Sum256([]byte("nonce"))
	methodNameKey    = sha256.Sum256([]byte("method_name"))
	argumentsKey     = sha256.Sum256([]byte("arg"))
	ingressExpiryKey = sha256.Sum256([]byte("ingress_expiry"))
	senderKey        = sha256.Sum256([]byte("sender"))
	pathKey          = sha256.Sum256([]byte("paths"))
)

type RequestId [32]byte

func RequestIdOf(req Request) RequestId {
	var (
		typeHash       = sha256.Sum256([]byte(req.Type))
		canisterIDHash = sha256.Sum256(req.CanisterID)
		methodNameHash = sha256.Sum256([]byte(req.MethodName))
		argumentsHash  = sha256.Sum256(req.Arguments)
	)
	hashes := [][]byte{}
	if len(req.Type) != 0 {
		hashes = append(hashes, append(typeKey[:], typeHash[:]...))
	}
	if req.CanisterID != nil {
		hashes = append(hashes, append(canisterIDKey[:], canisterIDHash[:]...))
	}
	if len(req.MethodName) != 0 {
		hashes = append(hashes, append(methodNameKey[:], methodNameHash[:]...))
	}
	if req.Arguments != nil {
		hashes = append(hashes, append(argumentsKey[:], argumentsHash[:]...))
	}

	if len(req.Sender) != 0 {
		senderHash := sha256.Sum256(req.Sender)
		hashes = append(hashes, append(senderKey[:], senderHash[:]...))
	}
	if req.IngressExpiry != 0 {
		ingressExpiryHash := sha256.Sum256(encodeLEB128(req.IngressExpiry))
		hashes = append(hashes, append(ingressExpiryKey[:], ingressExpiryHash[:]...))
	}
	if len(req.Nonce) != 0 {
		nonceHash := sha256.Sum256(req.Nonce)
		hashes = append(hashes, append(nonceKey[:], nonceHash[:]...))
	}
	if len(req.Paths) != 0 {
		pathHash := encodeList3D(req.Paths)
		hashes = append(hashes, append(pathKey[:], pathHash[:]...))
	}
	sort.Slice(hashes, func(i, j int) bool {
		return bytes.Compare(hashes[i], hashes[j]) == -1
	})
	return sha256.Sum256(bytes.Join(hashes, nil))
}

func encodeLEB128(i uint64) []byte {
	bi := big.NewInt(int64(i))
	e := utils.LebEncode(bi)
	return e
}

func encodeList3D(lists [][][]byte) [32]byte {
	var res []byte
	for _, v := range lists {
		code := encodeList2D(v)
		res = append(res, code[:]...)
	}
	return sha256.Sum256(res)
}

func encodeList2D(lists [][]byte) [32]byte {
	var res []byte
	for _, v := range lists {
		pathBytes := sha256.Sum256(v)
		res = append(res, pathBytes[:]...)
	}
	return sha256.Sum256(res)
}
