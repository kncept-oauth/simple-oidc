package jwtutil

import (
	"crypto/rsa"

	cjwt "github.com/cristalhq/jwt/v5"
)

// see https://github.com/lestrrat-go/jwx
// see https://github.com/cristalhq/jwt
func ClaimsToJwt(claims any, keyId string, key *rsa.PrivateKey) (string, error) {
	signer, err := cjwt.NewSignerRS(cjwt.RS512, key)
	if err != nil {
		return "", err
	}
	builder := cjwt.NewBuilder(signer, cjwt.WithKeyID(keyId))
	token, err := builder.Build(claims)
	if err != nil {
		return "", err
	}
	return token.String(), nil
}

// JWT Validation Details
func JwtKeyId(jwt string) string {
	token, err := cjwt.ParseNoVerify([]byte(jwt))
	if err != nil {
		return ""
	}
	return token.Header().KeyID
}
func JwtAlgorithm(jwt string) string {
	token, err := cjwt.ParseNoVerify([]byte(jwt))
	if err != nil {
		return ""
	}
	return token.Header().Algorithm.String()
}
func JwtToClaims(jwt string, key *rsa.PublicKey, dst any) error {
	verifier, err := cjwt.NewVerifierRS(cjwt.RS512, key)
	if err != nil {
		return err
	}
	return cjwt.ParseClaims([]byte(jwt), verifier, dst)

}
