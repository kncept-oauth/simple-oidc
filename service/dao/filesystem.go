package dao

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type FilesystemDao struct {
	RootDir string
}

func writeJson(rootDir string, id string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(rootDir, fmt.Sprintf("%v.json", id)), data, 600)
}

func readJson(rootDir string, id string, val interface{}) error {
	data, err := os.ReadFile(path.Join(rootDir, fmt.Sprintf("%v.json", id)))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, val)
}
func deleteJson(rootDir string, id string) error {
	return os.Remove(path.Join(rootDir, fmt.Sprintf("%v.json", id)))
}
func listDir(rootDir string) ([]string, error) {
	dirs := make([]string, 0)
	found, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}
	for _, f := range found {
		name := f.Name()
		if strings.HasSuffix(name, ".json") {
			dirs = append(dirs, name[:len(name)-5])
		}
	}
	return dirs, nil
}

// GetKeyStore implements dispatcher.DaoSource.
func (obj *FilesystemDao) GetKeyStore() keys.Keystore {
	return &fsKeyStore{
		RootDir: path.Join(obj.RootDir, "keys"),
	}
}

func (obj *FilesystemDao) GetClientStore() client.ClientStore {
	return &fsClientStore{
		RootDir: path.Join(obj.RootDir, "clients"),
	}
}

func (obj *FilesystemDao) GetUserStore() users.UserStore {
	return &fsUserStore{
		RootDir: path.Join(obj.RootDir, "users"),
	}
}

func (obj *FilesystemDao) GetSessionStore() session.SessionStore {
	return &fsSessionStore{
		RootDir: path.Join(obj.RootDir, "session"),
	}
}

func (obj *FilesystemDao) GetClientAuthorizationStore() client.ClientAuthorizationStore {
	return &clientAuthorizationStore{
		RootDir: path.Join(obj.RootDir, "client-authorizations"),
	}
}

// returns things like /tmp/go-build2313914230/b001 in test
func RootDirFromExePath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}
func RootDirFromWorkDir() (string, error) {
	return os.Getwd()
}

func NewFilesystemDao() dispatcher.DaoSource {
	workDir, err := RootDirFromWorkDir()
	if err != nil {
		panic(err)
	}
	return &FilesystemDao{
		RootDir: workDir,
	}
}

type fsClientStore struct {
	RootDir string
}

type fsKeyStore struct {
	RootDir string
}

type fsUserStore struct {
	RootDir string
}

type fsSessionStore struct {
	RootDir string
}

type clientAuthorizationStore struct {
	RootDir string
}

func (c *clientAuthorizationStore) All(scroller client.PaginationScroller) error {
	files, err := listDir(c.RootDir)
	if err != nil {
		return err
	}
	for _, id := range files {
		obj := &client.ClientAuthorization{}
		err = readJson(c.RootDir, id, obj)
		if err != nil {
			return err
		}
		keepScrolling := scroller([]*client.ClientAuthorization{
			obj,
		})
		if !keepScrolling {
			return nil
		}
	}
	return nil
}

func (c *clientAuthorizationStore) ClientAuthorizationsByClient(clientId string, scroller client.PaginationScroller) error {
	return c.All(func(page []*client.ClientAuthorization) bool {
		if page[0].ClientId == clientId {
			scroller(page)
		}
		return true
	})
}

func (c *clientAuthorizationStore) ClientAuthorizationsByUser(userId string, scroller client.PaginationScroller) error {
	return c.All(func(page []*client.ClientAuthorization) bool {
		if page[0].UserId == userId {
			scroller(page)
		}
		return true
	})
}

func (c *clientAuthorizationStore) DeleteClientAuthorization(authorizationSessionId string) error {
	return deleteJson(c.RootDir, authorizationSessionId)
}

func (c *clientAuthorizationStore) GetClientAuthorization(clientId string, userId string) (*client.ClientAuthorization, error) {
	var found *client.ClientAuthorization
	err := c.All(func(page []*client.ClientAuthorization) bool {
		if page[0].ClientId == clientId && page[0].UserId == userId {
			found = page[0]
			return false
		}
		return true
	})
	return found, err
}

func (c *clientAuthorizationStore) SaveClientAuthorization(clientAuthorization *client.ClientAuthorization) error {
	return writeJson(c.RootDir, clientAuthorization.AuthorizationSessionId, clientAuthorization)
}

func (f *fsKeyStore) GetKey(kid string) (val *keys.JwkKeypair, err error) {
	err = readJson(f.RootDir, kid, val)
	return
}

func (f *fsKeyStore) SaveKey(keypair *keys.JwkKeypair) error {
	return writeJson(f.RootDir, keypair.Kid, keypair)
}

func (f *fsKeyStore) ListKeys() ([]string, error) {
	return listDir(f.RootDir)
}

func (c *fsClientStore) GetClient(clientId string) (client.Client, error) {
	val := &client.ClientStruct{}
	err := readJson(c.RootDir, clientId, val)
	return val, err
}

func (c *fsClientStore) SaveClient(client client.ClientStruct) error {
	return writeJson(c.RootDir, client.ClientId, client)
}

func (c *fsClientStore) ListClients() ([]client.Client, error) {
	ids, err := listDir(c.RootDir)
	if err != nil {
		return nil, err
	}
	clients := make([]client.Client, len(ids))
	for idx, id := range ids {
		client, err := c.GetClient(id)
		if err != nil {
			return nil, err
		}
		clients[idx] = client
	}
	return clients, nil
}

func (c *fsUserStore) GetUser(id string) (*users.OidcUser, error) {
	usr := &users.OidcUser{}
	err := readJson(c.RootDir, id, usr)
	return usr, err
}
func (c *fsUserStore) SaveUser(user *users.OidcUser) error {
	return writeJson(c.RootDir, user.Id, user)
}

func (c *fsSessionStore) SaveSession(session *session.Session) error {
	return writeJson(c.RootDir, session.SessionId, session)
}
func (c *fsSessionStore) LoadSession(sessionId string) (*session.Session, error) {
	ses := &session.Session{}
	err := readJson(c.RootDir, sessionId, ses)
	return ses, err
}
