package main

import (
	"context"
	"fmt"
	"net/http"
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

	// fmt.Printf("Runmode: %s\nHostUrl: %s\n", runmode, hostUrl)

	switch runmode {
	case "aws-lambda":
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			panic(err)
		}
		daoSource := dao.NewDynamoDbDao(cfg, "")
		err = wrappedRunner(daoSource, hostUrl, func(handler http.Handler) error {
			handlerAdapter := httpadapter.New(handler)
			lambda.Start(handlerAdapter.ProxyWithContext)
			return nil
		})
		if err != nil {
			panic(err)
		}
	case "ecs-ssl":
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			panic(err)
		}
		daoSource := dao.NewDynamoDbDao(cfg, "")
		err = wrappedRunner(daoSource, hostUrl, func(handler http.Handler) error {
			_, err := development.RunLocally(daoSource, handler)
			return err
		})
		if err != nil {
			panic(err)
		}
	case "dev":
		daoSource := dao.NewFilesystemDao()
		err := wrappedRunner(daoSource, hostUrl, func(handler http.Handler) error {
			_, err := development.RunLocally(daoSource, handler)
			return err
		})
		if err != nil {
			panic(err)
		}
	default:
		panic(fmt.Errorf("unknown run mode: %v", runmode))
	}
}

func wrappedRunner(daoSource dao.DaoSource, hostUrl string, callback func(handler http.Handler) error) error {
	srv, err := dispatcher.NewApplication(
		daoSource,
		hostUrl,
	)
	if err != nil {
		return err
	}
	return callback(srv)
}
