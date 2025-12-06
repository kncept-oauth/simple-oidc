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

func (d *DynamoDbDaoSource) GetDaoSourceDescription() string {
	return "DynamoDbDaoSource"
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
			DdbEntityDetails: ddbutil.DdbEntityDetails{
				TableName:        d.tableName("auth-codes"),
				PartitionKeyName: "code",
			},
			Ddb: d.ddb,
			Supplier: func() *client.AuthorizationCode {
				return &client.AuthorizationCode{}
			},
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
		scroller.Scroll,
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
		scroller.Scroll,
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
			DdbEntityDetails: ddbutil.DdbEntityDetails{
				TableName:        d.tableName("client-authorization"),
				PartitionKeyName: "userId",
				SortKeyName:      "clientId",
			},
			Ddb: d.ddb,
			Supplier: func() *client.ClientAuthorization {
				return &client.ClientAuthorization{}
			},
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
			DdbEntityDetails: ddbutil.DdbEntityDetails{
				TableName:        d.tableName("clients"),
				PartitionKeyName: "clientId",
			},
			Ddb: d.ddb,
			Supplier: func() *client.Client {
				return &client.Client{}
			},
		},
	}
}

type DdbKeyStore struct {
	ddbutil.DdbEntityMapper[keys.JwkKeypair]
}

func (d *DynamoDbDaoSource) GetKeyStore(ctx context.Context) keys.Keystore {
	return &DdbKeyStore{
		DdbEntityMapper: ddbutil.DdbEntityMapper[keys.JwkKeypair]{
			DdbEntityDetails: ddbutil.DdbEntityDetails{
				TableName:        d.tableName("keys"),
				PartitionKeyName: "kid",
			},
			Ddb: d.ddb,
			Supplier: func() *keys.JwkKeypair {
				return &keys.JwkKeypair{}
			},
		},
	}
}

func (d *DdbKeyStore) ListKeys(ctx context.Context) ([]*keys.JwkKeypair, error) {
	return d.Scan(ctx)
}
func (d *DdbKeyStore) GetKey(ctx context.Context, kid string) (*keys.JwkKeypair, error) {
	return d.Get(ctx, kid, "")
}
func (d *DdbKeyStore) SaveKey(ctx context.Context, keypair *keys.JwkKeypair) error {
	return d.Save(ctx, keypair)
}

type DdbUserStore struct {
	ddbutil.DdbEntityMapper[users.OidcUser]
}

func (d *DdbUserStore) GetUser(ctx context.Context, id string) (*users.OidcUser, error) {
	return d.Get(ctx, id, "")
}

func (d *DdbUserStore) SaveUser(ctx context.Context, user *users.OidcUser) error {
	return d.Save(ctx, user)
}

func (d *DdbUserStore) EnumerateUsers(ctx context.Context, callback func(user *users.OidcUser) bool) error {
	return d.ScrollScan(ctx, dynamodb.ScanInput{
		TableName: &d.TableName,
	}, ddbutil.SimpleScrollCallback(callback),
	)
}

func (d *DynamoDbDaoSource) GetUserStore(ctx context.Context) users.UserStore {
	return &DdbUserStore{
		DdbEntityMapper: ddbutil.DdbEntityMapper[users.OidcUser]{
			DdbEntityDetails: ddbutil.DdbEntityDetails{
				TableName:        d.tableName("users"),
				PartitionKeyName: "id",
			},
			Ddb: d.ddb,
			Supplier: func() *users.OidcUser {
				return &users.OidcUser{}
			},
		},
	}
}

type DdbSessionStore struct {
	ddbutil.DdbEntityMapper[session.Session]
}

func (d *DdbSessionStore) ListUserSessions(ctx context.Context, userId string) ([]*session.Session, error) {
	scroller := &ddbutil.DepaginatedScroller[session.Session]{}
	d.ScrollQuery(
		ctx,
		dynamodb.QueryInput{
			TableName: &d.TableName,
			ExpressionAttributeNames: map[string]string{
				"#sk": d.SortKeyName,
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":sk": &types.AttributeValueMemberS{Value: userId},
			},
			KeyConditionExpression: aws.String("#sk = :sk"),
			IndexName:              aws.String("reverse"), // use the REVERSE index lookup
		},
		scroller.Scroll,
	)
	return scroller.Results, nil
}

func (d *DdbSessionStore) LoadSession(ctx context.Context, sessionId string, userId string) (*session.Session, error) {
	return d.Get(ctx, sessionId, userId)
}

func (d *DdbSessionStore) SaveSession(ctx context.Context, session *session.Session) error {
	return d.Save(ctx, session)
}

func (d *DynamoDbDaoSource) GetSessionStore(ctx context.Context) session.SessionStore {
	return &DdbSessionStore{
		DdbEntityMapper: ddbutil.DdbEntityMapper[session.Session]{
			DdbEntityDetails: ddbutil.DdbEntityDetails{
				TableName:        d.tableName("session-store"),
				PartitionKeyName: "id",
				SortKeyName:      "userId",
			},
			Ddb: d.ddb,
			Supplier: func() *session.Session {
				return &session.Session{}
			},
		},
	}
}
