package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao/ddbutil"
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
	return os.WriteFile(path.Join(rootDir, fmt.Sprintf("%v.json", id)), data, 0600)
}
func readJson[T interface{}](rootDir string, id string) (*T, error) {
	val := new(T)
	data, err := os.ReadFile(path.Join(rootDir, fmt.Sprintf("%v.json", id)))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal(data, val)
	if err != nil {
		return nil, err
	}
	return val, nil
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

func (obj *FilesystemDao) GetKeyStore(ctx context.Context) keys.Keystore {
	os.Mkdir(path.Join(obj.RootDir, "keys"), 0700)
	return &fsKeyStore{
		RootDir: path.Join(obj.RootDir, "keys"),
	}
}

func (obj *FilesystemDao) GetClientStore(ctx context.Context) client.ClientStore {
	os.Mkdir(path.Join(obj.RootDir, "clients"), 0700)
	return &fsClientStore{
		RootDir: path.Join(obj.RootDir, "clients"),
	}
}

func (obj *FilesystemDao) GetUserStore(ctx context.Context) users.UserStore {
	os.Mkdir(path.Join(obj.RootDir, "users"), 0700)
	return &fsUserStore{
		RootDir: path.Join(obj.RootDir, "users"),
	}
}

func (obj *FilesystemDao) GetSessionStore(ctx context.Context) session.SessionStore {
	os.Mkdir(path.Join(obj.RootDir, "session"), 0700)
	return &fsSessionStore{
		RootDir: path.Join(obj.RootDir, "session"),
	}
}

func (obj *FilesystemDao) GetClientAuthorizationStore(ctx context.Context) client.ClientAuthorizationStore {
	os.Mkdir(path.Join(obj.RootDir, "client-authorizations"), 0700)
	return &clientAuthorizationStore{
		RootDir: path.Join(obj.RootDir, "client-authorizations"),
	}
}

func (obj *FilesystemDao) GetAuthorizationCodeStore(ctx context.Context) client.AuthorizationCodeStore {
	os.Mkdir(path.Join(obj.RootDir, "authorization-codes"), 0700)
	return &authorizationCodeStore{
		RootDir: path.Join(obj.RootDir, "authorization-codes"),
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

func NewFilesystemDao() DaoSource {
	workDir, err := RootDirFromWorkDir()
	if err != nil {
		panic(err)
	}
	workDir = path.Join(workDir, ".data")

	os.Mkdir(workDir, 0700)
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

type authorizationCodeStore struct {
	RootDir string
}

func (c *clientAuthorizationStore) All(scrollFn func(page []*client.ClientAuthorization) bool) error {
	files, err := listDir(c.RootDir)
	if err != nil {
		return err
	}
	for _, id := range files {
		obj, err := readJson[client.ClientAuthorization](c.RootDir, id)
		if err != nil {
			return err
		}
		keepScrolling := scrollFn([]*client.ClientAuthorization{
			obj,
		})
		if !keepScrolling {
			return nil
		}
	}
	return nil
}

func (c *clientAuthorizationStore) ClientAuthorizationsByClient(ctx context.Context, clientId string, scroller ddbutil.SimpleScroller[client.ClientAuthorization]) error {
	return c.All(func(page []*client.ClientAuthorization) bool {
		if page[0].ClientId == clientId {
			scroller.Scroll(page)
		}
		return true
	})
}

func (c *clientAuthorizationStore) ClientAuthorizationsByUser(ctx context.Context, userId string, scroller ddbutil.SimpleScroller[client.ClientAuthorization]) error {
	return c.All(func(page []*client.ClientAuthorization) bool {
		if page[0].UserId == userId {
			scroller.Scroll(page)
		}
		return true
	})
}

func (c *clientAuthorizationStore) DeleteClientAuthorization(ctx context.Context, userId string, clientId string) error {
	return deleteJson(c.RootDir, fmt.Sprintf("%s-%s", userId, clientId))
}

func (c *clientAuthorizationStore) GetClientAuthorization(ctx context.Context, userId string, clientId string) (*client.ClientAuthorization, error) {
	return readJson[client.ClientAuthorization](c.RootDir, fmt.Sprintf("%s-%s", userId, clientId))
}

func (c *clientAuthorizationStore) SaveClientAuthorization(ctx context.Context, clientAuthorization *client.ClientAuthorization) error {
	return writeJson(c.RootDir, fmt.Sprintf("%s-%s", clientAuthorization.UserId, clientAuthorization.ClientId), clientAuthorization)
}

func (f *fsKeyStore) GetKey(ctx context.Context, kid string) (*keys.JwkKeypair, error) {
	return readJson[keys.JwkKeypair](f.RootDir, kid)
}

func (f *fsKeyStore) SaveKey(ctx context.Context, keypair *keys.JwkKeypair) error {
	return writeJson(f.RootDir, keypair.Kid, keypair)
}

func (f *fsKeyStore) ListKeys(ctx context.Context) ([]*keys.JwkKeypair, error) {
	keyIds, err := listDir(f.RootDir)
	if err != nil {
		return nil, err
	}
	keys := make([]*keys.JwkKeypair, len(keyIds))
	for i, id := range keyIds {
		key, err := f.GetKey(ctx, id)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}
	return keys, nil
}

func (c *fsClientStore) GetClient(ctx context.Context, clientId string) (*client.Client, error) {
	return readJson[client.Client](c.RootDir, clientId)
}

func (c *fsClientStore) SaveClient(ctx context.Context, client *client.Client) error {
	return writeJson(c.RootDir, client.ClientId, client)
}

func (c *fsClientStore) ListClients(ctx context.Context) ([]*client.Client, error) {
	ids, err := listDir(c.RootDir)
	if err != nil {
		return nil, err
	}
	clients := make([]*client.Client, len(ids))
	for idx, id := range ids {
		client, err := c.GetClient(ctx, id)
		if err != nil {
			return nil, err
		}
		clients[idx] = client
	}
	return clients, nil
}

func (c *fsClientStore) RemoveClient(ctx context.Context, clientId string) error {
	return deleteJson(c.RootDir, clientId)
}

func (c *fsUserStore) GetUser(ctx context.Context, id string) (*users.OidcUser, error) {
	return readJson[users.OidcUser](c.RootDir, id)

}
func (c *fsUserStore) SaveUser(ctx context.Context, user *users.OidcUser) error {
	return writeJson(c.RootDir, user.Id, user)
}

func (c *fsSessionStore) SaveSession(session *session.Session) error {
	return writeJson(c.RootDir, session.SessionId, session)
}
func (c *fsSessionStore) LoadSession(sessionId string) (*session.Session, error) {
	return readJson[session.Session](c.RootDir, sessionId)
}

func (a *authorizationCodeStore) GetAuthorizationCode(ctx context.Context, code string) (*client.AuthorizationCode, error) {
	return readJson[client.AuthorizationCode](a.RootDir, code)
}

func (a *authorizationCodeStore) SaveAuthorizationCode(ctx context.Context, code *client.AuthorizationCode) error {
	return writeJson(a.RootDir, code.Code, code)
}
