package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kncept-oauth/simple-oidc/service/jwtutil"
	"github.com/segmentio/ksuid"
)

type Session struct {
	SessionId string `dynamodbav:"id"`

	UserId   string `dynamodbav:"userId"`
	ClientId string `dynamodbav:"clientId"`

	Fingerprint string `dynamodbav:"fingerprint"` // device fingerprint

	Created      time.Time `dynamodbav:"created"`
	Refreshed    time.Time `dynamodbav:"refreshed"`
	IssueCounter int64     `dynamodbav:"issues"`

	RefreshCode string `dynamodbav:"refreshCode"` // refresh code needs to match when extracted from the RefreshToken JWT
}

// TODO: these two should _really_ be a single function
func (obj Session) IssueTokens(issuer string, audience ...string) (*jwtutil.IdToken, *jwtutil.RefreshClaimsJwt) {
	if len(audience) == 0 {
		panic("Must supply at least one audience")
	}
	obj.RefreshCode = uuid.NewString()
	obj.IssueCounter = obj.IssueCounter + 1
	obj.Refreshed = time.Now()

	expiry := obj.Refreshed.Add(9 * time.Hour)

	idToken := &jwtutil.IdToken{
		MinimalIdToken: jwtutil.MinimalIdToken{
			Iss: issuer,
			Sub: obj.UserId,
			Aud: audience,
			Exp: expiry.Unix(),
			Iat: obj.Refreshed.Unix(),
		},
		AdditionalCustomClaimsIdToken: jwtutil.AdditionalCustomClaimsIdToken{
			Sid: obj.SessionId,
		},
	}
	refreshToken := &jwtutil.RefreshClaimsJwt{
		MinimalIdToken: jwtutil.MinimalIdToken{
			Iss: issuer,
			Sub: obj.UserId,
			Aud: []string{issuer},
			Exp: expiry.Add(7 * 24 * time.Hour).Unix(),
			Iat: obj.Refreshed.Unix(),
		},
		AdditionalRefreshClaims: jwtutil.AdditionalRefreshClaims{
			Nbf:  expiry.Add(-time.Hour).Unix(),
			Ses:  obj.SessionId,
			Code: obj.RefreshCode,
		},
	}
	return idToken, refreshToken
}

type SessionStore interface {
	SaveSession(ctx context.Context, session *Session) error
	LoadSession(ctx context.Context, sessionId string, userId string) (*Session, error)
	ListUserSessions(ctx context.Context, userId string) ([]*Session, error)
}

func NewSession(userId string, clientId string) (*Session, error) {
	now := time.Now()
	k, err := ksuid.NewRandomWithTime(now)
	if err != nil {
		return nil, err
	}
	return &Session{
		SessionId:    k.String(),
		UserId:       userId,
		ClientId:     clientId,
		Created:      now,
		Refreshed:    now,
		IssueCounter: 0,
		RefreshCode:  uuid.NewString(),
	}, nil
}

// TODO: ExchanceRegreshToken operation
