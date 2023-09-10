package repo

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/doesnotcommit/outage_monitor/internal/outage"
	"github.com/samber/lo"
)

type DynamoWaterGovGe struct {
	waterGovGeTableName    string
	waterGovGePartitionKey string
	waterGovGeSortKey      string
	client                 *dynamodb.Client
	now                    func() time.Time
	sl                     *slog.Logger
}

func NewDynamoWaterGovGe(ctx context.Context, accessKey, secretAccessKey, region string, now func() time.Time, sl *slog.Logger) (DynamoWaterGovGe, error) {
	handleErr := func(err error) (DynamoWaterGovGe, error) {
		return DynamoWaterGovGe{}, fmt.Errorf("new dynamo water gov ge: %w", err)
	}
	creds := config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretAccessKey,
		}, nil
	}))
	regionCfg := config.WithRegion(region)
	conf, err := config.LoadDefaultConfig(ctx, creds, regionCfg)
	if err != nil {
		return handleErr(err)
	}
	conf.RetryMaxAttempts = 32
	retryMode, err := aws.ParseRetryMode("adaptive")
	if err != nil {
		return handleErr(err)
	}
	conf.RetryMode = retryMode
	client := dynamodb.NewFromConfig(conf)
	const (
		waterGovGeTableName    = "water.gov.ge"
		waterGovGePartitionKey = "locationTitle"
		waterGovGeSortKey      = "outageEnd"
	)
	return DynamoWaterGovGe{
		waterGovGeTableName,
		waterGovGePartitionKey,
		waterGovGeSortKey,
		client,
		now,
		sl,
	}, nil
}

func (w DynamoWaterGovGe) CreateTables(ctx context.Context) error {
	if _, err := w.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(w.waterGovGePartitionKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String(w.waterGovGeSortKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(w.waterGovGePartitionKey),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String(w.waterGovGeSortKey),
				KeyType:       types.KeyTypeRange,
			},
		},
		TableName:                 aws.String(w.waterGovGeTableName),
		BillingMode:               types.BillingModePayPerRequest,
		DeletionProtectionEnabled: aws.Bool(false),
	}); err != nil {
		return err
	}
	return nil
}

func (w DynamoWaterGovGe) SaveOutages(ctx context.Context, outages ...outage.WaterGovGe) error {
	handleErr := func(err error) error {
		return fmt.Errorf("save water outages: %w", err)
	}
	bwi := dynamodb.BatchWriteItemInput{
		ReturnConsumedCapacity: types.ReturnConsumedCapacityTotal,
		RequestItems:           make(map[string][]types.WriteRequest),
	}
	for _, outage := range outages {
		addressesGe := lo.Uniq(outage.AddressesGe)
		bwi.RequestItems[w.waterGovGeTableName] = append(bwi.RequestItems[w.waterGovGeTableName], types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: map[string]types.AttributeValue{
					w.waterGovGePartitionKey: &types.AttributeValueMemberS{
						Value: outage.Location.TitleLat,
					},
					w.waterGovGeSortKey: &types.AttributeValueMemberS{
						Value: outage.Start.Format(time.RFC3339),
					},
					"titleGe": &types.AttributeValueMemberS{
						Value: outage.Location.TitleGe,
					},
					"outageStart": &types.AttributeValueMemberS{
						Value: outage.Start.Format(time.RFC3339),
					},
					"affectedCustomers": &types.AttributeValueMemberN{
						Value: strconv.Itoa(outage.AffectedCustomers),
					},
					"locationLat": &types.AttributeValueMemberS{
						Value: outage.Location.Lat,
					},
					"locationLng": &types.AttributeValueMemberS{
						Value: outage.Location.Lng,
					},
					"addressesGe": &types.AttributeValueMemberSS{
						Value: addressesGe,
					},
					"locationId": &types.AttributeValueMemberS{
						Value: outage.Location.Id,
					},
				},
			},
		})
	}
	bwo, err := w.client.BatchWriteItem(ctx, &bwi)
	if err != nil {
		return handleErr(err)
	}
	w.sl.Info("saving outages", slog.Any("consumed capacity", bwo.ConsumedCapacity))
	return nil
}

func (w DynamoWaterGovGe) GetOutages(ctx context.Context, titleLat string, addressesGe ...string) ([]outage.WaterGovGe, error) {
	handleErr := func(err error) ([]outage.WaterGovGe, error) {
		return nil, fmt.Errorf("get water outages: %w", err)
	}
	now := w.now().Format(time.RFC3339)
	exp, err := expression.NewBuilder().
		WithKeyCondition(
			expression.Key(w.waterGovGePartitionKey).Equal(expression.Value(titleLat)).And(
				expression.Key(w.waterGovGeSortKey).GreaterThan(expression.Value(now)),
			),
		).
		WithProjection(expression.NamesList(
			expression.Name(w.waterGovGePartitionKey),
			expression.Name(w.waterGovGeSortKey),
			expression.Name("addressesGe"),
			expression.Name("affectedCustomers"),
			expression.Name("locationLat"),
			expression.Name("locationLng"),
			expression.Name("outageStart"),
			expression.Name("titleGe"),
			expression.Name("locationId"),
		)).
		Build()
	if err != nil {
		return handleErr(err)
	}
	qi := dynamodb.QueryInput{
		KeyConditionExpression:    exp.KeyCondition(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		ProjectionExpression:      exp.Projection(),
		TableName:                 &w.waterGovGeTableName,
		ReturnConsumedCapacity:    types.ReturnConsumedCapacityTotal,
	}
	qo, err := w.client.Query(ctx, &qi)
	if err != nil {
		return handleErr(err)
	}
	var outages []struct {
		LocationId        string
		LocationTitle     string
		OutageEnd         time.Time
		AddressesGe       []string
		AffectedCustomers int
		LocationLat       string
		LocationLng       string
		OutageStart       time.Time
		TitleGe           string
	}
	if err := attributevalue.UnmarshalListOfMaps(qo.Items, &outages); err != nil {
		return handleErr(err)
	}
	result := make([]outage.WaterGovGe, len(outages))
	for i, o := range outages {
		result[i] = outage.WaterGovGe{
			Start:             o.OutageStart,
			End:               o.OutageEnd,
			AffectedCustomers: o.AffectedCustomers,
			Location: outage.Location{
				Id:       o.LocationId,
				TitleGe:  o.TitleGe,
				TitleLat: o.LocationTitle,
				Lat:      o.LocationLat,
				Lng:      o.LocationLng,
			},
			AddressesGe: o.AddressesGe,
		}
	}
	return result, nil
}
