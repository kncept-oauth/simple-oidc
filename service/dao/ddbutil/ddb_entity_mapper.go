package ddbutil

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func SimpleScrollCallback[T any](callback func(item *T) bool) ScrollCallback[T] {
	return func(items []*T) bool {
		processMore := true
		i := 0
		for processMore && i < len(items) {
			processMore = callback(items[i])
			i++
		}
		return processMore
	}
}

type ScrollCallback[T any] func(items []*T) (processMore bool)

type SimpleScrollerFunc[T any] func(items []*T) (processMore bool)
type SimpleScroller[T any] interface {
	Scroll(items []*T) (processMore bool)
}

func (fn SimpleScrollerFunc[T]) Scroll(items []*T) bool {
	return fn(items)
}

type DepaginatedScroller[T any] struct {
	Results []*T
}

func (s *DepaginatedScroller[T]) Scroll(items []*T) bool {
	if s.Results == nil {
		s.Results = make([]*T, 0)
	}
	s.Results = append(s.Results, items...)
	return true
}

type DdbEntityDetails struct {
	TableName        string `json:"tableName"`
	PartitionKeyName string `json:"partitionKeyName"`
	SortKeyName      string `json:"sortKeyName,omitempty"`
}

// I like this pattern.
type DdbEntityMapper[T any] struct {
	DdbEntityDetails
	Ddb      *dynamodb.Client `json:"-"`
	Supplier func() *T        `json:"-"`
}

func (d *DdbEntityMapper[T]) Get(ctx context.Context, partitionKey string, sortKey string) (*T, error) {
	key, err := d.key(partitionKey, sortKey)
	if err != nil {
		return nil, err
	}
	res, err := d.Ddb.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &d.TableName,
		Key:       key,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Item) == 0 {
		return nil, nil
	}

	val := d.Supplier()
	err = attributevalue.UnmarshalMap(res.Item, val)
	if err != nil {
		return nil, err
	}
	return val, err
}

func (d *DdbEntityMapper[T]) Save(ctx context.Context, value *T) error {
	mapValue, err := attributevalue.MarshalMap(value)
	if err != nil {
		return err
	}
	_, err = d.Ddb.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item:      mapValue,
	})
	return err
}

func (d *DdbEntityMapper[T]) DeleteById(ctx context.Context, partitionKey string, sortKey string) error {
	key, err := d.key(partitionKey, sortKey)
	if err != nil {
		return err
	}
	_, err = d.Ddb.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &d.TableName,
		Key:       key,
	})
	return err
}

// this can scan the _entire_ table.
// a different approach should be used.
func (d *DdbEntityMapper[T]) Scan(ctx context.Context) ([]*T, error) {
	allResults := &DepaginatedScroller[T]{}
	err := d.ScrollScan(ctx, dynamodb.ScanInput{
		TableName: &d.TableName,
	}, allResults.Scroll)
	return allResults.Results, err
}

// input is a COPY because the value is modified
func (d *DdbEntityMapper[T]) ScrollScan(ctx context.Context, input dynamodb.ScanInput, callback ScrollCallback[T]) error {
	processMore := false
	scanRes, err := d.Ddb.Scan(ctx, &input)
	if err != nil {
		return err
	}
	if len(scanRes.Items) != 0 {
		items := make([]*T, 0)
		for _, item := range scanRes.Items {
			val := d.Supplier()
			err = attributevalue.UnmarshalMap(item, val)
			if err != nil {
				return err
			}
			items = append(items, val)
		}
		processMore = callback(items)
	}
	for processMore && scanRes.LastEvaluatedKey != nil {
		scanRes, err = d.Ddb.Scan(ctx, &input)
		if err != nil {
			return err
		}

		items := make([]*T, 0)
		for _, item := range scanRes.Items {
			val := d.Supplier()
			err = attributevalue.UnmarshalMap(item, val)
			if err != nil {
				return err
			}
			items = append(items, val)
		}
		input.ExclusiveStartKey = scanRes.LastEvaluatedKey
		processMore = callback(items)
	}
	return nil
}

func (d *DdbEntityMapper[T]) ScrollQuery(ctx context.Context, input dynamodb.QueryInput, callback ScrollCallback[T]) error {
	processMore := false
	scanRes, err := d.Ddb.Query(ctx, &input)
	if err != nil {
		return err
	}
	if len(scanRes.Items) != 0 {
		items := make([]*T, 0)
		for _, item := range scanRes.Items {
			val := d.Supplier()
			err = attributevalue.UnmarshalMap(item, val)
			if err != nil {
				return err
			}
			items = append(items, val)
		}
		processMore = callback(items)
	}
	for processMore && scanRes.LastEvaluatedKey != nil {
		scanRes, err = d.Ddb.Query(ctx, &input)
		if err != nil {
			return err
		}

		items := make([]*T, 0)
		for _, item := range scanRes.Items {
			val := d.Supplier()
			err = attributevalue.UnmarshalMap(item, val)
			if err != nil {
				return err
			}
			items = append(items, val)
		}
		input.ExclusiveStartKey = scanRes.LastEvaluatedKey
		processMore = callback(items)
	}
	return nil
}

func (d *DdbEntityMapper[T]) key(partitionKey string, sortKey string) (map[string]types.AttributeValue, error) {
	if d.SortKeyName == "" && sortKey == "" {
		return map[string]types.AttributeValue{
			d.PartitionKeyName: &types.AttributeValueMemberS{Value: partitionKey},
		}, nil
	}
	if d.SortKeyName != "" && sortKey != "" {
		return map[string]types.AttributeValue{
			d.PartitionKeyName: &types.AttributeValueMemberS{Value: partitionKey},
			d.SortKeyName:      &types.AttributeValueMemberS{Value: sortKey},
		}, nil
	}
	return nil, fmt.Errorf("sort key incorrectly defined")
}
