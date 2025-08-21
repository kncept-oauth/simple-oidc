package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kncept-oauth/simple-oidc/service/users"

	cjwt "github.com/cristalhq/jwt/v5"
)

type Session struct {
	SessionId string `dynamodbav:"id"`
	UserId    string `dynamodbav:"userId"`

	Fingerprint string `dynamodbav:"fingerprint"` // device fingerprint

	Created      time.Time `dynamodbav:"created"`
	Refreshed    time.Time `dynamodbav:"refreshed"`
	IssueCounter int64     `dynamodbav:"issues"`
	// TODO: JWT the RefreshCode
	RefreshCode string `dynamodbav:"refreshCode"` // refresh code needs to match when extracted from the RefreshToken JWT
}

// TODO: these two should _really_ be a single function
func (obj Session) MakeAuthTokenJwt(u *users.OidcUser, issuer string, audience ...string) *AuthTokenJwt {
	if len(audience) == 0 {
		panic("Must supply at least one audience")
	}
	jwt := &AuthTokenJwt{
		Audience:  audience,
		Issuer:    issuer,
		Sub:       obj.UserId,
		Exp:       cjwt.NewNumericDate(time.Now().UTC().Add(9 * time.Hour).Truncate(time.Second)),
		SessionId: uuid.NewString(),
	}

	// fill in name, email, and phone from user u _if present_

	return jwt
}
func (obj *Session) MakeRefreshTokenJwt(authTokenJwt AuthTokenJwt) *RefreshTokenJwt {
	obj.RefreshCode = uuid.NewString()
	obj.IssueCounter = obj.IssueCounter + obj.IssueCounter
	return &RefreshTokenJwt{
		Audience:    authTokenJwt.Audience,
		Issuer:      authTokenJwt.Issuer,
		SessionId:   obj.SessionId,
		RefreshCode: obj.RefreshCode,
		Nbf:         cjwt.NewNumericDate(authTokenJwt.Exp.Add(-time.Hour)),
		Exp:         cjwt.NewNumericDate(authTokenJwt.Exp.Add(7 * 24 * time.Hour)),
	}
}

// see https://openid.net/specs/openid-connect-basic-1_0.html#StandardClaims
//
// for additional claims, see https://www.iana.org/assignments/jwt/jwt.xhtml
type AuthTokenJwt struct {
	Audience  cjwt.Audience     `json:"aud"`
	Issuer    string            `json:"iss"`
	Sub       string            `json:"sub"`
	Exp       *cjwt.NumericDate `json:"exp"`
	SessionId string            `json:"sid"`
	// JwtId     string            `json:"jti"`

	Name          *string `json:"name,omitempty"`
	Email         *string `json:"email,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	Phone         *string `json:"phone_number,omitempty"`
	PhoneVerified *bool   `json:"phone_number_verified,omitempty"`
	// session id
	// and any custom claims??
}
type RefreshTokenJwt struct {
	Audience    cjwt.Audience `json:"aud,omitempty"`
	Issuer      string        `json:"iss,omitempty"`
	SessionId   string
	RefreshCode string
	Exp         *cjwt.NumericDate `json:"exp,omitempty"`
	Nbf         *cjwt.NumericDate `json:"nbf,omitempty"`
}

type SessionStore interface {
	SaveSession(ctx context.Context, session *Session) error
	LoadSession(ctx context.Context, sessionId string, userId string) (*Session, error)
	ListUserSessions(ctx context.Context, userId string) ([]*Session, error)
}

func NewSession(userId string) *Session {
	now := time.Now().UTC()
	return &Session{
		UserId:       userId,
		SessionId:    uuid.NewString(),
		Created:      now,
		Refreshed:    now,
		IssueCounter: 1,
		RefreshCode:  uuid.NewString(),
	}
}

// TODO: ExchanceRegreshToken operation
