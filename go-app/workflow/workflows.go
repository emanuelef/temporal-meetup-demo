package workflow

import (
	"time"

	// "github.com/emanuelef/temporal-meetup-demo/go-app/otel_instrumentation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var tracer trace.Tracer

const SpanContextKey = "span-context-key-workflow"

type ServiceWorkflowInput struct {
	Name      string
	Metadata  string
	DeviceMac string
}

type ServiceWorkflowOutput struct {
	Name      string
	Activated bool
}

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("github.com/emanuelef/temporal-meetup-demo/go-app/workflow")
}

func ProvisioningWorkflow(ctx workflow.Context, service ServiceWorkflowInput) (ServiceWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Main workflow started", "name", service.Name)

	result := ServiceWorkflowOutput{
		Name: service.Name,
	}

	span, ok := ctx.Value(SpanContextKey).(trace.Span)

	if !ok {
		span = noop.Span{}
	}

	span.SetAttributes(attribute.String("provisioning", service.Name), attribute.String("device.mac", service.DeviceMac))

	if service.DeviceMac == "FF:BB:CC:11:11:77" {
		span.SetAttributes(attribute.String("firmware.version", "1.9"))
	} else {
		span.SetAttributes(attribute.String("firmware.version", "2.1"))
	}

	// _ = otel_instrumentation.AddLogEvent(span, service)

	// TODO: How to get the Baggage from workflow.Context ?
	// extractedBaggage := baggage.FromContext(ctx)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	})

	// Simulate task to prepare what is needed to start activities
	_ = workflow.Sleep(ctx, 300*time.Millisecond)
	span.AddEvent("Completed Activities preparation")

	err := workflow.ExecuteActivity(ctx, FirstActivity, "configGRE").Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return result, err
	}

	// TODO: How to generate a child span from workflow.Context ?
	_ = workflow.Sleep(ctx, 500*time.Millisecond)

	span.AddEvent("Finished using first activity results")

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    2 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
		MaximumAttempts:    3,
	}

	activityoptions := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 2 * time.Minute,
		RetryPolicy:            retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, activityoptions)

	selector := workflow.NewSelector(ctx)

	secondActivityFuture := workflow.ExecuteActivity(ctx, SecondActivity, service.Name, service.DeviceMac)
	if err != nil {
		logger.Error("Second Activity failed.", "Error", err)
		return result, err
	}

	thirdActivityFuture := workflow.ExecuteActivity(ctx, ThirdActivity, "configHotspot")
	if err != nil {
		logger.Error("Second Activity failed.", "Error", err)
		return result, err
	}

	pendingFutures := []workflow.Future{secondActivityFuture, thirdActivityFuture}

	for _, future := range pendingFutures {
		selector.AddFuture(future, func(f workflow.Future) {
			err1 := f.Get(ctx, nil)
			if err1 != nil {
				err = err1
				return
			}
		})
	}

	for range len(pendingFutures) {
		selector.Select(ctx)
		if err != nil {
			return result, err
		}
	}

	logger.Info("HelloWorld workflow completed.")

	return result, nil
}
