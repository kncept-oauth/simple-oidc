package dispatcher

import (
	"fmt"
	"net/http"

	"github.com/klauspost/compress/gzhttp"
	servicedispatcher "github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/testharness/webcontent"
)

func NewApplication(daoSource servicedispatcher.DaoSource) (http.HandlerFunc, error) {
	fmt.Printf("New Testharness Application\n")

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.HandlerFunc(Index))
	handler := gzhttp.GzipHandler(serveMux)
	return handler, nil
}

func Index(w http.ResponseWriter, r *http.Request) {
	data, err := webcontent.Fs.ReadFile("index.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("%v", err)))
	}

	w.WriteHeader(200)
	w.Write(data)
}
