package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/kncept-oauth/simple-oidc/dispatcher"
)

func main() {
	srv, err := dispatcher.NewApplication()
	if err != nil {
		log.Fatal(err)
	}
	handlerAdapter := httpadapter.NewV2(srv)
	lambda.Start(handlerAdapter.ProxyWithContext)
}
