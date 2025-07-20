package dispatcher

import (
	"testing"

	"github.com/kncept-oauth/simple-oidc/service/client"
)

func TestRegexRedirectUriValidity(t *testing.T) {
	client := &client.Client{
		AllowedRedirectUris: []string{
			"http://valid/$",
			"http://wildcard/.*$",
			"https://path/with/uri$",
			"https://path/with/slash/",
			"https://path/with/wildslah/*",
		},
		AllowRegexForRedirectUri: true,
	}
	validStrings := []string{
		"http://valid/",
		"http://wildcard/",
		"http://wildcard/123",
		"http://wildcard/123/xyz",
		"https://path/with/uri",
		"https://path/with/slash/",
		"https://path/with/wildslah/*",
	}
	invalidStrings := []string{
		"http://valid",
		"http://valid/nope",
		"https://valid/",
		"http://wildcard",
		"https://path/with/uri/",
		"https://path/with/slash",
	}

	for _, validString := range validStrings {
		valid := isValidRedirectUri(client, validString)
		if !valid {
			t.Errorf("Incorrectly Invalid: %v", validString)
		}
	}

	for _, invalidString := range invalidStrings {
		valid := isValidRedirectUri(client, invalidString)
		if valid {
			t.Errorf("Incorrectly Valid: %v", invalidString)
		}
	}
}

func TestPrefixRedirectUriValidity(t *testing.T) {
	client := &client.Client{
		AllowedRedirectUris: []string{
			"http://valid/",
			"http://noslash",
			"https://params?",
		},
		AllowRegexForRedirectUri: false,
	}

	validStrings := []string{
		"http://valid/",
		"http://valid/yes",
		"https://params?",
		"https://params?a=b",
	}
	invalidStrings := []string{
		"http://valid",
		"https://valid/",
		"http://params?",
		"https://params#?a=b",
	}

	for _, validString := range validStrings {
		valid := isValidRedirectUri(client, validString)
		if !valid {
			t.Errorf("Incorrectly Invalid: %v", validString)
		}
	}

	for _, invalidString := range invalidStrings {
		valid := isValidRedirectUri(client, invalidString)
		if valid {
			t.Errorf("Incorrectly Valid: %v", invalidString)
		}
	}
}
