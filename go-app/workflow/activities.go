package workflow

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/emanuelef/temporal-meetup-demo/go-app/dynamo"
	"github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/go-app/protos"
	"github.com/emanuelef/temporal-meetup-demo/go-app/s3"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/activity"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Activity(ctx context.Context, name string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	extractedBaggage := baggage.FromContext(ctx)

	// Get current span and add new attributes
	span := trace.SpanFromContext(ctx)

	// Add attributes from baggage
	for _, val := range extractedBaggage.Members() {
		span.SetAttributes(attribute.String(val.Key(), val.Value()))
	}

	_ = otel_instrumentation.AddLogEvent(span, ServiceWorkflowInput{Name: "Good", Metadata: "Day"})

	// Create a child span
	_, childSpan := tracer.Start(ctx, "custom-span")
	time.Sleep(time.Duration(300+rand.Intn(200)) * time.Millisecond)
	childSpan.End()

	group := errgroup.Group{}

	for _, val := range []string{"a", "b", "c"} {
		group.Go(func() error {
			time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)
			_, childSpan := tracer.Start(ctx, "custom-span-"+val) // no longer an issue with Go 1.22
			time.Sleep(time.Duration(100+rand.Intn(400)) * time.Millisecond)
			childSpan.End()
			return nil
		})
	}

	// Wait for all goroutines to finish
	if err := group.Wait(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	dynamoClient, err := dynamo.NewDynamoDBClient(ctx, "services")
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

	externalURL := anomalyHostAddress + "/check"
	resp, err := otelhttp.Get(ctx, externalURL)
	if err != nil {
		return err
	}

	_, _ = io.ReadAll(resp.Body)

	// Add an event to the current span
	span.AddEvent("Done Activity")

	return nil
}

func SecondActivity(ctx context.Context, serviceName, deviceMac string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", serviceName)

	externalURL := "https://pokeapi.co/api/v2/pokemon/ditto"
	resp, err := otelhttp.Get(ctx, externalURL)
	if err != nil {
		return err
	}

	_, _ = io.ReadAll(resp.Body)

	time.Sleep(400 * time.Millisecond)

	// Simulate a longer than usual operation for 
	if deviceMac == "FF:BB:CC:11:11:77" {
		time.Sleep(time.Duration(2600+rand.Intn(1000)) * time.Millisecond)
	}

	s3Client, err := s3.NewS3Client(ctx, "scripts")
	if err != nil {
		return err
	}

	_, err = s3Client.ListScripts(ctx)

	if err != nil {
		return err
	}

	return nil
}

func ThirdActivity(ctx context.Context, name string) error {
	grpcHost := utils.GetEnv("GRPC_TARGET", "localhost")
	grpcTarget := fmt.Sprintf("%s:7070", grpcHost)

	conn, err := grpc.NewClient(grpcTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("Did not connect: %v", err)
		return nil
	}

	defer conn.Close()
	cli := protos.NewGreeterClient(conn)

	r, err := cli.SayHello(ctx, &protos.HelloRequest{Greeting: "ciao"})
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}

	log.Printf("Greeting: %s", r.GetReply())

	return nil
}
