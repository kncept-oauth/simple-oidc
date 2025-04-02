package dao

import (
	"os"
	"path"
	"path/filepath"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type FilesystemDao struct {
	RootDir string
}

// GetKeyStore implements dispatcher.DaoSource.
func (obj *FilesystemDao) GetKeyStore() keys.Keystore {
	return &fsKeyStore{
		RootDir: path.Join(obj.RootDir, "keys"),
	}
}

func (obj *FilesystemDao) GetClientStore() authorizer.ClientStore {
	return &fsClientStore{
		RootDir: path.Join(obj.RootDir, "clients"),
	}
}

func (obj *FilesystemDao) GetUserStore() users.UserStore {
	return &fsUserStore{
		RootDir: path.Join(obj.RootDir, "users"),
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

// GetKey implements keys.Keystore.
func (f *fsKeyStore) GetKey(kid string) (*keys.JwkKeypair, error) {
	panic("unimplemented")
}

// SaveKey implements keys.Keystore.
func (f *fsKeyStore) SaveKey(keypair *keys.JwkKeypair) error {
	panic("unimplemented")
}

func (c *fsClientStore) Get(clientId string) (authorizer.Client, error) {
	panic("unimplemented")
}

func (c *fsClientStore) Save(client authorizer.ClientStruct) error {
	panic("unimplemented")
}

func (c *fsClientStore) List() ([]authorizer.Client, error) {
	panic("unimplemented")
}

func (c *fsUserStore) GetUser(id string) (*users.OidcUser, error) {
	panic("unimplemented")
}
func (c *fsUserStore) SaveUser(user *users.OidcUser) error {
	panic("unimplemented")
}
