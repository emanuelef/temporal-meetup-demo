package workflow

import (
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var tracer trace.Tracer

type ServiceWorkflowInput struct {
	Name     string
	Metadata string
}

type ServiceWorkflowOutput struct {
	Name      string
	Activated bool
}

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("github.com/emanuelef/temporal-meetup-demo")
}

func Workflow(ctx workflow.Context, service ServiceWorkflowInput) (ServiceWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", service.Name)

	result := ServiceWorkflowOutput{
		Name: service.Name,
	}

	// TODO: How to get the ctx from the caller to get the span ?

	test1 := ctx.Value("SpanContextKey")

	if test1 != nil {
		logger.Info("Test")
	}

	//extractedBaggage := baggage.FromContext(ctx)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		// HeartbeatTimeout:    10 * time.Second,
	})

	err := workflow.ExecuteActivity(ctx, Activity).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return result, err
	}

	retrypolicy := &temporal.RetryPolicy{
		InitialInterval:    2 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Second * 100,
		MaximumAttempts:    3,
	}

	activityoptions := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 10 * time.Second,
		RetryPolicy:            retrypolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, activityoptions)

	err = workflow.ExecuteActivity(ctx, SecondActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("Second Activity failed.", "Error", err)
		return result, err
	}

	err = workflow.ExecuteActivity(ctx, ThirdActivity).Get(ctx, nil)
	if err != nil {
		logger.Error("Second Activity failed.", "Error", err)
		return result, err
	}

	logger.Info("HelloWorld workflow completed.")

	return result, nil
}
