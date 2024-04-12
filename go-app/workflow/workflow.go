package workflow

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/emanuelef/temporal-meetup-demo/go-app/dynamo"
	"github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/go-app/s3"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var tracer trace.Tracer

type ServiceWorkflowInput struct {
	Name     string
	Metadata string
}

type ServiceWorkflowOutput struct {
	Name      string
	Activated bool
}

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("github.com/emanuelef/temporal-meetup-demo")
}

func Workflow(ctx workflow.Context, service ServiceWorkflowInput) (ServiceWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", service.Name)

	result := ServiceWorkflowOutput{
		Name: service.Name,
	}

	// TODO: How to get the ctx from the caller to get the span ?

	test1 := ctx.Value("SpanContextKey")

	if test1 != nil {
		logger.Info("Test")
	}

	//extractedBaggage := baggage.FromContext(ctx)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		// HeartbeatTimeout:    10 * time.Second,
	})

	err := workflow.ExecuteActivity(ctx, Activity).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return result, err
	}

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    2 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
		MaximumAttempts:    3,
	}

	activityoptions := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 10 * time.Second,
		RetryPolicy:            retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, activityoptions)

	err = workflow.ExecuteActivity(ctx, SecondActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("Second Activity failed.", "Error", err)
		return result, err
	}

	logger.Info("HelloWorld workflow completed.")

	return result, nil
}

func Activity(ctx context.Context, name string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	extractedBaggage := baggage.FromContext(ctx)

	// Get current span and add new attributes
	span := trace.SpanFromContext(ctx)

	for _, val := range extractedBaggage.Members() {
		if val.Key() == "original_request_endpoint" {
			span.SetAttributes(attribute.String(val.Key(), val.Value()))
		}
	}

	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "Ciao"))

	otel_instrumentation.AddLogEvent(span, ServiceWorkflowInput{Name: "Good", Metadata: "Day"})

	// Create a child span
	_, childSpan := tracer.Start(ctx, "custom-span")
	time.Sleep(3 * time.Second)
	childSpan.End()

	dynamoClient, err := dynamo.NewDynamoDBClient(ctx, "ciao")
	if err != nil {
		return err
	}

	_, err = dynamoClient.ListItems(ctx)

	if err != nil {
		return err
	}

	/* 	externalURL := "http://rust-app:8080/hello"
	   	resp, err := otelhttp.Get(ctx, externalURL)

	   	if err != nil {
	   		return err
	   	}

	   	_, _ = io.ReadAll(resp.Body) */

	host := utils.GetEnv("ANOMALY_HOST", "localhost")
	port := utils.GetEnv("ANOMALY_PORT", "8086")
	anomalyHostAddress := fmt.Sprintf("http://%s", net.JoinHostPort(host, port))

/* 	externalURL := anomalyHostAddress + "/predict?repo=databricks/dbrx"
	resp, err := otelhttp.Get(ctx, externalURL)
	if err != nil {
		return err
	} */

	externalURL := anomalyHostAddress + "/test"
	resp, err := otelhttp.Get(ctx, externalURL)
	if err != nil {
		return err
	}

	_, _ = io.ReadAll(resp.Body)

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
