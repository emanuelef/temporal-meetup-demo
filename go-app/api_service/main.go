package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/go-app/protos"
	"github.com/emanuelef/temporal-meetup-demo/go-app/starter"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tracer            trace.Tracer
	notToLogEndpoints = []string{"/health", "/metrics"}
)

func init() {
	tracer = otel.Tracer("github.com/emanuelef/temporal-meetup-demo/go-app/api_service")
}

func FilterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}

func main() {
	ctx := context.Background()

	tp, exp, err := otel_instrumentation.InitializeGlobalTracerProvider(ctx)
	if err != nil {
		log.Fatalf("error creating OTel instrumentation: %v", err)
	}

	defer func() {
		_ = exp.Shutdown(ctx)
		_ = tp.Shutdown(ctx)
	}()

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(otelgin.Middleware("my-server", otelgin.WithFilter(FilterTraces)))

	// Just to check health and an example of a very frequent request
	// that we might not want to generate traces
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusNoContent, gin.H{})
	})

	r.GET("/hello", func(c *gin.Context) {
		ctx, childSpan := tracer.Start(c.Request.Context(), "custom-child-span")
		time.Sleep(10 * time.Millisecond) // simulate some work

		externalURL := "http://localhost:8090/rand"
		resp, _ := otelhttp.Get(ctx, externalURL)

		_, _ = io.ReadAll(resp.Body)

		childSpan.End()

		externalURL = "http://localhost:8086/predict?repo=databricks/dbrx"
		resp, _ = otelhttp.Get(ctx, externalURL)

		_, _ = io.ReadAll(resp.Body)

		c.JSON(http.StatusNoContent, gin.H{})
	})

	r.GET("/hello-grpc", func(c *gin.Context) {
		grpcHost := utils.GetEnv("GRPC_TARGET", "localhost")
		grpcTarget := fmt.Sprintf("%s:7070", grpcHost)

		conn, err := grpc.NewClient(grpcTarget,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
		if err != nil {
			log.Printf("Did not connect: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		defer conn.Close()
		cli := protos.NewDeviceConfiguratorClient(conn)

		r, err := cli.UpdateDeviceConfig(c.Request.Context(), &protos.DeviceConfig{ConfigWiFi: "ciao"})
		if err != nil {
			log.Printf("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		log.Printf("ConfigWiFi: %s", r.GetAck())

		c.JSON(http.StatusNoContent, gin.H{})
	})

	r.GET("/start", func(c *gin.Context) {
		ctx, childSpan := tracer.Start(c.Request.Context(), "prepare-workflow-payload")
		defer childSpan.End()

		// Example of an HTTP request needed to prepare the Workflow payload
		externalURL := "https://pokeapi.co/api/v2/pokemon/ditto"
		resp, err := otelhttp.Get(ctx, externalURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, _ = io.ReadAll(resp.Body)

		clientTemporal, err := starter.GetTemporalClient(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		workflowID, err := clientTemporal.StartWorkflow(c.Request.Context(), starter.Service{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"workflowID": workflowID})
	})

	r.POST("/provision", func(c *gin.Context) {
		ctx := c.Request.Context()

		span := trace.SpanFromContext(ctx)
		ctx, childSpan := tracer.Start(ctx, "prepare-workflow-payload")
		defer childSpan.End()

		time.Sleep(200 * time.Millisecond)

		var service starter.Service

		err := c.BindJSON(&service)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		span.SetAttributes(attribute.String("provisioning", service.Name), attribute.String("device.mac", service.DeviceMac))

		// Example of an HTTP request needed to prepare the Workflow payload
		externalURL := "https://pokeapi.co/api/v2/pokemon/ditto"
		resp, err := otelhttp.Get(ctx, externalURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, _ = io.ReadAll(resp.Body)

		clientTemporal, err := starter.GetTemporalClient(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		workflowID, err := clientTemporal.StartWorkflow(ctx, service)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"workflowID": workflowID})
	})

	host := utils.GetEnv("HOST", "localhost")
	port := utils.GetEnv("PORT", "8080")
	hostAddress := fmt.Sprintf("%s:%s", host, port)

	log.Printf("Starting web server %s\n", hostAddress)

	err = r.Run(hostAddress)
	if err != nil {
		log.Printf("Starting router failed, %v", err)
	}
}
