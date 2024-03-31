package starter

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	workflow "github.com/emanuelef/temporal-meetup-demo/go-app/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
)

const TASK_QUEUE = "MeetupExample"

var (
	instance *TemporalClient
)

type TemporalClient struct {
	client client.Client
	table  string
}

func GetTemporalClient(ctx context.Context) (*TemporalClient, error) {
	var initErr error
	span := trace.SpanFromContext(ctx)

	// shortcut, might be using once instead
	if instance == nil {
		span.AddEvent("First initialization of the Temporal client")
		tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
		if err != nil {
			log.Println("Unable to create interceptor", err)
			span.AddEvent("Unable to create interceptor")
			initErr = err
			return nil, initErr
		}

		temporalEndpoint := fmt.Sprintf("%s:%s",
			utils.GetEnv("TEMPORAL_HOST", "localhost"),
			utils.GetEnv("TEMPORAL_PORT", "7233"))

		options := client.Options{
			Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
			HostPort:     temporalEndpoint,
		}

		// The client is a heavyweight object that should be created once per process.
		c, err := client.Dial(options)
		if err != nil {
			log.Println("Unable to create client", err)
			span.AddEvent("Unable to create client")
			initErr = err
			return nil, initErr
		}
		// defer c.Close() // When is the client closed ?

		instance = &TemporalClient{
			client: c,
			table:  "ciao",
		}
	}

	return instance, nil
}

func (c *TemporalClient) StartWorkflow(ctx context.Context) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "Ciao"))

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
	}

	temporalWorkflowId := fmt.Sprintf("service-%s", utils.GenerateUUID())

	workflowOptions := client.StartWorkflowOptions{
		ID:                 temporalWorkflowId,
		TaskQueue:          TASK_QUEUE,
		RetryPolicy:        retrypolicy,
		WorkflowRunTimeout: 6 * time.Minute,
	}

	if c.client == nil {
		err := fmt.Errorf("The Temporal Client has not been initialized")
		return "", err
	}

	we, err := c.client.ExecuteWorkflow(ctx, workflowOptions, workflow.Workflow, "Temporal")
	if err != nil {
		log.Println("Unable to execute workflow", err)
		span.AddEvent("Unable to execute workflow")
		return "", err
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	span.AddEvent("Started workflow")

	// Synchronously wait for the workflow completion.
	/* 	var result string
	   	err = we.Get(context.Background(), &result)
	   	if err != nil {
	   		log.Println("Unable get workflow result", err)
	   		span.AddEvent("Unable get workflow result")
	   		return "", err
	   	}
	   	log.Println("Workflow result:", result) */
	return we.GetID(), nil
}
