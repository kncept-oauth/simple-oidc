package client

import (
	"testing"
	"time"
)

func TestAuthcodeCreated(t *testing.T) {

	before := time.Now().Truncate(time.Second)

	authCode, err := NewAuthCodeId("oidc params")
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	created, err := authCode.Created()
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	after := time.Now().Truncate(time.Second)

	if after.After(created) {
		t.Fatalf("created timestamp is in the future")
	}
	if before.Before(created) {
		t.Fatalf("created timestamp is in the past")
	}

}
