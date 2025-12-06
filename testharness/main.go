package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/kncept-oauth/simple-oidc/service/crypto"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	servicedao "github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/development"
	servicedispatcher "github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/testharness/dispatcher"
)

func main() {
	var wg sync.WaitGroup

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// datastore := servicedao.NewMemoryDao()

	daoSource := servicedao.NewFilesystemDao()
	srv, err := servicedispatcher.NewApplication(
		daoSource,
		"https://localhost:8443",
	)
	if err != nil {
		panic(err)
	}

	appServer, err := development.RunLocally(daoSource, srv)
	if err != nil {
		panic(err)
	}

	// var app *fiber.App
	app, err := RunAppAsHttps(daoSource)
	if err != nil {
		panic(err)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		<-shutdownChan
		wg.Done()
		fmt.Printf("Shutting down\n")
		app.Shutdown()
		appServer.Shutdown(context.Background())
	}()

	wg.Wait()
}

// runs 'ListenAndServeTLS' in a goroutine
func RunAppAsHttps(daoSource dao.DaoSource) (*fiber.App, error) {

	app := dispatcher.NewApplication(daoSource)
	testHarnessPort := "3000"

	// TLS Certificates: generate if absent
	generateCerts := false
	var pkcs8PrivateKey []byte
	x509Cert, err := os.ReadFile("../testharness/server.crt")
	if err != nil || len(x509Cert) == 0 {
		generateCerts = true
	}

	if !generateCerts {
		pkcs8PrivateKey, err = os.ReadFile("../testharness/server.key")
		generateCerts = err != nil || len(pkcs8PrivateKey) == 0
	}
	if generateCerts {
		x509Cert, pkcs8PrivateKey, err = crypto.GenerateTlsCertificate("localhost", crypto.RSA2048)
		if err == nil {
			err = os.WriteFile("../testharness/server.crt", x509Cert, 0644)
		}
		if err == nil {
			err = os.WriteFile("../testharness/server.key", pkcs8PrivateKey, 0644)
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

	go func() {
		fmt.Printf("Starting Testharness on https://localhost:%s/\n", testHarnessPort)
		if err := app.ListenTLSWithCertificate(":"+testHarnessPort, cert); err != nil {
			panic(err)
		}
		fmt.Printf("Testharness Shutdown\n")
	}()

	return app, nil
}
