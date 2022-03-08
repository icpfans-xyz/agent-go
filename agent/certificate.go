package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/icpfans-xyz/agent-go/principal/utils"
)

type NodeId int

const (
	Empty NodeId = iota
	Fork
	Labeled
	Leaf
	Pruned
)

// type HashTree struct {
// 	ID    NodeId    `cbor:"id"`
// 	Label []byte    `cbor:"label"`
// 	Left  *HashTree `cbor:"left"`
// 	Right *HashTree `cbor:"right"`
// }

type HashTree []interface{}

func HashTreeToString(tree HashTree) string {
	indentFunc := func(s string) string {
		strs := strings.Split(s, "\n")
		for idx, str := range strs {
			strs[idx] = "  " + str
		}
		return strings.Join(strs, "\n")
	}

	labelToString := func(label []byte) string {
		bytes, err := json.Marshal(label)
		if err != nil {
			return fmt.Sprintf("data(...%d bytes)", len(label))
		}
		return string(bytes)
	}

	switch tree[0].(NodeId) {
	case Empty:
		return "()"
	case Fork:
		left := HashTreeToString(tree[1].(HashTree))
		right := HashTreeToString(tree[2].(HashTree))
		return fmt.Sprintf("sub(\n left:\n%s\n---\n right:\n%s\n)", indentFunc(left), indentFunc(right))
	case Labeled:
		label := labelToString(tree[1].([]byte))
		sub := HashTreeToString(tree[2].(HashTree))
		return fmt.Sprintf("label(\n label:\n%s sub:\n%s\n)", label, sub)
	case Leaf:
		return fmt.Sprintf("leaf(...%d bytes", len(tree[1].([]byte)))
	case Pruned:
		return fmt.Sprintf("pruned(%s)", utils.Hex(tree[1].([]byte)))
	default:
		return fmt.Sprintf("unknown(%v)", tree[0])
	}
}

type Delegation struct {
	SubnetId    []byte `cbor:"subnet_id"`
	Certificate []byte `cbor:"certificate"`
}

type Cert struct {
	Tree HashTree `cbor:"tree"`

	Signature  []byte     `cbor:"signature"`
	Delegation Delegation `cbor:"delegation"`
}

type Certificate struct {
	cert     *Cert
	verified bool
	rootKey  []byte
}

func NewCertificate(resp ReadStateResponse, agent Agent) (*Certificate, error) {
	var cert Cert
	err := cbor.Unmarshal(resp.Certificate, &cert)
	if err != nil {
		return nil, err
	}
	return &Certificate{
		cert: &cert,
	}, nil
}

func (c *Certificate) checkState() error {
	if !c.verified {
		return errors.New("Cannot lookup unverified certificate. Call 'verify()' first.")
	}
	return nil
}

func (c *Certificate) flattenForks(tree HashTree) []HashTree {
	if tree[0] == Empty {
		return []HashTree{}
	} else if tree[0] == Fork {
		return append(c.flattenForks(tree[1].(HashTree)), c.flattenForks(tree[2].(HashTree))...)
	} else {
		return []HashTree{tree}
	}
}

func (c *Certificate) findLabel(label []byte, trees []HashTree) (HashTree, error) {
	if len(label) == 0 {
		return nil, errors.New("undefined")
	}
	for _, tree := range trees {
		if tree[0] == Labeled {
			if bytes.Equal(label, tree[1].([]byte)) {
				return tree, nil
			}
		}
	}
	return nil, errors.New("undefined")
}

func (c *Certificate) lookupPath(path [][]byte, tree HashTree) ([]byte, error) {
	if len(path) == 0 {
		if tree[0] == Leaf {
			return tree[1].([]byte), nil
		} else {
			return nil, errors.New("undefined")
		}
	}
	label := path[0]
	t, err := c.findLabel(label, c.flattenForks(tree))
	if err != nil {
		return nil, err
	}
	if t != nil {
		return c.lookupPath(path[1:], t)
	}

	return nil, nil
}

func (c *Certificate) Lookup(path [][]byte) ([]byte, error) {
	if err := c.checkState(); err != nil {
		return nil, err
	}
	return c.lookupPath(path, c.cert.Tree)
}

func (c *Certificate) Verify() bool {
	// root, err := Reconstruct(c.cert.Tree)
	// if err != nil {
	// 	return false
	// }

	//TODO: verify root bls
	// const rootHash = await reconstruct(this.cert.tree);
	// const derKey = await this._checkDelegation(this.cert.delegation);
	// const sig = this.cert.signature;
	// const key = extractDER(derKey);
	// const msg = concat(domain_sep('ic-state-root'), rootHash);
	// const res = await blsVerify(new Uint8Array(key), new Uint8Array(sig), new Uint8Array(msg));
	// this.verified = res;

	c.verified = true
	return true
}

var (
	DER_PREFIX, _ = utils.FromHex("308182301d060d2b0601040182dc7c0503010201060c2b0601040182dc7c05030201036100")
)

const KEY_LENGTH = 96

func ExtractDER(der []byte) ([]byte, error) {
	prefixLen := len(DER_PREFIX)
	expectedLength := prefixLen + KEY_LENGTH
	if expectedLength != len(der) {
		return nil, errors.New("BLS DER-encoded public key must be ${expectedLength} bytes long")
	}
	prefix := der[:prefixLen]
	if bytes.Equal(prefix, DER_PREFIX) {
		return nil, errors.New("BLS DER-encoded public key is invalid. Expect the following prefix: ${DER_PREFIX}, but get ${prefix}")
	}
	return der[prefixLen:], nil
}

func Reconstruct(tree HashTree) ([]byte, error) {
	switch tree[0].(NodeId) {
	case Empty:
		return utils.Sha256(domainSep("ic-hashtree-empty")), nil
	case Fork:
		left, err := Reconstruct(tree[1].(HashTree))
		if err != nil {
			return nil, err
		}
		right, err := Reconstruct(tree[2].(HashTree))
		if err != nil {
			return nil, err
		}
		return utils.Sha256(domainSep("ic-hashtree-fork"), left, right), nil
	case Labeled:
		bytes, err := Reconstruct(tree[2].(HashTree))
		if err != nil {
			return nil, err
		}
		return utils.Sha256(domainSep("ic-hashtree-labeled"), tree[1].([]byte), bytes), nil
	case Leaf:
		return utils.Sha256(domainSep("ic-hashtree-leaf"), tree[1].([]byte)), nil
	case Pruned:
		return tree[1].([]byte), nil
	default:
		return nil, fmt.Errorf("unknown(%v)", tree[0])
	}
}

func domainSep(s string) []byte {
	return append(utils.FromUint32(uint32(len(s))), []byte(s)...)
}

func LookupPath(path [][]byte, t HashTree) ([]byte, error) {
	if len(path) == 0 {
		if t[0] == Leaf {
			return t[1].([]byte), nil
		}
	} else {
		return nil, errors.New("undefined")
	}
	ts, err := flattenForks(t)
	if err != nil {
		return nil, err
	}
	tree, err := findLabel(path[0], ts)
	if err != nil {
		return nil, err
	}
	return LookupPath(path[1:], tree)
}

func flattenForks(t HashTree) ([]HashTree, error) {
	if t[0] == Empty {
		return []HashTree{}, nil
	} else if t[0] == Fork {
		lt, err := flattenForks(t[1].(HashTree))
		if err != nil {
			return nil, err
		}
		rt, err := flattenForks(t[2].(HashTree))
		if err != nil {
			return nil, err
		}
		return append(lt, rt...), nil
	} else {
		return []HashTree{t}, nil
	}
}

func findLabel(l []byte, ts []HashTree) (HashTree, error) {
	if len(ts) == 0 {
		return nil, errors.New("undefined")
	}
	for _, t := range ts {
		if t[0] == Labeled {
			if bytes.Equal(l, t[1].([]byte)) {
				return t[1].(HashTree), nil
			}
		}
	}
	return nil, errors.New("undefined")
}
