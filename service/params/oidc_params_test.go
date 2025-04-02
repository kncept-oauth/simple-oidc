package params

import (
	"testing"
)

func TestStructFromMap(t *testing.T) {
	testMap := map[string]string{
		"ignore1":       "ignore1Val",
		"response_type": "response_type_val",
		"state":         "confusion",
	}
	v := OidcParamsFromMap(testMap)
	if v.ResponseType != "response_type_val" {
		t.Fatalf("Unexpected response type: %v", v.ResponseType)
	}
	if v.ToQueryParams() != "response_type=response_type_val&client_id=&scope=&redirect_uri=&state=confusion" {
		t.Fatalf("Unexpected query params: %v", v.ToQueryParams())
	}

	v2, err := OidcParamsFromQuery(v.ToQueryParams())
	if err != nil {
		t.Fatalf("%v", err)
	}
	if v.ToQueryParams() != v2.ToQueryParams() {
		t.Fatalf("Incorrect reconstituted query params: %v", v2.ToQueryParams())
	}
}

func TestEmptyOidcParamsFromQuery(t *testing.T) {
	v, err := OidcParamsFromQuery("")
	if err != nil {
		t.Fatalf("%v", err)
	}
	if v.ToQueryParams() != "response_type=&client_id=&scope=&redirect_uri=" {
		t.Fatalf("Unexpected query params: %v", v.ToQueryParams())
	}
}
