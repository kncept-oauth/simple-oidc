package dao

import (
	"fmt"
	"sync"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
)

type MemoryDao struct {
	clientStore *memClientStore
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

type memClientStore struct {
	clients sync.Map
}

// GetClient implements authorizer.ClientStore.
func (c *memClientStore) Get(clientId string) (authorizer.Client, error) {
	client, ok := c.clients.Load(clientId)
	fmt.Printf("loading client id %v\n", clientId)
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
