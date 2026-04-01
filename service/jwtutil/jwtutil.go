package jwtutil

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	cjwt "github.com/cristalhq/jwt/v5"
	"github.com/kncept-oauth/simple-oidc/service/dao"
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

func ParseJwtString(ctx context.Context, jwt string, daoSource dao.DaoSource, issuer string) (*IdClaimsJwt, error) {
	if jwt == "" {
		return nil, nil
	}
	token, err := cjwt.ParseNoVerify([]byte(jwt))
	if err != nil {
		return nil, err
	}
	keyPair, err := daoSource.GetKeyStore(ctx).GetKey(ctx, token.Header().KeyID)
	if err != nil {
		return nil, err
	}
	rsaPrivateKey, err := keyPair.DecodeRsaKey()
	if err != nil {
		return nil, err
	}
	idClaimsJwt := &IdClaimsJwt{}
	err = JwtToClaims(jwt, &rsaPrivateKey.PublicKey, idClaimsJwt)
	if err != nil {
		return nil, err
	}

	// verify things now
	err = idClaimsJwt.Verify(issuer)
	if err != nil {
		return nil, err
	}

	return idClaimsJwt, nil

}

// MINIMAL id claims JWT
type IdClaimsJwt struct {
	Iss string `json:"iss"` // issuer
	Sub string `json:"sub"` // subject
	Aud string `json:"aud"` // audience
	Exp int64  `json:"exp"` // expiry
	Nbf int64  `json:"nbf"` // not before
	Iat int64  `json:"iat"` // issued at
}

func (jwt IdClaimsJwt) Verify(issuer string) error {
	if jwt.Iss != issuer {
		return fmt.Errorf("Issuer")
	}
	return jwt.VerifyDateClaims()
}

func (jwt IdClaimsJwt) VerifyDateClaims() error {
	return jwt.VerifyDateClaimsAsOf(time.Now())
}

func (jwt IdClaimsJwt) VerifyDateClaimsAsOf(nowTime time.Time) error {
	now := nowTime.Unix()
	if jwt.Nbf > now {
		return fmt.Errorf("Not Before")
	}
	if jwt.Exp < now {
		return fmt.Errorf("Expired")
	}
	return nil
}
