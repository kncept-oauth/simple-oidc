package client

import "time"

type PaginationScroller func(page []*ClientAuthorization) bool

type DepaginatedScroller struct {
	Results []*ClientAuthorization
}

func (scroller *DepaginatedScroller) Scroller(page []*ClientAuthorization) bool {
	if scroller.Results == nil {
		scroller.Results = make([]*ClientAuthorization, 0)
	}
	scroller.Results = append(scroller.Results, page...)
	return true
}

type ClientAuthorizationStore interface {
	ClientAuthorizationsByUser(userId string, scroller PaginationScroller) error
	ClientAuthorizationsByClient(clientId string, scroller PaginationScroller) error // could be a very large amount.
	GetClientAuthorization(clientId string, userId string) (*ClientAuthorization, error)

	SaveClientAuthorization(clientAuthorization *ClientAuthorization) error
	DeleteClientAuthorization(authorizationSessionId string) error
}

type ClientAuthorization struct {
	ClientId               string
	UserId                 string
	AuthorizedAt           time.Time
	LastRefreshedAt        time.Time
	AuthorizationSessionId string
}
