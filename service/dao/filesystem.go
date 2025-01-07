package dao

import (
	"os"
	"path/filepath"

	"github.com/kncept-oauth/simple-oidc/service/authorizer"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
)

type FilesystemDao struct {
	RootDir string
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

func (obj *FilesystemDao) GetClientStore() authorizer.ClientStore {
	return &clientStore{
		RootDir: obj.RootDir,
	}
}

type clientStore struct {
	RootDir string
}

// GetClient implements authorizer.ClientStore.
func (c *clientStore) GetClient(clientId string) (authorizer.Client, error) {
	panic("unimplemented")
}

// Save implements authorizer.ClientStore.
func (c *clientStore) Save(client authorizer.ClientStruct) error {
	panic("unimplemented")
}
