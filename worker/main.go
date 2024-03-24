package main

import (
	"context"
	"log"

	otel "github.com/emanuelef/temporal-meetup-demo/otel"
	workflow "github.com/emanuelef/temporal-meetup-demo/workflow"
	_ "github.com/joho/godotenv/autoload"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := otel.InitializeGlobalTracerProvider(ctx)
	if err != nil {
		log.Fatalln("Unable to create a global trace provider", err)
	}

	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Println("Error shutting down trace provider:", err)
		}
	}()

	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Fatalln("Unable to create interceptor", err)
	}

	options := client.Options{
		Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(options)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "otel", worker.Options{})

	w.RegisterWorkflow(workflow.Workflow)
	w.RegisterActivity(workflow.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Worker run failed", err)
	}
}
