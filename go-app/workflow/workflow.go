package workflow

import (
	"context"
	"io"
	"time"

	"github.com/emanuelef/temporal-meetup-demo/go-app/dynamo"
	"github.com/emanuelef/temporal-meetup-demo/go-app/s3"
	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

var tracer trace.Tracer

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("github.com/emanuelef/temporal-meetup-demo")
}

func Workflow(ctx workflow.Context, name string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		// HeartbeatTimeout:    10 * time.Second,
	})

	err := workflow.ExecuteActivity(ctx, Activity).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, SecondActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("Second Activity failed.", "Error", err)
		return err
	}

	logger.Info("HelloWorld workflow completed.")
	return nil
}

func Activity(ctx context.Context, name string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	// Get current span and add new attributes
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "Ciao"))

	// Create a child span
	_, childSpan := tracer.Start(ctx, "custom-span")
	time.Sleep(11 * time.Second)
	childSpan.End()

	dynamoClient, err := dynamo.NewDynamoDBClient(ctx, "ciao")

	if err != nil {
		return err
	}

	_, err = dynamoClient.ListItems(ctx)

	if err != nil {
		return err
	}

	// Add an event to the current span
	span.AddEvent("Done Activity")

	return nil
}

func SecondActivity(ctx context.Context, name string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	externalURL := "https://pokeapi.co/api/v2/pokemon/ditto"
	resp, err := otelhttp.Get(ctx, externalURL)

	if err != nil {
		return err
	}

	_, _ = io.ReadAll(resp.Body)

	time.Sleep(1 * time.Second)

	s3Client, err := s3.NewS3Client(ctx, "ciao")

	if err != nil {
		return err
	}

	_, err = s3Client.ListScripts(ctx)

	if err != nil {
		return err
	}

	return nil
}
