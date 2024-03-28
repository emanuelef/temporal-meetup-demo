package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"github.com/emanuelef/temporal-meetup-demo/go-app/starter"
	"github.com/emanuelef/temporal-meetup-demo/go-app/utils"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

var notToLogEndpoints = []string{"/health", "/metrics"}

func init() {
	tracer = otel.Tracer("github.com/emanuelef/go-gin-honeycomb")
}

func FilterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}

func main() {
	ctx := context.Background()

	tp, exp, err := otel_instrumentation.InitializeGlobalTracerProvider(ctx)
	if err != nil {
		log.Fatalf("error creating OTeL instrumentation: %v", err)
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
		_, childSpan := tracer.Start(c.Request.Context(), "custom-child-span")
		time.Sleep(10 * time.Millisecond) // simulate some work
		childSpan.End()
		c.JSON(http.StatusNoContent, gin.H{})
	})

	r.GET("/start", func(c *gin.Context) {
		clientTemporal, _ := starter.NewTemporalClient(c.Request.Context())

		workflowID, err := clientTemporal.StartWorkflow(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"workflowID": workflowID})
	})

	host := utils.GetEnv("HOST", "localhost")
	port := utils.GetEnv("PORT", "8080")
	hostAddress := fmt.Sprintf("%s:%s", host, port)

	err = r.Run(hostAddress)
	if err != nil {
		log.Printf("Starting router failed, %v", err)
	}
}