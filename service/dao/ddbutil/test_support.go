package ddbutil

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func InitializeTable(ctx context.Context, client *dynamodb.Client, tableDetails *DdbEntityDetails) error {
	exists, err := TableExists(ctx, client, tableDetails.TableName)
	if err != nil {
		return fmt.Errorf("Unable to determine if table exists: %v", err)
	}
	if exists {
		return fmt.Errorf("Cannot create a table that already exists")
	}
	err = CreateTable(ctx, client, tableDetails)
	if err != nil {
		return fmt.Errorf("Unable to create table: %v", err)
	}
	exists, err = WaitForTable(ctx, client, tableDetails.TableName)
	if err != nil {
		return fmt.Errorf("Unable to verify if table exists: %v", err)
	}
	if !exists {
		return fmt.Errorf("Table was not created")
	}
	return nil
}
func TableExists(ctx context.Context, client *dynamodb.Client, tableName string) (bool, error) {
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		if strings.HasSuffix(err.Error(), "Cannot do operations on a non-existent table") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func CreateTable(ctx context.Context, client *dynamodb.Client, tableDetails *DdbEntityDetails) error {
	keySchema := []types.KeySchemaElement{
		{
			AttributeName: &tableDetails.PartitionKeyName,
			KeyType:       types.KeyTypeHash,
		},
	}
	globalSecondaryIndexes := []types.GlobalSecondaryIndex{}

	if tableDetails.SortKeyName != "" {
		keySchema = append(keySchema, types.KeySchemaElement{
			AttributeName: &tableDetails.SortKeyName,
			KeyType:       types.KeyTypeRange,
		})
		// create a reverse index if there is a pk AND an sk
		globalSecondaryIndexes = append(globalSecondaryIndexes, types.GlobalSecondaryIndex{
			IndexName: aws.String("reverse"),
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: &tableDetails.SortKeyName,
					KeyType:       types.KeyTypeHash,
				},
				{
					AttributeName: &tableDetails.PartitionKeyName,
					KeyType:       types.KeyTypeRange,
				},
			},
			Projection: &types.Projection{
				ProjectionType: types.ProjectionTypeAll,
			},
		})
	}

	attributeDefinitions := []types.AttributeDefinition{}
	for _, keyElement := range keySchema {
		attributeDefinitions = append(attributeDefinitions, types.AttributeDefinition{
			AttributeName: keyElement.AttributeName,
			AttributeType: types.ScalarAttributeTypeS,
		})
	}

	// mmm... aws seems to care if this is nil or not :/
	if len(globalSecondaryIndexes) == 0 {
		globalSecondaryIndexes = nil
	}
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:              &tableDetails.TableName,
		BillingMode:            types.BillingModePayPerRequest,
		KeySchema:              keySchema,
		AttributeDefinitions:   attributeDefinitions,
		GlobalSecondaryIndexes: globalSecondaryIndexes,
	})
	return err
}
func WaitForTable(ctx context.Context, client *dynamodb.Client, tableName string) (exists bool, err error) {
	endTime := time.Now().Add(10 * time.Second)
	for !exists && time.Now().Before(endTime) {
		time.Sleep(1 * time.Second)
		exists, err = TableExists(ctx, client, tableName)
		if err != nil {
			return
		}
	}
	return
}
