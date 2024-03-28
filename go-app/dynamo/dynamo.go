package dynamo

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/trace"
)

const (
	DEFAULT_DYNAMO_HOST = "localhost"
	DEFAULT_DYNAMO_PORT = "8009"
)

var (
	once     sync.Once
	instance *DynamoDBClient
)

type DynamoDBClient struct {
	client *dynamodb.Client
	table  string
}

func buildLocalConfiguration() (cfg aws.Config, err error) {
	// Define endpoint resolver for local DynamoDB use
	host := utils.GetEnv("DYNAMO_HOST", DEFAULT_DYNAMO_HOST)
	port := utils.GetEnv("DYNAMO_PORT", DEFAULT_DYNAMO_PORT)
	hostAddress := fmt.Sprintf("http://%s", net.JoinHostPort(host, port))

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: hostAddress, SigningRegion: "localhost",
		}, nil
	})

	cfg, err = config.LoadDefaultConfig(context.Background(),
		config.WithRegion("localhost"),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     "local",
				SecretAccessKey: "local",
			},
		}),
	)

	return cfg, err
}

func NewDynamoDBClient(ctx context.Context, table string) (*DynamoDBClient, error) {
	var err error
	span := trace.SpanFromContext(ctx)
	once.Do(func() {
		span.AddEvent("First initialization of the DynamoDB client")
		cfg, loadErr := buildLocalConfiguration()

		otelaws.AppendMiddlewares(&cfg.APIOptions)

		if loadErr != nil {
			err = fmt.Errorf("unable to load SDK config: %v", loadErr)
			return
		}

		client := dynamodb.NewFromConfig(cfg)

		instance = &DynamoDBClient{
			client: client,
			table:  table,
		}
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (c *DynamoDBClient) PutItem(item map[string]types.AttributeValue) error {
	input := &dynamodb.PutItemInput{
		TableName: &c.table,
		Item:      item,
	}

	_, err := c.client.PutItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to put item: %v", err)
	}

	return nil
}

func (c *DynamoDBClient) ListItems(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("SELECT %s FROM %q WHERE %s = '%s'",
		"ID", "Services",
		"ID", "1")
	_, err := c.client.ExecuteStatement(ctx, &dynamodb.ExecuteStatementInput{
		Statement: aws.String(query),
	})
	if err != nil {
		return nil, err
	}

	return []string{}, nil
}
