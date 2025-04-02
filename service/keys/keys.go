package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Keystore interface {
	GetKey(kid string) (*JwkKeypair, error)
	SaveKey(keypair *JwkKeypair) error
}

type JwkKeypair struct {
	Pvt    *rsa.PrivateKey
	Kid    string
	Expiry *time.Time
}

func NewKeyId(prefix string) string {
	return fmt.Sprintf("%v-%v", prefix, uuid.NewString())
}
func KeyIdPrefix(keyId string) string {
	idx := strings.Index(keyId, "-")
	if idx == -1 || idx == 0 || idx == len(keyId) {
		return ""
	}
	return keyId[:idx]
}

type JwkDetails struct {
	Kty string `json:"kty"` // "RSA"
	Kid string `json:"kid"` // Key ID
	Use string `json:"use"` // sig
	Alg string `json:"alg"` // RS512
	N   string `json:"n"`   // n value
	E   string `json:"e"`   // e (exponent) value
}

func (obj JwkDetails) ToPublicKey() (*rsa.PublicKey, error) {
	if obj.Kty != "RSA" {
		return nil, fmt.Errorf("only rsa key type supported")
	}
	eb, err := base64.RawURLEncoding.WithPadding(base64.NoPadding).DecodeString(obj.E)
	if err != nil {
		return nil, err
	}
	e := big.NewInt(0).SetBytes(eb)

	nb, err := base64.RawURLEncoding.WithPadding(base64.NoPadding).DecodeString(obj.N)
	if err != nil {
		return nil, err
	}
	n := big.NewInt(0).SetBytes(nb)

	return &rsa.PublicKey{
		E: int(e.Int64()),
		N: n,
	}, nil
}

func GenerateRsaKey() (*rsa.PrivateKey, error) {
	keySize := 4096
	return rsa.GenerateKey(rand.Reader, keySize)
}

func PemEncodeRsaPrivate(key *rsa.PrivateKey) string {
	privatePem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		})
	return string(privatePem)
}
func PemEncodeRsaPublic(key *rsa.PublicKey) string {
	privatePem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(key),
		})
	return string(privatePem)
}

func RsaToJwk(keyId string, rsa *rsa.PublicKey) *JwkDetails {
	fmt.Printf("rsa.Size() %v\n", rsa.Size())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(rsa.E)).Bytes())
	n := base64.RawURLEncoding.EncodeToString(rsa.N.Bytes())
	return &JwkDetails{
		Kty: "RSA",
		Kid: keyId,
		Use: "sig",
		Alg: "RS512",
		N:   n,
		E:   e,
	}
}

func EcToJwt(keyId string, key interface{}) *JwkDetails {
	exampleJson := `{
      "kty": "RSA",
      "e": "xxx",
      "use": "sig",
      "kid": "xxx",
      "x5t": "xx",
      "x5c": [
        "xxx"
      ],
      "n": "xxx
}`
	panic(fmt.Errorf("unimplemented - need something like %v", exampleJson))
}
