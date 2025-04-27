package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	servicedao "github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/development"
	"github.com/kncept-oauth/simple-oidc/testharness/dispatcher"
)

func main() {
	var wg sync.WaitGroup

	datastore := servicedao.NewMemoryDao()

	// run a the application, with access to the underlying datastore
	// appPort := "8080"
	// handler, err := servicedispatcher.NewApplication(datastore)
	// if err != nil {
	// panic(err)
	// }
	// server := http.Server{Addr: ":" + appPort, Handler: handler}

	appServer, err := development.RunLocally(datastore, "https://localhost:8443")
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
	app := dispatcher.NewApplication(datastore)
	testHarnessPort := "3000"

	go func() {
		fmt.Printf("Starting Testharness on http://localhost:%s/\n", testHarnessPort)
		if err := app.Listen(":" + testHarnessPort); err != nil {
			panic(err)
		}
		fmt.Printf("Testharness Shutdown\n")
	}()

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
