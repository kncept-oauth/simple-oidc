package dao

import (
	"context"
	"fmt"
	"sync"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao/ddbutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type MemoryDao struct {
	clients              sync.Map
	keys                 sync.Map
	users                sync.Map
	sessions             sync.Map
	clientAuthorizations sync.Map
	authorizationCodes   sync.Map
}

func NewMemoryDao() DaoSource {
	return &MemoryDao{}
}

func (obj *MemoryDao) GetClientStore(ctx context.Context) client.ClientStore {
	return obj
}

func (obj *MemoryDao) GetUserStore(ctx context.Context) users.UserStore {
	return obj
}

func (obj *MemoryDao) GetSessionStore(ctx context.Context) session.SessionStore {
	return obj
}

func (obj *MemoryDao) GetClientAuthorizationStore(ctx context.Context) client.ClientAuthorizationStore {
	return obj
}

func (obj *MemoryDao) GetKeyStore(ctx context.Context) keys.Keystore {
	return obj
}

func (obj *MemoryDao) GetAuthorizationCodeStore(ctx context.Context) client.AuthorizationCodeStore {
	return obj
}

func (obj *MemoryDao) GetKey(kid string) (*keys.JwkKeypair, error) {
	keypair, ok := obj.keys.Load(kid)
	if ok {
		return keypair.(*keys.JwkKeypair), nil
	}
	return nil, nil
}

func (obj *MemoryDao) SaveKey(keypair *keys.JwkKeypair) error {
	obj.keys.Store(keypair.Kid, keypair)
	return nil
}

func (obj *MemoryDao) ListKeys() ([]string, error) {
	keys := make([]string, 0)
	obj.keys.Range(func(key any, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys, nil
}

func (obj *MemoryDao) GetClient(ctx context.Context, clientId string) (*client.Client, error) {
	c, ok := obj.clients.Load(clientId)
	if ok {
		return c.(*client.Client), nil
	}
	return nil, nil
}

func (obj *MemoryDao) SaveClient(ctx context.Context, c *client.Client) error {
	existing, err := obj.GetClient(ctx, c.ClientId)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("client already exists: %v", c.ClientId)
	}
	obj.clients.Store(c.ClientId, c)
	return nil
}

func (obj *MemoryDao) ListClients(ctx context.Context) ([]*client.Client, error) {
	clients := make([]*client.Client, 0)
	obj.clients.Range(func(key, value any) bool {
		if c, ok := value.(*client.Client); ok {
			clients = append(clients, c)
		}
		return true
	})
	return clients, nil
}

func (obj *MemoryDao) RemoveClient(ctx context.Context, clientId string) error {
	obj.clients.Delete(clientId)
	return nil
}

func (obj *MemoryDao) GetUser(id string) (*users.OidcUser, error) {
	val, ok := obj.users.Load(id)
	if !ok {
		return nil, nil
	}
	return val.(*users.OidcUser), nil
}
func (obj *MemoryDao) SaveUser(user *users.OidcUser) error {
	obj.users.Store(user.Id, user)
	return nil
}

func (obj *MemoryDao) SaveSession(session *session.Session) error {
	obj.sessions.Store(session.SessionId, session)
	return nil
}
func (obj *MemoryDao) LoadSession(sessionId string) (*session.Session, error) {
	sessionObj, _ := obj.sessions.Load(sessionId)
	return sessionObj.(*session.Session), nil
}

func (obj *MemoryDao) DeleteClientAuthorization(ctx context.Context, userId string, clientId string) error {
	obj.clientAuthorizations.Delete(fmt.Sprintf("%s-%s", userId, clientId))
	return nil
}

func (obj *MemoryDao) SaveClientAuthorization(ctx context.Context, clientAuthorization *client.ClientAuthorization) error {
	obj.clientAuthorizations.Store(fmt.Sprintf("%s-%s", clientAuthorization.UserId, clientAuthorization.ClientId), clientAuthorization)
	return nil
}

func (obj *MemoryDao) ClientAuthorizationsByClient(ctx context.Context, clientId string, scroller ddbutil.SimpleScroller[client.ClientAuthorization]) error {
	obj.clientAuthorizations.Range(func(key, value any) bool {
		ca := value.(*client.ClientAuthorization)
		if ca.ClientId != clientId {
			return true
		}
		return scroller.Scroll([]*client.ClientAuthorization{
			ca,
		})
	})
	return nil
}

func (obj *MemoryDao) ClientAuthorizationsByUser(ctx context.Context, userId string, scroller ddbutil.SimpleScroller[client.ClientAuthorization]) error {
	obj.clientAuthorizations.Range(func(key, value any) bool {
		ca := value.(*client.ClientAuthorization)
		if ca.UserId != userId {
			return true
		}
		return scroller.Scroll([]*client.ClientAuthorization{
			ca,
		})
	})
	return nil
}

func (obj *MemoryDao) GetClientAuthorization(ctx context.Context, userId string, clientId string) (*client.ClientAuthorization, error) {
	value, ok := obj.clientAuthorizations.Load(fmt.Sprintf("%s-%s", userId, clientId))
	if !ok {
		return nil, nil
	}
	return value.(*client.ClientAuthorization), nil
}

func (obj *MemoryDao) GetAuthorizationCode(ctx context.Context, code string) (*client.AuthorizationCode, error) {
	c, _ := obj.clientAuthorizations.Load(code)
	return c.(*client.AuthorizationCode), nil
}

func (obj *MemoryDao) SaveAuthorizationCode(ctx context.Context, code *client.AuthorizationCode) error {
	obj.clientAuthorizations.Store(code.Code, code)
	return nil
}
