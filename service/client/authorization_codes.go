package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
)

type AuthorizationCodeStore interface {
	SaveAuthorizationCode(ctx context.Context, code *AuthorizationCode) error
	GetAuthorizationCode(ctx context.Context, code string) (*AuthorizationCode, error)
}

type AuthorizationCode struct {
	Code       string     `dynamodbav:"code"` // partition key
	UserId     string     `dynamodbav:"userId"`
	Expiry     *time.Time `dynamodbav:"expiry"`
	OidcParams string     `dynamodbav:"params"`
	// session id
}

func (ac *AuthorizationCode) Created() (time.Time, error) {
	k, err := ksuid.Parse(ac.Code[0:27])
	if err != nil {
		return time.Time{}, err
	}
	return k.Time(), nil
}

func NewAuthorizationCode(userId string, oidcParams string) (*AuthorizationCode, error) {
	k, err := ksuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &AuthorizationCode{
		Code:       fmt.Sprintf("%v-%v", k, uuid.New()),
		UserId:     userId,
		OidcParams: oidcParams,
	}, nil
}
