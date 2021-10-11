package principal

import (
	"errors"
	"regexp"
	"strings"

	"github.com/dfinity/agent-go/principal/utils"
)

const SELF_AUTHENTICATING_SUFFIX = 2
const ANONYMOUS_SUFFIX = 4

type Principal struct {
	bytes       []byte
	isPrincipal bool
}

func NewPrincipal(bytes []byte) *Principal {
	return &Principal{bytes: bytes}
}

func Anonymous() *Principal {
	return NewPrincipal(make([]byte, ANONYMOUS_SUFFIX))
}

func SelfAuthenticating(publicKey []byte) (*Principal, error) {
	sha := utils.Sha224(publicKey)
	bytes := append(sha, SELF_AUTHENTICATING_SUFFIX)
	return NewPrincipal(bytes), nil
}

func FromHex(hex string) (*Principal, error) {
	bytes, err := utils.FromHex(hex)
	if err != nil {
		return nil, err
	}
	return NewPrincipal(bytes), nil
}

func FromString(str string) (*Principal, error) {
	lowerStr := strings.ToLower(str)
	regex := regexp.MustCompile(`/-/g`)
	fixedStr := regex.ReplaceAllString(lowerStr, "")
	bytes, err := utils.Base32Decode(fixedStr)
	if err != nil {
		return nil, err
	}
	bytes = bytes[4:]
	p := NewPrincipal(bytes)
	if p.ToString() != str {
		return nil, errors.New("Principal does not have a valid checksum")
	}
	return p, nil
}

func FromBytes(bytes []byte) (*Principal, error) {
	return NewPrincipal(bytes), nil
}

func (p *Principal) IsAnonymous() bool {
	return len(p.bytes) == 1 && p.bytes[0] == ANONYMOUS_SUFFIX
}

func (p *Principal) ToBytes() []byte {
	return p.bytes
}

func (p *Principal) ToHex() string {
	return strings.ToUpper(utils.Hex(p.bytes))
}

func (p *Principal) ToString() string {
	checkSum := utils.FromUint32(utils.Crc32(p.bytes))
	bytes := append(checkSum, p.bytes...)
	regex := regexp.MustCompile(`/.{1,5}/g`)
	matchs := regex.FindStringSubmatch(utils.Base32Encode(bytes))
	return strings.Join(matchs, "-")
}
