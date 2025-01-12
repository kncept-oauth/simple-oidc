package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
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
	var wg sync.WaitGroup

	handler, err := dispatcher.NewApplication(
		dao.NewFilesystemDao(),
	)
	if err != nil {
		panic(err)
	}
	appPort := "8080"
	server := http.Server{Addr: ":" + appPort, Handler: handler}
	wg.Add(1)
	go func() {
		fmt.Printf("Starting App on http://127.0.0.1:%s/\n", appPort)
		if err := server.ListenAndServe(); err != nil {
			if http.ErrServerClosed != err { // _why_ is this an error?
				panic(err)
			}
		}
		fmt.Printf("App Shutdown\n")
		wg.Done()
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdownChan
		fmt.Printf("Shutting down\n")
		server.Shutdown(context.Background())
	}()
	wg.Wait()
}
