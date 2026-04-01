package dispatcher

import (
	"net/http"
	"strings"

	"github.com/klauspost/compress/gzhttp"

	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher/httpdispatcher"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher/oapidispatcher"
	"github.com/kncept-oauth/simple-oidc/service/dispatcherauth"
	"github.com/kncept-oauth/simple-oidc/service/gen/api"
)

// optional config options

func NewApplication(
	daoSource dao.DaoSource,
	urlPrefix string,
	devModeLiveFilesystemBase *string,
) (http.HandlerFunc, error) {
	urlPrefix = strings.TrimSuffix(urlPrefix, "/")

	serveMux := httpdispatcher.NewAcceptOidcHandler(daoSource, urlPrefix, devModeLiveFilesystemBase)
	staticFileHandler := httpdispatcher.NewStaticFilesDispatcher(devModeLiveFilesystemBase)
	templateHandler := httpdispatcher.NewTemplateDispatcher(devModeLiveFilesystemBase)

	openApiHandler := oapidispatcher.NewOapiDispatcher(daoSource, urlPrefix)

	server, err := api.NewServer(
		openApiHandler,
		&dispatcherauth.Handler{}, // auth handler
		api.WithNotFound(func(w http.ResponseWriter, r *http.Request) {

			if strings.HasSuffix(r.URL.Path, ".js") {
				staticFileHandler.RespondWithStaticFile(r.URL.Path, "application/javascript", 200)(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, ".css") {
				staticFileHandler.RespondWithStaticFile(r.URL.Path, "text/css", 200)(w, r)
				return
			}

			if r.URL.Path == "/" {
				// TODO: detect login, and set flag (eg: display login or account details link)
				templateHandler.RespondWithTemplate("index.html", 200, w, nil)
			} else {
				w.WriteHeader(404)
				// acceptOidcHandler.respondWithTemplate("notfound.html", 404, w, nil) // todo
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	serveMux.Handle("/", server)
	handler := gzhttp.GzipHandler(serveMux)
	return handler, err
}
