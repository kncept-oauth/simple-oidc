package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/kncept-oauth/simple-oidc/dao"
	"github.com/kncept-oauth/simple-oidc/dispatcher"
)

func main() {
	runmode := os.Getenv("RUN_MODE")
	if runmode == "" {
		lambdaFunctionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
		if lambdaFunctionName != "" {
			runmode = "aws-lambda"
		} else {
			runmode = "dev"
		}
	}

	switch runmode {
	case "aws-lambda":
		RunAsLambda()
	case "dev":
		RunLocally()
	default:
		panic(fmt.Errorf("unknown run mode: %v", runmode))
	}

}

func RunAsLambda() {
	srv, err := dispatcher.NewApplication(dao.NewDynamoDbDao())
	if err != nil {
		log.Fatal(err)
	}
	handlerAdapter := httpadapter.NewV2(srv)
	lambda.Start(handlerAdapter.ProxyWithContext)
}

func RunLocally() {
	srv, err := dispatcher.NewApplication(
		dao.NewFilesystemDao(),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("starting server on http://127.0.0.1:8080/\n")
	if err := http.ListenAndServe(":8080", srv); err != nil {
		log.Fatal(err)
	}
}
