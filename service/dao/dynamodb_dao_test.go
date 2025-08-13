package dao

// x go:build integration
// x +build integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao/ddbutil"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"
)

var AwsCfg *aws.Config

func RunWrappedMain(m *testing.M) {
	ctx := context.Background()

	awsRegion := "ap-southeast-2"

	// https://golang.testcontainers.org/modules/localstack/
	localstackContainer, err := localstack.Run(
		ctx,
		"localstack/localstack:4.7", //"localstack/localstack:1.4.0",
		testcontainers.WithEnv(map[string]string{
			"SERVICES":   "dynamodb",
			"AWS_REGION": awsRegion,
		}),
		testcontainers.WithWaitStrategy(wait.ForHealthCheck()),
	)

	// defer func() {
	// 	if err := testcontainers.TerminateContainer(localstackContainer); err != nil {
	// 		log.Printf("failed to terminate container: %s", err)
	// 	}
	// }()
	if err != nil {
		panic(fmt.Sprintf("failed to start container: %s", err))
	}

	// inspectRes, err := localstackContainer.Inspect(ctx)
	// if err != nil {
	// 	t.Fatalf("unable to inspect container: %s", err)
	// }
	// portMap := inspectRes.NetworkSettings.Ports
	// fmt.Printf("PortMap:\n%v\n", portMap)

	mappedPort, err := localstackContainer.MappedPort(ctx, "4566/tcp")
	if err != nil {
		panic(fmt.Sprintf("failed to get mapped port: %s", err))
	}
	awsEndpoint := fmt.Sprintf("http://localhost:%s", mappedPort.Port())

	// mappedPort, err := localstackContainer.Endpoint(ctx, "5466/tcp")
	// awsEndpoint := fmt.Sprintf("http://localhost:%s", mappedPort)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("id", "secret", "session")),
		config.WithBaseEndpoint(awsEndpoint),
		// config.WithCredentialsCacheOptions()
	)
	if err != nil {
		panic(fmt.Sprintf("failed to get localstack aws config: %s", err))
	}

	cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			// fmt.Printf("EndpointResolverWithOptionsFunc %v\n", awsEndpoint)
			return aws.Endpoint{
				URL:           awsEndpoint,
				SigningRegion: awsRegion,
			}, nil
		})
	cfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		// fmt.Printf("EndpointResolverFunc %v\n", awsEndpoint)
		return aws.Endpoint{
			URL:           awsEndpoint,
			SigningRegion: awsRegion,
		}, nil
	})
	// cfg.BaseEndpoint = &awsEndpoint

	AwsCfg = &cfg
	InitAllTables(ctx, cfg)

	exitCode := m.Run()

	if err := testcontainers.TerminateContainer(localstackContainer); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}
	os.Exit(exitCode)
}

func InitAllTables(ctx context.Context, cfg aws.Config) {
	// init the tables now
	ddb := dynamodb.NewFromConfig(cfg)
	dao := &DynamoDbDaoSource{
		ddb:         ddb,
		tablePrefix: "",
	}

	tablesNames := []string{}

	if obj, ok := dao.GetAuthorizationCodeStore(ctx).(*DdbAuthorizationCodeStore); ok {
		if err := ddbutil.InitializeTable(ctx, ddb, &obj.DdbEntityMapper); err != nil {
			panic(err)
		}
		tablesNames = append(tablesNames, obj.DdbEntityMapper.TableName)
	}
	if obj, ok := dao.GetClientStore(ctx).(*DdbClientStore); ok {
		if err := ddbutil.InitializeTable(ctx, ddb, &obj.DdbEntityMapper); err != nil {
			panic(err)
		}
		tablesNames = append(tablesNames, obj.DdbEntityMapper.TableName)
	}
	if obj, ok := dao.GetClientAuthorizationStore(ctx).(*DdbClientAuthorizationStore); ok {
		if err := ddbutil.InitializeTable(ctx, ddb, &obj.DdbEntityMapper); err != nil {
			panic(err)
		}
		tablesNames = append(tablesNames, obj.DdbEntityMapper.TableName)
	}
}

func TestMain(m *testing.M) {
	RunWrappedMain(m)
}

func TestClientStore(t *testing.T) {
	cfg := *AwsCfg
	ctx := t.Context()
	dao := NewDynamoDbDao(cfg, "")
	clientStore := dao.GetClientStore(ctx)

	foundClient, err := clientStore.GetClient(ctx, "does not exists")
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}
	if foundClient != nil {
		t.Fatalf("unexpectedly found a client")
	}

	newClient := &client.Client{
		ClientId: uuid.NewString(),
	}
	err = clientStore.SaveClient(ctx, newClient)
	if err != nil {
		t.Fatalf("SaveClient error: %v", err)
	}

	foundClient, err = clientStore.GetClient(ctx, newClient.ClientId)
	if err != nil {
		t.Fatalf("GetClient error: %v", err)
	}
	if foundClient == nil {
		t.Fatalf("Didn't find the client")
	}
	if foundClient.ClientId != newClient.ClientId {
		fmt.Printf("Expected %v but got %v for the client id", newClient.ClientId, foundClient.ClientId)
	}

	listedClients, err := clientStore.ListClients(ctx)
	if err != nil {
		t.Fatalf("ListClients error")
	}
	fmt.Printf("listedClients len %v\n", len(listedClients))
	for _, listedClient := range listedClients {
		fmt.Printf("client %+v\n", listedClient)
		if listedClient.ClientId == newClient.ClientId {
			return //end test
		}
	}
	t.Fatalf("Unable to find newly persisted client in list")
}

func TestClientAuthorizationStore(t *testing.T) {
	cfg := *AwsCfg
	ctx := t.Context()
	dao := NewDynamoDbDao(cfg, "")
	clientAuths := dao.GetClientAuthorizationStore(ctx)

	userId1 := uuid.NewString()
	userId2 := uuid.NewString()
	clientId1 := uuid.NewString()
	clientId2 := uuid.NewString()

	foundAuth, err := clientAuths.GetClientAuthorization(ctx, userId1, clientId1)
	if err != nil {
		t.Fatalf("GetClientAuthorization error: %v", err)
	}
	if foundAuth != nil {
		t.Fatalf("Unexpectedly found a clientauth")
	}

	scroller := &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
	err = clientAuths.ClientAuthorizationsByUser(ctx, userId1, scroller)
	if err != nil {
		t.Fatalf("ClientAuthorizationsByUser: %v", err)
	}
	if len(scroller.Results) != 0 {
		t.Fatalf("ClientAuthorizationsByUser found results but expected none")
	}
	scroller = &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
	err = clientAuths.ClientAuthorizationsByClient(ctx, clientId1, scroller)
	if err != nil {
		t.Fatalf("ClientAuthorizationsByClient: %v", err)
	}
	if len(scroller.Results) != 0 {
		t.Fatalf("ClientAuthorizationsByClient found results but expected none")
	}

	now := time.Now()
	if err = clientAuths.SaveClientAuthorization(ctx, &client.ClientAuthorization{
		UserId:          userId1,
		ClientId:        clientId1,
		AuthorizedAt:    now,
		LastRefreshedAt: now,
	}); err != nil {
		t.Fatalf("Failed to save client authorization")
	}
	if err = clientAuths.SaveClientAuthorization(ctx, &client.ClientAuthorization{
		UserId:          userId1,
		ClientId:        clientId2,
		AuthorizedAt:    now,
		LastRefreshedAt: now,
	}); err != nil {
		t.Fatalf("Failed to save client authorization")
	}
	if err = clientAuths.SaveClientAuthorization(ctx, &client.ClientAuthorization{
		UserId:          userId2,
		ClientId:        clientId1,
		AuthorizedAt:    now,
		LastRefreshedAt: now,
	}); err != nil {
		t.Fatalf("Failed to save client authorization")
	}
	if err = clientAuths.SaveClientAuthorization(ctx, &client.ClientAuthorization{
		UserId:          userId2,
		ClientId:        clientId2,
		AuthorizedAt:    now,
		LastRefreshedAt: now,
	}); err != nil {
		t.Fatalf("Failed to save client authorization")
	}

	foundAuth, err = clientAuths.GetClientAuthorization(ctx, userId1, clientId1)
	if err != nil {
		t.Fatalf("GetClientAuthorization error: %v", err)
	}
	if foundAuth == nil {
		t.Fatalf("should have found a client authorization")
	}
	if foundAuth.UserId != userId1 {
		t.Fatalf("UserId mismatch")
	}
	if foundAuth.ClientId != clientId1 {
		t.Fatalf("ClientId mismatch")
	}

	scroller = &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
	err = clientAuths.ClientAuthorizationsByUser(ctx, userId1, scroller)
	if err != nil {
		t.Fatalf("ClientAuthorizationsByUser: %v", err)
	}
	if len(scroller.Results) != 2 {
		t.Fatalf("ClientAuthorizationsByUser found %v results but expected 2", len(scroller.Results))
	}
	scroller = &ddbutil.DepaginatedScroller[client.ClientAuthorization]{}
	err = clientAuths.ClientAuthorizationsByClient(ctx, clientId1, scroller)
	if err != nil {
		t.Fatalf("ClientAuthorizationsByClient: %v", err)
	}
	if len(scroller.Results) != 2 {
		t.Fatalf("ClientAuthorizationsByClient found %v results but expected 2", len(scroller.Results))
	}

	err = clientAuths.DeleteClientAuthorization(ctx, userId1, clientId1)
	if err != nil {
		t.Fatalf("GetClientAuthorization error: %v", err)
	}
	foundAuth, err = clientAuths.GetClientAuthorization(ctx, userId1, clientId1)
	if err != nil {
		t.Fatalf("GetClientAuthorization error: %v", err)
	}
	if foundAuth != nil {
		t.Fatalf("Unexpectedly found a clientauth")
	}
}

func TestAuthorizationCodeStore(t *testing.T) {
	cfg := *AwsCfg
	ctx := t.Context()
	dao := NewDynamoDbDao(cfg, "")
	authCodes := dao.GetAuthorizationCodeStore(ctx)

	code, err := authCodes.GetAuthorizationCode(ctx, "does not exist")
	if err != nil {
		t.Fatalf("GetAuthorizationCode failed: %s", err)
	}
	if code != nil {
		t.Fatalf("Found a code when it shouldn't have: %+v", code)
	}

	newAuthCode := client.AuthorizationCode{
		Code:       uuid.NewString(),
		UserId:     uuid.NewString(),
		OidcParams: "",
	}

	err = authCodes.SaveAuthorizationCode(ctx, &newAuthCode)
	if err != nil {
		t.Fatalf("SaveAuthorizationCode failed: %v", err)
	}
	code, err = authCodes.GetAuthorizationCode(ctx, newAuthCode.Code)
	if err != nil {
		t.Fatalf("GetAuthorizationCode failed: %s", err)
	}
	if code == nil {
		t.Fatalf("failed to find a code")
	}
	if code.Code != newAuthCode.Code {
		t.Fatalf("Expected %v but got %v as the code", newAuthCode.Code, code.Code)
	}
}
