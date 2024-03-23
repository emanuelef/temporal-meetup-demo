package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
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

func GetEnv(envVar, defaultValue string) string {
	val, exists := os.LookupEnv(envVar)
	if !exists {
		return defaultValue
	}
	return val
}

func buildLocalConfiguration() (cfg aws.Config, err error) {
	// Define endpoint resolver for local DynamoDB use
	host := GetEnv("DYNAMO_HOST", DEFAULT_DYNAMO_HOST)
	port := GetEnv("DYNAMO_PORT", DEFAULT_DYNAMO_PORT)
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

func NewDynamoDBClient(table string) (*DynamoDBClient, error) {
	var err error
	once.Do(func() {
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

func (c *DynamoDBClient) GetItem(key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	input := &dynamodb.GetItemInput{
		TableName: &c.table,
		Key:       key,
	}

	output, err := c.client.GetItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %v", err)
	}

	return output.Item, nil
}

func main() {
	client, err := NewDynamoDBClient("YourTableName")
	if err != nil {
		log.Fatalf("error creating DynamoDB client: %v", err)
	}

	item := map[string]types.AttributeValue{
		"Key": &types.AttributeValueMemberS{Value: "Value"},
	}

	if err := client.PutItem(item); err != nil {
		log.Fatalf("error putting item: %v", err)
	}

	key := map[string]types.AttributeValue{
		"Key": &types.AttributeValueMemberS{Value: "Value"},
	}

	result, err := client.GetItem(key)
	if err != nil {
		log.Fatalf("error getting item: %v", err)
	}

	fmt.Println("Retrieved Item:", result)
}
