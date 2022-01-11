package principal

import (
	"errors"
	"regexp"
	"strings"

	"github.com/dfinity/agent-go/principal/utils"
)

const SELF_AUTHENTICATING_SUFFIX = 0x02
const ANONYMOUS_SUFFIX = 0x04

type Principal struct {
	Bytes []byte
}

func NewPrincipal(bytes []byte) *Principal {
	return &Principal{Bytes: bytes}
}

func Anonymous() *Principal {
	return NewPrincipal(make([]byte, ANONYMOUS_SUFFIX))
}

func SelfAuthenticating(publicKey []byte) (*Principal, error) {
	sha := utils.Sha224(publicKey)
	bytes := append(sha[:], SELF_AUTHENTICATING_SUFFIX)
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
	fixedStr := strings.ReplaceAll(lowerStr, "-", "")
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

func (p *Principal) MarshalCBOR() ([]byte, error) {
	return p.ToBytes(), nil
}

func (p *Principal) IsAnonymous() bool {
	return len(p.Bytes) == 1 && p.Bytes[0] == ANONYMOUS_SUFFIX
}

func (p *Principal) ToBytes() []byte {
	return p.Bytes
}

func (p *Principal) ToHex() string {
	return strings.ToUpper(utils.Hex(p.Bytes))
}

func (p *Principal) ToString() string {
	checkSum := utils.FromUint32(utils.Crc32(p.Bytes))
	bytes := append(checkSum, p.Bytes...)
	result := utils.Base32Encode(bytes)
	re := regexp.MustCompile(`.{1,5}`)
	matchs := re.FindAllString(result, -1)
	return strings.Join(matchs, "-")
}

func (p *Principal) Encode() []byte {
	return p.ToBytes()
}

func (p *Principal) Decode(bytes []byte) error {
	p.Bytes = bytes
	return nil
}
