package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kncept-oauth/simple-oidc/dispatcher"
)

func main() {
	srv, err := dispatcher.NewApplication()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("starting server on http://127.0.0.1:8080/\n")
	if err = http.ListenAndServe(":8080", srv); err != nil {
		log.Fatal(err)
	}
}
