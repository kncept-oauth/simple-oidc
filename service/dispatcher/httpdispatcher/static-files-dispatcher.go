package httpdispatcher

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/kncept-oauth/simple-oidc/service/webcontent"
)

var _ StaticFilesDispatcher = (*devModeLiveFilesystem)(nil)
var _ StaticFilesDispatcher = (*embeddedFilesystem)(nil)

type StaticFilesDispatcher interface {
	RespondWithStaticFile(filename string, contentType string, statusCode int) http.HandlerFunc
}

func NewStaticFilesDispatcher(devModeLiveFilesystemBase *string) StaticFilesDispatcher {
	if devModeLiveFilesystemBase != nil {
		return &devModeLiveFilesystem{devModeLiveFilesystemBase: *devModeLiveFilesystemBase}
	}
	return &embeddedFilesystem{}
}

type devModeLiveFilesystem struct {
	devModeLiveFilesystemBase string
}
type embeddedFilesystem struct {
}

func (obj *devModeLiveFilesystem) RespondWithStaticFile(filename string, contentType string, statusCode int) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(filename, "/") || strings.Contains(filename, "..") {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		file, err := os.Open(fmt.Sprintf("%s/webcontent/static%s", obj.devModeLiveFilesystemBase, filename))
		if err == nil {
			fileContent, err := io.ReadAll(file)
			if err == nil {
				if contentType != "" {
					res.Header().Add("Content-Type", contentType)
				}
				res.WriteHeader(statusCode)
				res.Write(fileContent)
				return
			}
		}
		if errors.Is(err, os.ErrNotExist) {
			res.WriteHeader(http.StatusNotFound)
		} else {
			res.WriteHeader(http.StatusInternalServerError)
		}
		return

	}
}

func (obj *embeddedFilesystem) RespondWithStaticFile(filename string, contentType string, statusCode int) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if !strings.HasPrefix(filename, "/") || strings.Contains(filename, "..") {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		file, err := webcontent.Fs.Open(fmt.Sprintf("static%v", filename))
		if err == nil {
			fileContent, err := io.ReadAll(file)
			if err == nil {
				if contentType != "" {
					res.Header().Add("Content-Type", contentType)
				}
				res.WriteHeader(statusCode)
				res.Write(fileContent)
				return
			}
		}
		if errors.Is(err, os.ErrNotExist) {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
	}
}
