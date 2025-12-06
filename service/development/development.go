package development

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/kncept-oauth/simple-oidc/service/crypto"
	"github.com/kncept-oauth/simple-oidc/service/dao"
)

// runs 'ListenAndServeTLS' in a goroutine
func RunLocally(daoSource dao.DaoSource, handler http.Handler) (*http.Server, error) {
	appPort := "8443"

	// TLS Certificates: generate if absent
	generateCerts := false
	var x509Cert, pkcs8PrivateKey []byte
	var err error
	x509Cert, err = os.ReadFile("../service/server.crt")
	if err != nil || len(x509Cert) == 0 {
		generateCerts = true
	}
	if !generateCerts {
		pkcs8PrivateKey, err = os.ReadFile("../service/server.key")
		generateCerts = err != nil || len(pkcs8PrivateKey) == 0
	}
	if generateCerts {
		x509Cert, pkcs8PrivateKey, err = crypto.GenerateTlsCertificate("localhost", crypto.RSA2048)
		if err == nil {
			err = os.WriteFile("../service/server.crt", x509Cert, 0644)
		}
		if err == nil {
			err = os.WriteFile("../service/server.key", pkcs8PrivateKey, 0644)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to generate certificate: %w", err)
	}

	// side url: for root certchain handling see
	// https://stackoverflow.com/questions/63588254/how-to-set-up-an-https-server-with-a-self-signed-certificate-in-golang
	cert, err := tls.X509KeyPair(x509Cert, pkcs8PrivateKey)
	// cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %w", err)
	}
	server := &http.Server{
		Addr:    ":" + appPort,
		Handler: handler,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		},
		MaxHeaderBytes: 4096 * 4,
	}

	go func() {
		fmt.Printf("Starting App on https://localhost:%s/\n", appPort)
		if err := server.ListenAndServeTLS("", ""); err != nil {
			if http.ErrServerClosed != err { // _why_ is this an error?
				panic(err)
			}
		}
		fmt.Printf("App Shutdown\n")
	}()

	return server, nil
}
