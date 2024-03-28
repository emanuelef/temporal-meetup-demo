package s3

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel/trace"
)

var SCRIPTS_S3_BUCKET = utils.GetEnv("SCRIPTS_S3_BUCKET", "local-asm-bucket")

const (
	DEFAULT_LOCALSTACK_HOST = "localhost"
	DEFAULT_LOCALSTACK_PORT = "4566"
)

var (
	once     sync.Once
	instance *S3Client
)

type S3Client struct {
	client *s3.Client
	table  string
}

func buildLocalConfiguration() (cfg aws.Config, err error) {
	const defaultRegion = "us-east-1"

	host := utils.GetEnv("LOCALSTACK_HOST", DEFAULT_LOCALSTACK_HOST)
	port := utils.GetEnv("LOCALSTACK_PORT", DEFAULT_LOCALSTACK_PORT)
	hostAddress := fmt.Sprintf("http://%s", net.JoinHostPort(host, port))

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               hostAddress,
			SigningRegion:     defaultRegion,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err = config.LoadDefaultConfig(context.Background(),
		config.WithRegion(defaultRegion),
		config.WithEndpointResolverWithOptions(resolver),
	)

	return cfg, err
}

func NewS3Client(ctx context.Context, table string) (*S3Client, error) {
	var err error
	span := trace.SpanFromContext(ctx)
	once.Do(func() {
		span.AddEvent("First initialization of the S3 client")
		cfg, loadErr := buildLocalConfiguration()

		otelaws.AppendMiddlewares(&cfg.APIOptions)

		if loadErr != nil {
			err = fmt.Errorf("unable to load SDK config: %v", loadErr)
			return
		}

		client := s3.NewFromConfig(cfg)

		instance = &S3Client{
			client: client,
			table:  table,
		}
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (c *S3Client) ListScripts(ctx context.Context) ([]string, error) {
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(SCRIPTS_S3_BUCKET),
	})

	scripts := []string{}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			scripts = append(scripts, strings.Replace(*obj.Key, "", "", 1))
		}
	}

	return scripts, nil
}
