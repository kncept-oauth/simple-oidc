package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/development"
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
		RunAsLambda(dao.NewFilesystemDao(), "https://localhost:8443")
	case "dev":
		development.RunLocally(dao.NewFilesystemDao(), "https://localhost:8443")
	default:
		panic(fmt.Errorf("unknown run mode: %v", runmode))
	}

}

func RunAsLambda(daoSource dao.DaoSource, urlPrefix string) {
	srv, err := dispatcher.NewApplication(
		daoSource,
		urlPrefix,
	)
	if err != nil {
		log.Fatal(err)
	}
	handlerAdapter := httpadapter.NewV2(srv)
	lambda.Start(handlerAdapter.ProxyWithContext)
}
