package authorizer

import (
	"testing"

	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

func TestAuthorizerInterface(t *testing.T) {
	isAuthorizationHandler(&Authorizer{})

}

func isAuthorizationHandler(receiver api.AuthorizationHandler) {}

func TestClientInterface(t *testing.T) {
	isClient(&ClientStruct{})
}
func isClient(receiver Client) {}

func TestRegexRedirectUriValidity(t *testing.T) {
	client := &ClientStruct{
		AllowedRedirectUris: []string{
			"http://valid/$",
			"http://wildcard/.*$",
			"https://path/with/uri$",
			"https://path/with/slash/",
			"https://path/with/wildslah/*",
		},
		RegexAllowedRedirectUri: true,
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
	client := &ClientStruct{
		AllowedRedirectUris: []string{
			"http://valid/",
			"http://noslash",
			"https://params?",
		},
		RegexAllowedRedirectUri: false,
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
