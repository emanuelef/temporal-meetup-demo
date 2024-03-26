package starter

import (
	"context"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	workflow "github.com/emanuelef/temporal-meetup-demo/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
)

func StartWorkflow(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "Ciao"))

	span.AddEvent("Done Activity")

	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Println("Unable to create interceptor", err)
		span.AddEvent("Unable to create interceptor")
		return err
	}

	options := client.Options{
		Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(options)
	if err != nil {
		log.Println("Unable to create client", err)
		span.AddEvent("Unable to create client")
		return err
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "otel_workflowID",
		TaskQueue: "otel",
	}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, workflow.Workflow, "Temporal")
	if err != nil {
		log.Println("Unable to execute workflow", err)
		span.AddEvent("Unable to execute workflow")
		return err
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Println("Unable get workflow result", err)
		span.AddEvent("Unable get workflow result")
		return err
	}
	log.Println("Workflow result:", result)
	return nil
}
