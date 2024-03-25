package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/emanuelef/temporal-meetup-demo/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/utils"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
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

var tracer trace.Tracer

var notToLogEndpoints = []string{"/health", "/metrics"}

func init() {
	tracer = otel.Tracer("github.com/emanuelef/go-gin-honeycomb")
}

func FilterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}

func main() {

	ctx := context.Background()

	tp, exp, err := otel_instrumentation.InitializeGlobalTracerProvider(ctx)

	if err != nil {
		log.Fatalf("error creating OTeL instrimentation: %v", err)
	}

	defer func() {
		_ = exp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
	}()

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("my-server", otelgin.WithFilter(FilterTraces)))

	// Just to check health and an example of a very frequent request
	// that we might not want to generate traces
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusNoContent, gin.H{})
	})

	r.GET("/hello", func(c *gin.Context) {
		_, childSpan := tracer.Start(c.Request.Context(), "custom-child-span")
		time.Sleep(10 * time.Millisecond) // simulate some work
		childSpan.End()
		c.JSON(http.StatusNoContent, gin.H{})
	})

	host := utils.GetEnv("HOST", "localhost")
	port := utils.GetEnv("PORT", "8080")
	hostAddress := fmt.Sprintf("%s:%s", host, port)

	err = r.Run(hostAddress)
	if err != nil {
		log.Printf("Starting router failed, %v", err)
	}
	/*
		dynamoClient, err := NewDynamoDBClient("Services")
		if err != nil {
			log.Fatalf("error creating DynamoDB client: %v", err)
		}

		partiQLStatement := "INSERT INTO Services VALUE { 'ID': 'example-id', 'Name': 'Example Service' }"

		// Execute the PartiQL statement
		output, err := dynamoClient.client.ExecuteStatement(context.TODO(), &dynamodb.ExecuteStatementInput{
			Statement: aws.String(partiQLStatement),
		})
		if err != nil {
			panic(fmt.Sprintf("unable to execute PartiQL statement, %v", err))
		}

		fmt.Println("Item added successfully:", output)
	*/
}
