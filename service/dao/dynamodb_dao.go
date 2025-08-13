package dao

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/kncept-oauth/simple-oidc/service/client"
	"github.com/kncept-oauth/simple-oidc/service/dao/ddbutil"
	"github.com/kncept-oauth/simple-oidc/service/keys"
	"github.com/kncept-oauth/simple-oidc/service/session"
	"github.com/kncept-oauth/simple-oidc/service/users"
)

type DynamoDbDaoSource struct {
	ddb         *dynamodb.Client
	tablePrefix string
}

func NewDynamoDbDao(cfg aws.Config, tablePrefix string) DaoSource {
	// Using the Config value, create the DynamoDB client
	ddb := dynamodb.NewFromConfig(cfg)
	return &DynamoDbDaoSource{
		ddb:         ddb,
		tablePrefix: tablePrefix,
	}
}

// called to initialize the DdbEntityMapper
func (d *DynamoDbDaoSource) tableName(name string) string {
	if d.tablePrefix == "" {
		return name
	}
	suffixes := []string{
		"_", "-", ".",
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(d.tablePrefix, suffix) {
			return d.tablePrefix + suffix
		}
	}
	return d.tablePrefix + "-" + name
}

type DdbAuthorizationCodeStore struct {
	ddbutil.DdbEntityMapper[client.AuthorizationCode]
}

func (d *DdbAuthorizationCodeStore) GetAuthorizationCode(ctx context.Context, code string) (*client.AuthorizationCode, error) {
	return d.Get(ctx, code, "")
}

func (d *DdbAuthorizationCodeStore) SaveAuthorizationCode(ctx context.Context, code *client.AuthorizationCode) error {
	return d.Save(ctx, code)
}

func (d *DynamoDbDaoSource) GetAuthorizationCodeStore(ctx context.Context) client.AuthorizationCodeStore {
	return &DdbAuthorizationCodeStore{
		DdbEntityMapper: ddbutil.DdbEntityMapper[client.AuthorizationCode]{
			Ddb:       d.ddb,
			TableName: d.tableName("auth-codes"),
			Supplier: func() *client.AuthorizationCode {
				return &client.AuthorizationCode{}
			},
			PartitionKeyName: "code",
		},
	}
}

type DdbClientAuthorizationStore struct {
	ddbutil.DdbEntityMapper[client.ClientAuthorization]
}

func (c *DdbClientAuthorizationStore) ClientAuthorizationsByClient(ctx context.Context, clientId string, scroller ddbutil.SimpleScroller[client.ClientAuthorization]) error {
	return c.ScrollQuery(
		ctx,
		dynamodb.QueryInput{
			TableName: &c.TableName,
			ExpressionAttributeNames: map[string]string{
				"#sk": c.SortKeyName,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":sk": &types.AttributeValueMemberS{Value: clientId},
			},
			KeyConditionExpression: aws.String("#sk = :sk"),
			IndexName:              aws.String("reverse"), // use the REVERSE index lookup
		},
		scroller,
	)
}

func (c *DdbClientAuthorizationStore) ClientAuthorizationsByUser(ctx context.Context, userId string, scroller ddbutil.SimpleScroller[client.ClientAuthorization]) error {
	return c.ScrollQuery(
		ctx,
		dynamodb.QueryInput{
			TableName: &c.TableName,
			ExpressionAttributeNames: map[string]string{
				"#pk": c.PartitionKeyName,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: userId},
			},
			KeyConditionExpression: aws.String("#pk = :pk"),
		},
		scroller,
	)
}

func (c *DdbClientAuthorizationStore) DeleteClientAuthorization(ctx context.Context, userId string, clientId string) error {
	return c.DeleteById(ctx, userId, clientId)
}

func (c *DdbClientAuthorizationStore) GetClientAuthorization(ctx context.Context, userId string, clientId string) (*client.ClientAuthorization, error) {
	return c.Get(ctx, userId, clientId)
}

func (c *DdbClientAuthorizationStore) SaveClientAuthorization(ctx context.Context, clientAuthorization *client.ClientAuthorization) error {
	return c.Save(ctx, clientAuthorization)
}

func (d *DynamoDbDaoSource) GetClientAuthorizationStore(ctx context.Context) client.ClientAuthorizationStore {
	return &DdbClientAuthorizationStore{
		DdbEntityMapper: ddbutil.DdbEntityMapper[client.ClientAuthorization]{
			Ddb:       d.ddb,
			TableName: d.tableName("client-authorization"),
			Supplier: func() *client.ClientAuthorization {
				return &client.ClientAuthorization{}
			},
			PartitionKeyName: "userId",
			SortKeyName:      "clientId",
		},
	}
}

type DdbClientStore struct {
	ddbutil.DdbEntityMapper[client.Client]
}

func (d *DdbClientStore) GetClient(ctx context.Context, clientId string) (*client.Client, error) {
	return d.Get(ctx, clientId, "")
}

func (d *DdbClientStore) ListClients(ctx context.Context) ([]*client.Client, error) {
	return d.Scan(ctx)
}

func (d *DdbClientStore) RemoveClient(ctx context.Context, clientId string) error {
	return d.DeleteById(ctx, clientId, "")
}

func (d *DdbClientStore) SaveClient(ctx context.Context, client *client.Client) error {
	return d.Save(ctx, client)
}

func (d *DynamoDbDaoSource) GetClientStore(ctx context.Context) client.ClientStore {
	return &DdbClientStore{
		DdbEntityMapper: ddbutil.DdbEntityMapper[client.Client]{
			Ddb:       d.ddb,
			TableName: d.tableName("clients"),
			Supplier: func() *client.Client {
				return &client.Client{}
			},
			PartitionKeyName: "clientId",
		},
	}
}

func (d *DynamoDbDaoSource) GetKeyStore(ctx context.Context) keys.Keystore {
	panic("unimplemented")
}

func (d *DynamoDbDaoSource) GetSessionStore(ctx context.Context) session.SessionStore {
	panic("unimplemented")
}

func (d *DynamoDbDaoSource) GetUserStore(ctx context.Context) users.UserStore {
	panic("unimplemented")
}
