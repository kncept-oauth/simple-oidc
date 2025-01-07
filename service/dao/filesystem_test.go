package dao

import (
	"testing"

	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
)

func TestIsDaoSource(t *testing.T) {
	assertIsDaoSource(&FilesystemDao{})
}
func assertIsDaoSource(receiver dispatcher.DaoSource) {}

func TestNew(t *testing.T) {
	NewFilesystemDao()
}
