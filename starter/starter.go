package starter

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/emanuelef/temporal-meetup-demo/utils"
	workflow "github.com/emanuelef/temporal-meetup-demo/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
)

var (
	once     sync.Once
	instance *TemporalClient
)

type TemporalClient struct {
	client client.Client
	table  string
}

func NewTemporalClient(ctx context.Context) (*TemporalClient, error) {
	var initErr error
	span := trace.SpanFromContext(ctx)
	once.Do(func() {
		span.AddEvent("First initialization of the Temporal client")
		// Check here error shadowing and cheeck after once.Do
		tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
		if err != nil {
			log.Println("Unable to create interceptor", err)
			span.AddEvent("Unable to create interceptor")
			return
		}

		options := client.Options{
			Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
		}

		// The client is a heavyweight object that should be created once per process.
		c, err := client.Dial(options)
		if err != nil {
			log.Println("Unable to create client", err)
			span.AddEvent("Unable to create client")
			return
		}
		//defer c.Close()

		instance = &TemporalClient{
			client: c,
			table:  "ciao",
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

func (c *TemporalClient) StartWorkflow(ctx context.Context) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Bool("isTrue", true), attribute.String("stringAttr", "Ciao"))

	span.AddEvent("Done Activity")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
	}

	temporalWorkflowId := fmt.Sprintf("service-%s", utils.GenerateUUID())

	workflowOptions := client.StartWorkflowOptions{
		ID:                 temporalWorkflowId,
		TaskQueue:          "otel",
		RetryPolicy:        retrypolicy,
		WorkflowRunTimeout: 6 * time.Minute,
	}

	we, err := c.client.ExecuteWorkflow(ctx, workflowOptions, workflow.Workflow, "Temporal")
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
