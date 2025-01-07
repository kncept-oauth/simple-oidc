package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	servicedao "github.com/kncept-oauth/simple-oidc/service/dao"
	servicedispatcher "github.com/kncept-oauth/simple-oidc/service/dispatcher"
	"github.com/kncept-oauth/simple-oidc/testharness/dispatcher"
)

func main() {

	var wg sync.WaitGroup

	datastore := servicedao.NewFilesystemDao()

	wg.Add(1)
	go func() {
		srv, err := servicedispatcher.NewApplication(datastore)
		if err != nil {
			panic(err)
		}
		fmt.Printf("starting nested app on http://127.0.0.1:8080/\n")
		if err := http.ListenAndServe(":8080", srv); err != nil {
			wg.Done()
			log.Fatal(err)
		}
	}()

	wg.Add(1)
	go func() {
		// run a test harness on :3000
		srv, err := dispatcher.NewApplication(datastore)
		if err != nil {
			panic(err)
		}
		fmt.Printf("starting testharness on http://127.0.0.1:3000/\n")
		if err := http.ListenAndServe(":3000", srv); err != nil {
			wg.Done()
			log.Fatal(err)
		}
		wg.Done()
	}()

	wg.Wait()
}
