package agent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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

type HashTree struct {
	ID    NodeId
	Label []byte
	Left  *HashTree
	Right *HashTree
}

func HashTreeToString(tree *HashTree) string {
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
	if tree.ID == Empty {
		return "()"
	} else if tree.ID == Fork {
		left := HashTreeToString(tree.Left)
		right := HashTreeToString(tree.Right)
		return fmt.Sprintf("sub(\n left:\n%s\n---\n right:\n%s\n)", indentFunc(left), indentFunc(right))
	} else if tree.ID == Labeled {
		label := labelToString(tree.Label)
		sub := HashTreeToString(tree.Left)
		return fmt.Sprintf("label(\n label:\n%s sub:\n%s\n)", label, sub)
	} else if tree.ID == Leaf {
		return fmt.Sprintf("leaf(...%d bytes", len(tree.Label))
	} else if tree.ID == Pruned {
		return fmt.Sprintf("pruned(%s)", utils.Hex(tree.Label))
	} else {
		return "unknow()"
	}

}

type Delegation struct {
	SubnetId    []byte
	Certificate []byte
}

type Cert struct {
	Tree *HashTree

	Signature  []byte
	Delegation *Delegation
}

type Certificate struct {
	cert     *Cert
	verified bool
	rootKey  []byte
}

func (c *Certificate) checkState() error {
	if !c.verified {
		return errors.New("Cannot lookup unverified certificate. Call 'verify()' first.")
	}
	return nil
}

func (c *Certificate) flattenForks(tree *HashTree) []*HashTree {
	if tree.ID == Empty {
		return []*HashTree{}
	} else if tree.ID == Fork {
		return append(c.flattenForks(tree.Left), c.flattenForks(tree.Right)...)
	} else {
		return []*HashTree{tree}
	}
}

func (c *Certificate) findLabel(label []byte, trees []*HashTree) (*HashTree, error) {
	if len(label) == 0 {
		return nil, errors.New("undefined")
	}
	for _, tree := range trees {
		if tree.ID == Labeled {
			if bytes.Compare(label, tree.Label) == 0 {
				return tree, nil
			}
		}
	}
	return nil, errors.New("undefined")
}

func (c *Certificate) lookupPath(path [][]byte, tree *HashTree) ([]byte, error) {
	if len(path) == 0 {
		if tree.ID == Leaf {
			return tree.Label, nil
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

func (c *Certificate) verify() bool {
	return false
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
	if bytes.Compare(prefix, DER_PREFIX) != 0 {
		return nil, errors.New("BLS DER-encoded public key is invalid. Expect the following prefix: ${DER_PREFIX}, but get ${prefix}")
	}
	return der[prefixLen:], nil
}

func Reconstruct(t *HashTree) ([]byte, error) {
	if t.ID == Empty {
		return utils.Sha256(domainSep("ic-hashtree-empty")), nil
	} else if t.ID == Pruned {
		return t.Label, nil
	} else if t.ID == Leaf {
		return utils.Sha256(domainSep("ic-hashtree-leaf"), t.Label), nil
	} else if t.ID == Labeled {
		bytes, err := Reconstruct(t.Left)
		if err != nil {
			return nil, err
		}
		return utils.Sha256(domainSep("ic-hashtree-labeled"), t.Label, bytes), nil
	} else if t.ID == Fork {
		left, err := Reconstruct(t.Left)
		if err != nil {
			return nil, err
		}
		right, err := Reconstruct(t.Right)
		if err != nil {
			return nil, err
		}
		return utils.Sha256(domainSep("ic-hashtree-fork"), left, right), nil
	} else {
		return nil, errors.New("undefined")
	}
}

func domainSep(s string) []byte {
	return append(utils.FromUint32(uint32(len(s))), []byte(s)...)
}

func LookupPath(path [][]byte, t *HashTree) ([]byte, error) {
	if len(path) == 0 {
		if t.ID == Leaf {
			return t.Label, nil
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

func flattenForks(t *HashTree) ([]*HashTree, error) {
	if t.ID == Empty {
		return []*HashTree{}, nil
	} else if t.ID == Fork {
		lt, err := flattenForks(t.Left)
		if err != nil {
			return nil, err
		}
		rt, err := flattenForks(t.Right)
		if err != nil {
			return nil, err
		}
		return append(lt, rt...), nil
	} else {
		return []*HashTree{t}, nil
	}
}

func findLabel(l []byte, ts []*HashTree) (*HashTree, error) {
	if len(ts) == 0 {
		return nil, errors.New("undefined")
	}
	for _, t := range ts {
		if t.ID == Labeled {
			if bytes.Compare(l, t.Label) == 0 {
				return t.Left, nil
			}
		}
	}
	return nil, errors.New("undefined")
}
