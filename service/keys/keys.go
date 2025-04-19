package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
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
	ListKeys() ([]string, error)
	GetKey(kid string) (*JwkKeypair, error)
	SaveKey(keypair *JwkKeypair) error
}

func GetCurrentKey(store Keystore, asof ...time.Time) (*JwkKeypair, error) {
	if len(asof) > 1 {
		panic("must only provie one asof arg")
	}
	if len(asof) != 1 {
		asof = []time.Time{
			time.Now().UTC().Truncate(time.Second),
		}
	}
	keys, err := store.ListKeys()
	if err != nil {
		return nil, err
	}
	for _, kid := range keys {
		key, err := store.GetKey(kid)
		if err != nil {
			return nil, err
		}
		if key.InDate(asof[0]) {
			return key, nil
		}
	}

	key, err := GenerateJwkKeypair(asof[0])
	if err != nil {
		return nil, err
	}
	err = store.SaveKey(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

type JwkKeypair struct {
	Kid string // Key ID
	Kty string // eg: RSA

	Rsa *rsa.PrivateKey
	Pem string //  STORE as a PEM, a struct

	Exp *time.Time // expiry
	Nbf *time.Time // not before time
}

func (jwk *JwkKeypair) InDate(when time.Time) bool {
	if jwk.Nbf != nil && when.Before(*jwk.Nbf) {
		return false
	}
	if jwk.Exp != nil && when.After(*jwk.Exp) {
		return false
	}
	return true
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

func GenerateJwkKeypair(asof ...time.Time) (*JwkKeypair, error) {
	if len(asof) > 1 {
		panic("must only provie one asof arg")
	}
	if len(asof) != 1 {
		asof = []time.Time{
			time.Now().UTC().Truncate(time.Second),
		}
	}
	rsaKey, err := GenerateRsaKey()
	if err != nil {
		return nil, err
	}
	now := asof[0].UTC().Truncate(time.Second)
	exp := now.Add(30 * 24 * time.Hour)
	return &JwkKeypair{
		Kid: uuid.NewString(),
		Kty: "RSA",
		Rsa: rsaKey,
		Pem: "",

		Nbf: &now,
		Exp: &exp,
	}, nil
}

func (key *JwkKeypair) ToJwkDetails() *JwkDetails {
	switch key.Kty {
	case "RSA":
		return JwkFromRsa(key.Kid, &key.Rsa.PublicKey)
	default:
		panic("unable to convert to JwkDetails: " + key.Kty)
	}
}

// used for PUBLIC keys only.
type JwkDetails struct {
	Kty string `json:"kty"` // "RSA" "EC"
	Kid string `json:"kid"` // Key ID
	Use string `json:"use"` // sig
	Alg string `json:"alg"` // RS512

	// valid for RSA Keys only
	N string `json:"n,omitempty"` // n value
	E string `json:"e,omitempty"` // e (exponent) value

	// valid for EC keys only
	Crv string `json:"crv,omitempty"` // curve name
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

// example cache from https://github.com/Spomky-Labs/jose/blob/master/doc/object/jwk.md

// RSA public key
// $jwk = new JWK([
//     'kty' => 'RSA',
//     'n'   => 'sXchDaQebHnPiGvyDOAT4saGEUetSyo9MKLOoWFsueri23bOdgWp4Dy1WlUzewbgBHod5pcM9H95GQRV3JDXboIRROSBigeC5yjU1hGzHHyXss8UDprecbAYxknTcQkhslANGRUZmdTOQ5qTRsLAt6BTYuyvVRdhS8exSZEy_c4gs_7svlJJQ4H9_NxsiIoLwAEk7-Q3UXERGYw_75IDrGA84-lA_-Ct4eTlXHBIY2EaV7t7LjJaynVJCpkv4LKjTTAumiGUIuQhrNhZLuF_RJLqHpM2kgWFLU7-VTdL1VbC2tejvcI2BlMkEpk1BzBZI0KQB0GaDWFLN-aEAw3vRw',
//     'e'   => 'AQAB',
// ]);
// RSA private key
// $jwk = new JWK([
//     'kty' => 'RSA',
//     'n'   => 'sXchDaQebHnPiGvyDOAT4saGEUetSyo9MKLOoWFsueri23bOdgWp4Dy1WlUzewbgBHod5pcM9H95GQRV3JDXboIRROSBigeC5yjU1hGzHHyXss8UDprecbAYxknTcQkhslANGRUZmdTOQ5qTRsLAt6BTYuyvVRdhS8exSZEy_c4gs_7svlJJQ4H9_NxsiIoLwAEk7-Q3UXERGYw_75IDrGA84-lA_-Ct4eTlXHBIY2EaV7t7LjJaynVJCpkv4LKjTTAumiGUIuQhrNhZLuF_RJLqHpM2kgWFLU7-VTdL1VbC2tejvcI2BlMkEpk1BzBZI0KQB0GaDWFLN-aEAw3vRw',
//     'e'   => 'AQAB',
//     'd'   => 'VFCWOqXr8nvZNyaaJLXdnNPXZKRaWCjkU5Q2egQQpTBMwhprMzWzpR8Sxq1OPThh_J6MUD8Z35wky9b8eEO0pwNS8xlh1lOFRRBoNqDIKVOku0aZb-rynq8cxjDTLZQ6Fz7jSjR1Klop-YKaUHc9GsEofQqYruPhzSA-QgajZGPbE_0ZaVDJHfyd7UUBUKunFMScbflYAAOYJqVIVwaYR5zWEEceUjNnTNo_CVSj-VvXLO5VZfCUAVLgW4dpf1SrtZjSt34YLsRarSb127reG_DUwg9Ch-KyvjT1SkHgUWRVGcyly7uvVGRSDwsXypdrNinPA4jlhoNdizK2zF2CWQ',
//     'p'   => '9gY2w6I6S6L0juEKsbeDAwpd9WMfgqFoeA9vEyEUuk4kLwBKcoe1x4HG68ik918hdDSE9vDQSccA3xXHOAFOPJ8R9EeIAbTi1VwBYnbTp87X-xcPWlEPkrdoUKW60tgs1aNd_Nnc9LEVVPMS390zbFxt8TN_biaBgelNgbC95sM',
//     'q'   => 'uKlCKvKv_ZJMVcdIs5vVSU_6cPtYI1ljWytExV_skstvRSNi9r66jdd9-yBhVfuG4shsp2j7rGnIio901RBeHo6TPKWVVykPu1iYhQXw1jIABfw-MVsN-3bQ76WLdt2SDxsHs7q7zPyUyHXmps7ycZ5c72wGkUwNOjYelmkiNS0',
//     'dp'  => 'w0kZbV63cVRvVX6yk3C8cMxo2qCM4Y8nsq1lmMSYhG4EcL6FWbX5h9yuvngs4iLEFk6eALoUS4vIWEwcL4txw9LsWH_zKI-hwoReoP77cOdSL4AVcraHawlkpyd2TWjE5evgbhWtOxnZee3cXJBkAi64Ik6jZxbvk-RR3pEhnCs',
//     'dq'  => 'o_8V14SezckO6CNLKs_btPdFiO9_kC1DsuUTd2LAfIIVeMZ7jn1Gus_Ff7B7IVx3p5KuBGOVF8L-qifLb6nQnLysgHDh132NDioZkhH7mI7hPG-PYE_odApKdnqECHWw0J-F0JWnUd6D2B_1TvF9mXA2Qx-iGYn8OVV1Bsmp6qU',
//     'qi'  => 'eNho5yRBEBxhGBtQRww9QirZsB66TrfFReG_CcteI1aCneT0ELGhYlRlCtUkTRclIfuEPmNsNDPbLoLqqCVznFbvdB7x-Tl-m0l_eFTj2KiqwGqE9PZB9nNTwMVvH3VRRSLWACvPnSiwP8N5Usy-WRXS-V7TbpxIhvepTfE0NNo',
// ]);

// EC public key
// $jwk = new JWK([
//     'kty' => 'EC',
//     'crv' => 'P-521',
//     'x'   => 'AekpBQ8ST8a8VcfVOTNl353vSrDCLLJXmPk06wTjxrrjcBpXp5EOnYG_NjFZ6OvLFV1jSfS9tsz4qUxcWceqwQGk',
//     'y'   => 'ADSmRA43Z1DSNx_RvcLI87cdL07l6jQyyBXMoxVg_l2Th-x3S1WDhjDly79ajL4Kkd0AZMaZmh9ubmf63e3kyMj2',
// ]);
// EC private key
// $jwk = new JWK([
//     'kty' => 'EC',
//     'crv' => 'P-521',
//     'x'   => 'AekpBQ8ST8a8VcfVOTNl353vSrDCLLJXmPk06wTjxrrjcBpXp5EOnYG_NjFZ6OvLFV1jSfS9tsz4qUxcWceqwQGk',
//     'y'   => 'ADSmRA43Z1DSNx_RvcLI87cdL07l6jQyyBXMoxVg_l2Th-x3S1WDhjDly79ajL4Kkd0AZMaZmh9ubmf63e3kyMj2',
//     'd'   => 'AY5pb7A0UFiB3RELSD64fTLOSV_jazdF7fLYyuTw8lOfRhWg6Y6rUrPAxerEzgdRhajnu0ferB0d53vM9mE15j2C',
// ]);

// OKP type --> EdDSA
// A private OKP key
// $public = new JWK([
// 	'kty' => 'OKP',
// 	'crv' => 'Ed25519',
// 	'x'   => '11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo',
//  ]);
// A private OKP key
//  $private = new JWK([
// 	'kty' => 'OKP',
// 	'crv' => 'Ed25519',
// 	'x'   => '11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo',
// 	'd'   => 'nWGxne_9WmC6hEr0kuwsxERJxWl7MmkZcDusAxyuf2A',
//  ]);

func (obj JwkDetails) ToPublicKey() (any, error) {
	if obj.Kty == "RSA" {
		return obj.ToRsaPublicKey()
	}
	return nil, fmt.Errorf("Unable to decode key type: %v", obj.Kty)
}

func (obj JwkDetails) ToRsaPublicKey() (*rsa.PublicKey, error) {
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
func GenerateEcdsaKey() (*ecdsa.PrivateKey, error) {
	curve := elliptic.P521()
	return ecdsa.GenerateKey(curve, rand.Reader)
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

func JwkFromRsa(keyId string, key *rsa.PublicKey) *JwkDetails {
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())
	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	return &JwkDetails{
		Kty: "RSA",
		Kid: keyId,
		Use: "sig",
		Alg: "RS512",

		N: n,
		E: e,
	}
}
func JwkFromEcDSA(keyId string, key *ecdsa.PublicKey) *JwkDetails {
	var crv string
	switch key.Curve {
	case elliptic.P256():
		crv = "P-256"
	case elliptic.P384():
		crv = "P-384"
	case elliptic.P521():
		crv = "P-521"
	default:
		panic(fmt.Errorf("unsupported curve"))
	}
	x := base64.RawURLEncoding.EncodeToString(key.X.Bytes())
	y := base64.RawURLEncoding.EncodeToString(key.Y.Bytes())

	return &JwkDetails{
		Kty: "EC",
		Kid: keyId,
		Use: "sig",
		Alg: "EC",

		Crv: crv,
		X:   x,
		Y:   y,
	}
}

// func JwkToEcdh(keyId string, key *ecdh.PublicKey) *JwkDetails {
// 	// Curve25519
// 	// Curve448
// 	// Ed25519
// 	// Ed448
// 	var crv string
// 	// switch key.Curve() {
// 	// case ecdh.Ed25519():
// 	crv = "Ed25519"
// 	// case elliptic.Ed448():
// 	// crv = "Ed448"
// 	// default:
// 	// panic(fmt.Errorf("unsupported curve %t %v", key.Curve(), key.Curve()))
// 	// }
// 	x := base64.RawURLEncoding.EncodeToString(key.X.Bytes())
// 	y := base64.RawURLEncoding.EncodeToString(key.Y.Bytes())

// 	return &JwkDetails{
// 		Kty: "OKP",
// 		Kid: keyId,
// 		Use: "sig",

// 		Crv: crv,
// 		X:   x,
// 		Y:   y,
// 	}
// }
