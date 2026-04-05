package jwtutil

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	cjwt "github.com/cristalhq/jwt/v5"
	"github.com/kncept-oauth/simple-oidc/service/keys"
)

type VerifiableClaims interface {
	Verify(issuer string) error
}

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

func ParseIdToken(ctx context.Context, jwt string, keySource keys.Keystore, issuer string) (*IdToken, error) {
	claims := &IdToken{}
	err := ParseJwt(ctx, jwt, keySource, issuer, claims)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func ParseJwt(ctx context.Context, jwt string, keySource keys.Keystore, issuer string, claims VerifiableClaims) error {
	if jwt == "" {
		return nil
	}
	token, err := cjwt.ParseNoVerify([]byte(jwt))
	if err != nil {
		return err
	}
	keyPair, err := keySource.GetKey(ctx, token.Header().KeyID)
	if err != nil {
		return err
	}
	rsaPrivateKey, err := keyPair.DecodeRsaKey()
	if err != nil {
		return err
	}
	err = JwtToClaims(jwt, &rsaPrivateKey.PublicKey, claims)
	if err != nil {
		return err
	}

	err = claims.Verify(issuer)
	if err != nil {
		return err
	}

	return nil

}

// MINIMAL id claims JWT
// see https://openid.net/specs/openid-connect-basic-1_0.html#StandardClaims
//
// for additional claims, see https://www.iana.org/assignments/jwt/jwt.xhtml

type MinimalIdToken struct {
	Iss string        `json:"iss"` // issuer
	Sub string        `json:"sub"` // subject
	Aud cjwt.Audience `json:"aud"` // audience (may be string OR string[])
	Exp int64         `json:"exp"` // expiry
	Iat int64         `json:"iat"` // issued at
}

type AdditionalStandardClaimsIdToken struct {
	AuthTime int64    `json:"auth_time,omitempty"` // only useful with 'max age' header. Ignore, but here cos it's in the spec
	Nonce    string   `json:"nonce,omitempty"`
	AtHash   string   `json:"at_hash,omitempty"`
	ACR      string   `json:"acr,omitempty"` //  Authentication Context Class Reference
	AMR      []string `json:"amr,omitempty"` //   Authentication Methods References
	AZP      string   `json:"azp,omitempty"` //  Authorized third party
}

type AdditionalCustomClaimsIdToken struct {
	Nbf int64  `json:"nbf,omitempty"` // not before
	Sid string `json:"sid,omitempty"`
	// scopes ?
}

type IdToken struct {
	MinimalIdToken
	AdditionalStandardClaimsIdToken
	AdditionalCustomClaimsIdToken
}

func (jwt IdToken) Verify(issuer string) error {
	if jwt.Iss != issuer {
		return fmt.Errorf("Issuer")
	}
	return jwt.VerifyDateClaims()
}

func (jwt IdToken) VerifyDateClaims() error {
	return jwt.VerifyDateClaimsAsOf(time.Now())
}

func (jwt IdToken) VerifyDateClaimsAsOf(now time.Time) error {
	nowUnix := now.Unix()
	if jwt.Nbf > nowUnix {
		return fmt.Errorf("Not Before")
	}
	if jwt.Iat > nowUnix {
		return fmt.Errorf("Issued At")
	}
	if jwt.Exp < nowUnix {
		return fmt.Errorf("Expired")
	}
	return nil
}

type AdditionalRefreshClaims struct {
	Nbf  int64  `json:"nbf"`  // not before
	Ses  string `json:"sid"`  // session ID
	Code string `json:"code"` // SINGLE use lookup code
}

type RefreshClaimsJwt struct {
	MinimalIdToken
	AdditionalRefreshClaims
}

func (jwt RefreshClaimsJwt) Verify(issuer string) error {
	if jwt.Iss != issuer {
		return fmt.Errorf("Issuer")
	}
	return jwt.VerifyDateClaims()
}

func (jwt RefreshClaimsJwt) VerifyDateClaims() error {
	return jwt.VerifyDateClaimsAsOf(time.Now())
}

func (jwt RefreshClaimsJwt) VerifyDateClaimsAsOf(now time.Time) error {
	nowUnix := now.Unix()
	if jwt.Nbf > nowUnix {
		return fmt.Errorf("Not Before")
	}
	if jwt.Exp < nowUnix {
		return fmt.Errorf("Expired")
	}
	return nil
}
