package dao

import (
	"testing"
)

func TestIsDaoSource(t *testing.T) {
	assertIsDaoSource(&FilesystemDao{})
}
func assertIsDaoSource(receiver DaoSource) {}

func TestNew(t *testing.T) {
	NewFilesystemDao()
}
