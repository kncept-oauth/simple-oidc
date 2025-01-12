package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	servicedao "github.com/kncept-oauth/simple-oidc/service/dao"
	servicedispatcher "github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/testharness/dispatcher"
)

func main() {
	var wg sync.WaitGroup

	datastore := servicedao.NewMemoryDao()

	// run a the application, with access to the underlying datastore
	appPort := "8080"
	handler, err := servicedispatcher.NewApplication(datastore)
	if err != nil {
		panic(err)
	}
	server := http.Server{Addr: ":" + appPort, Handler: handler}
	wg.Add(1)
	go func() {
		fmt.Printf("Starting Nested App on http://127.0.0.1:%s/\n", appPort)
		if err := server.ListenAndServe(); err != nil {
			if http.ErrServerClosed != err { // _why_ is this an error?
				panic(err)
			}
		}
		fmt.Printf("Nested App Shutdown\n")
		wg.Done()
	}()

	time.Sleep(1 * time.Second)
	// run a test harness
	app := dispatcher.NewApplication(datastore)
	testHarnessPort := "3000"
	wg.Add(1)
	go func() {

		fmt.Printf("Starting Testharness on http://127.0.0.1:%s/\n", testHarnessPort)
		if err := app.Listen(":" + testHarnessPort); err != nil {
			panic(err)
		}
		fmt.Printf("Testharness Shutdown\n")
		wg.Done()
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdownChan
		fmt.Printf("Shutting down\n")
		app.Shutdown()
		server.Shutdown(context.Background())
	}()

	wg.Wait()
}
