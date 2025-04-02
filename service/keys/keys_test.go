package keys

import (
	"strings"
	"testing"
)

func TestCanGenerateRsaKey(t *testing.T) {
	key, err := GenerateRsaKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if key == nil {
		t.Fatalf("no key generated")
	}
}

func TestKeyIdPrefix(t *testing.T) {
	kid := NewKeyId("test")
	if !strings.HasPrefix(kid, "test-") {
		t.Fatalf("incorrect prefix: %v", kid)
	}
	if KeyIdPrefix(kid) != "test" {
		t.Fatalf("incorrect prefix: %v", kid)
	}
}

func TestCanConvertToJwksAndBack(t *testing.T) {
	key, err := GenerateRsaKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	jwk := RsaToJwk(NewKeyId("test"), &key.PublicKey)
	publicKey, err := jwk.ToPublicKey()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if key.E != publicKey.E {
		t.Fatalf("different exponent")
	}
	if key.N.Int64() != publicKey.N.Int64() {
		t.Fatalf("different n")
	}
}
