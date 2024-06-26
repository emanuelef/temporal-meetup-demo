package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/go-app/protos"

	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var tracer trace.Tracer

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("github.com/emanuelef/temporal-meetup-demo/go-app/grpc-server")
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// server is used to implement helloworld.DeviceConfiguratorServer.
type server struct {
	protos.UnimplementedDeviceConfiguratorServer
}

// UpdateDeviceConfig implements helloworld.DeviceConfiguratorServer
func (s *server) UpdateDeviceConfig(ctx context.Context, in *protos.DeviceConfig) (*protos.ConfigResult, error) {
	log.Printf("Received: %v", in.GetConfigWiFi())

	_, childSpan := tracer.Start(ctx, "UpdateDeviceConfigCustom")
	time.Sleep(600 * time.Millisecond)
	childSpan.End()

	time.Sleep(200 * time.Millisecond)

	if in.ConfigWiFi == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request missing required field: ConfigWiFi")
	}

	return &protos.ConfigResult{Ack: "Hello " + in.GetConfigWiFi()}, nil
}

func main() {
	ctx := context.Background()
	tp, exp, _ := otel_instrumentation.InitializeGlobalTracerProvider(ctx)

	// Handle shutdown to ensure all sub processes are closed correctly and telemetry is exported
	defer func() {
		_ = exp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
	}()

	host := getEnv("HOST", "localhost")
	port := getEnv("PORT", "7070")
	hostAddress := fmt.Sprintf("%s:%s", host, port)

	lis, err := net.Listen("tcp", hostAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	// Register the server
	protos.RegisterDeviceConfiguratorServer(grpcServer, &server{})

	log.Printf("Starting server on address %s", lis.Addr().String())

	// Start listening
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}
