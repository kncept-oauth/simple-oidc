package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/kncept-oauth/simple-oidc/service/dao"
	"github.com/kncept-oauth/simple-oidc/service/development"
	"github.com/kncept-oauth/simple-oidc/service/dispatcher"
)

func main() {
	ctx := context.Background()
	runmode := os.Getenv("RUN_MODE")
	if runmode == "" {
		lambdaFunctionName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
		if lambdaFunctionName != "" {
			runmode = "aws-lambda"
		} else {
			runmode = "dev"
		}
	}

	hostUrl := "https://localhost:8443"
	hostName := os.Getenv("host_name")
	if hostName != "" {
		hostUrl = fmt.Sprintf("https://%s", hostName)
	}

	switch runmode {
	case "aws-lambda":
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			panic(err)
		}

		RunAsLambda(dao.NewDynamoDbDao(cfg, ""), hostUrl)
	case "dev":
		development.RunLocally(dao.NewFilesystemDao(), hostUrl)
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
