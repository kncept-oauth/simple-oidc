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
	return &fsClientStore{
		RootDir: obj.RootDir,
	}
}

type fsClientStore struct {
	RootDir string
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
