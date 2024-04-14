package starter

import (
	"context"
	"fmt"
	"log"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	workflow "github.com/emanuelef/temporal-meetup-demo/go-app/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	//"go.temporal.io/sdk/temporal"
)

const TASK_QUEUE = "MeetupExample"

type Service struct {
	Name      string `json:"name"`
	DeviceMac string `json:"deviceMac"`
}

var instance *TemporalClient

type TemporalClient struct {
	client client.Client
}

func GetTemporalClient(ctx context.Context) (*TemporalClient, error) {
	var initErr error
	span := trace.SpanFromContext(ctx)

	// shortcut, might be using once instead
	if instance == nil {
		span.AddEvent("First initialization of the Temporal client")

		// Note: A custom SpanContextKey is needed to retrieve the current span in the Workflow definition
		// The Activities can get the current span from the standard context without the need to use a custom SpanContextKey
		tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
			SpanContextKey: workflow.SpanContextKey,
		})

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
		}
	}

	return instance, nil
}

func (c *TemporalClient) StartWorkflow(ctx context.Context, service Service) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("service", service.Name), attribute.String("device.mac", service.DeviceMac))

	/* 	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
	} */

	temporalWorkflowId := fmt.Sprintf("service-%s-%s", service.Name, utils.GenerateUUID())

	workflowOptions := client.StartWorkflowOptions{
		ID:        temporalWorkflowId,
		TaskQueue: TASK_QUEUE,
		//	RetryPolicy:        retrypolicy,
		// WorkflowRunTimeout: 6 * time.Minute,
	}

	if c.client == nil {
		err := fmt.Errorf("The Temporal Client has not been initialized")
		return "", err
	}

	workflowInput := workflow.ServiceWorkflowInput{
		Name:      service.Name,
		Metadata:  "simple payload",
		DeviceMac: service.DeviceMac,
	}

	// Baggage is propagated to all spans but attributes must be created if needed
	m0, _ := baggage.NewMemberRaw("service_name", workflowInput.Name)
	m1, _ := baggage.NewMemberRaw("device_mac", workflowInput.DeviceMac)
	b, _ := baggage.New(m0, m1)
	ctx = baggage.ContextWithBaggage(ctx, b)

	we, err := c.client.ExecuteWorkflow(ctx, workflowOptions, workflow.Workflow, workflowInput)
	if err != nil {
		log.Println("Unable to execute workflow", err)
		span.AddEvent("Unable to execute workflow")
		span.SetAttributes(attribute.String("error.message", err.Error()))
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
