package jwtutil

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"testing"
	"time"

	cjwt "github.com/cristalhq/jwt/v5"
	"github.com/google/uuid"
)

type ClaimsTestStruct struct {
	Audience  cjwt.Audience     `json:"aud,omitempty"`
	Issuer    string            `json:"iss,omitempty"`
	ExpiresAt *cjwt.NumericDate `json:"exp,omitempty"`

	// session id
	// and any custom claims??
}

var key *rsa.PrivateKey

func rsaKey() *rsa.PrivateKey {
	if key == nil {
		keySize := 4096
		var err error
		key, err = rsa.GenerateKey(rand.Reader, keySize)
		if err != nil {
			panic(err)
		}
	}
	return key
}

func TestClaimsToJwt(t *testing.T) {
	testClaims := &ClaimsTestStruct{
		Audience: cjwt.Audience{
			"audience 1",
			uuid.NewString(),
		},
		Issuer:    "simple oidc test",
		ExpiresAt: cjwt.NewNumericDate(time.Now().UTC().Add(1 * time.Hour).Truncate(time.Second)),
	}

	jwt, err := ClaimsToJwt(testClaims, "test-key", rsaKey())
	if err != nil {
		t.Fatalf("Error generating JWT: %v", err)
	}
	fmt.Printf("%v\n", jwt)

	jwt2, err := ClaimsToJwt(testClaims, "test-key", rsaKey())
	if err != nil {
		t.Fatalf("Error generating JWT: %v", err)
	}
	if jwt != jwt2 {
		fmt.Printf("REGENERATED incorrectly as %v\n", jwt)
		t.Fatalf("Regenerate of JWT had different output, should be deterministic")
	}

	// otherJwt, err := MapClaimsToJwt(map[string]any{
	// 	"iss": "simple oidc test",
	// }, "test-key", rsaKey())
	// if err != nil {
	// 	t.Fatalf("Error generating JWT: %v", err)
	// }
	// fmt.Printf("OTHER JWT is %v\n", otherJwt)

	keyId := JwtKeyId(jwt)
	if keyId != "test-key" {
		t.Fatalf("Unexpected key id: %v", keyId)
	}
	parsedClaims := &ClaimsTestStruct{}
	err = JwtToClaims(jwt, &rsaKey().PublicKey, parsedClaims)
	if err != nil {
		t.Fatalf("unable to parse claims: %v", err)
	}
	// fmt.Printf("t claims %+v\n", testClaims)
	// fmt.Printf("p claims %+v\n", parsedClaims)

	errClaims := &ClaimsTestStruct{}
	key = nil
	err = JwtToClaims(jwt, &rsaKey().PublicKey, errClaims)
	if err == nil {
		fmt.Printf("e claims %+v\n", errClaims)
	}

	if !errors.Is(err, cjwt.ErrInvalidSignature) {
		t.Fatalf("Expected signature is not valid, got %v", err)
	}

}
