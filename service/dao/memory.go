package dao

import (
	"fmt"
	"sync"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type MemoryDao struct {
	clients  sync.Map
	keys     sync.Map
	users    sync.Map
	sessions sync.Map
}

// GetKeyStore implements dispatcher.DaoSource.
func (obj *MemoryDao) GetKeyStore() keys.Keystore {
	return obj
}

func NewMemoryDao() dispatcher.DaoSource {
	return &MemoryDao{}
}

func (obj *MemoryDao) GetClientStore() authorizer.ClientStore {
	return obj
}

func (obj *MemoryDao) GetUserStore() users.UserStore {
	return obj
}

func (obj *MemoryDao) GetSessionStore() session.SessionStore {
	return obj
}

// GetKey implements keys.Keystore.
func (obj *MemoryDao) GetKey(kid string) (*keys.JwkKeypair, error) {
	keypair, ok := obj.keys.Load(kid)
	if ok {
		return keypair.(*keys.JwkKeypair), nil
	}
	return nil, nil
}

// SaveKey implements keys.Keystore.
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

// GetClient implements authorizer.ClientStore.
func (obj *MemoryDao) GetClient(clientId string) (authorizer.Client, error) {
	client, ok := obj.clients.Load(clientId)
	if ok {
		return client.(authorizer.Client), nil
	}
	return nil, nil
}

// Save implements authorizer.ClientStore.
func (obj *MemoryDao) SaveClient(client authorizer.ClientStruct) error {
	existing, err := obj.GetClient(client.ClientId)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("client already exists: %v", client.ClientId)
	}
	obj.clients.Store(client.ClientId, client)
	return nil
}

func (obj *MemoryDao) ListClients() ([]authorizer.Client, error) {
	clients := make([]authorizer.Client, 0)
	obj.clients.Range(func(key, value any) bool {
		if client, ok := value.(authorizer.Client); ok {
			clients = append(clients, client)
		}
		return true
	})
	return clients, nil
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
