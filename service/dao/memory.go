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
	clientStore  *memClientStore
	keyStore     *memKeyStore
	userStore    *memUserStore
	sessionStore *memSessionStore
}

// GetKeyStore implements dispatcher.DaoSource.
func (obj *MemoryDao) GetKeyStore() keys.Keystore {
	if obj.keyStore == nil {
		obj.keyStore = &memKeyStore{}
	}
	return obj.keyStore
}

func NewMemoryDao() dispatcher.DaoSource {
	return &MemoryDao{}
}

func (obj *MemoryDao) GetClientStore() authorizer.ClientStore {
	if obj.clientStore == nil {
		obj.clientStore = &memClientStore{}
	}
	return obj.clientStore
}

func (obj *MemoryDao) GetUserStore() users.UserStore {
	if obj.userStore == nil {
		obj.userStore = &memUserStore{}
	}
	return obj.userStore
}

func (obj *MemoryDao) GetSessionStore() session.SessionStore {
	if obj.sessionStore == nil {
		obj.sessionStore = &memSessionStore{}
	}
	return obj.sessionStore
}

type memClientStore struct {
	clients sync.Map
}

type memKeyStore struct {
	keys sync.Map
}

type memUserStore struct {
	users sync.Map
}

type memSessionStore struct {
	sessions sync.Map
}

// GetKey implements keys.Keystore.
func (m *memKeyStore) GetKey(kid string) (*keys.JwkKeypair, error) {
	keypair, ok := m.keys.Load(kid)
	if ok {
		return keypair.(*keys.JwkKeypair), nil
	}
	return nil, nil
}

// SaveKey implements keys.Keystore.
func (m *memKeyStore) SaveKey(keypair *keys.JwkKeypair) error {
	m.keys.Store(keypair.Kid, keypair)
	return nil
}

func (m *memKeyStore) ListKeys() ([]string, error) {
	keys := make([]string, 0)
	m.keys.Range(func(key any, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys, nil
}

// GetClient implements authorizer.ClientStore.
func (c *memClientStore) Get(clientId string) (authorizer.Client, error) {
	client, ok := c.clients.Load(clientId)
	if ok {
		return client.(authorizer.Client), nil
	}
	return nil, nil
}

// Save implements authorizer.ClientStore.
func (c *memClientStore) Save(client authorizer.ClientStruct) error {
	existing, err := c.Get(client.ClientId)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("client already exists: %v", client.ClientId)
	}
	c.clients.Store(client.ClientId, client)
	return nil
}

func (c *memClientStore) List() ([]authorizer.Client, error) {
	clients := make([]authorizer.Client, 0)
	c.clients.Range(func(key, value any) bool {
		if client, ok := value.(authorizer.Client); ok {
			clients = append(clients, client)
		}
		return true
	})
	return clients, nil
}

func (obj *memUserStore) GetUser(id string) (*users.OidcUser, error) {
	val, ok := obj.users.Load(id)
	if !ok {
		return nil, nil
	}
	return val.(*users.OidcUser), nil
}
func (obj *memUserStore) SaveUser(user *users.OidcUser) error {
	obj.users.Store(user.Id, user)
	return nil
}

func (obj *memSessionStore) Save(session *session.Session) error {
	obj.sessions.Store(session.SessionId, session)
	return nil
}
func (obj *memSessionStore) Load(sessionId string) (*session.Session, error) {
	sessionObj, _ := obj.sessions.Load(sessionId)
	return sessionObj.(*session.Session), nil
}
