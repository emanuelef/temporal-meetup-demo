package main

import (
	"context"
	"fmt"
	"log"

	"github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	workflow "github.com/emanuelef/temporal-meetup-demo/go-app/workflow"
	_ "github.com/joho/godotenv/autoload"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

const TASK_QUEUE = "MeetupExample"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, _, err := otel_instrumentation.InitializeGlobalTracerProvider(ctx)
	if err != nil {
		log.Fatalln("Unable to create a global trace provider", err)
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Println("Error shutting down trace provider:", err)
		}
	}()

	// Note: A custom SpanContextKey is needed to retrieve the current span in the Workflow definition
	// The Activities can get the current span from the standard context without the need to use a custom SpanContextKey
	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		SpanContextKey: workflow.SpanContextKey,
	})
	if err != nil {
		log.Fatalln("Unable to create interceptor", err)
	}

	temporalEndpoint := fmt.Sprintf("%s:%s",
		utils.GetEnv("TEMPORAL_HOST", "localhost"),
		utils.GetEnv("TEMPORAL_PORT", "7233"))

	options := client.Options{
		Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
		HostPort:     temporalEndpoint,
	}

	c, err := client.Dial(options)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, TASK_QUEUE, worker.Options{})

	w.RegisterWorkflow(workflow.ProvisioningWorkflow)
	w.RegisterActivity(workflow.FirstActivity)
	w.RegisterActivity(workflow.SecondActivity)
	w.RegisterActivity(workflow.ThirdActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Worker run failed", err)
	}
}
