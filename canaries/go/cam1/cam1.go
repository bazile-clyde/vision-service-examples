// # set up creds
// $ STACKDRIVER_PROJECT_ID="engineering-tools-310515" go run cam1.go
package main

import (
	"context"
	"github.com/edaniels/golog"
	"go.opencensus.io/stats"
	"go.opencensus.io/trace"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/perf"
	"go.viam.com/utils/rpc"
)

var (
	// The task latency in milliseconds.
	latencyMs = stats.Float64("task_latency", "The task latency in milliseconds", "ms")
)

func connect(logger golog.Logger) (*client.RobotClient, error) {
	return client.New(
		context.Background(),
		"starter-bot-main.k4xl69bmso.viam.cloud",
		logger,
		client.WithDialOptions(rpc.WithCredentials(rpc.Credentials{
			Type:    utils.CredentialsTypeRobotLocationSecret,
			Payload: "e7ztx7s67d4pnnnor54qkn2wlhf58cd7erukcldonha0cc62",
		})),
	)
}

func main() {
	logger := golog.NewDevelopmentLogger("client")
	handleErr := func(err error) {
		if err != nil {
			logger.Fatal(err)
		}
	}

	ctx := context.Background()
	robot, err := connect(logger)
	handleErr(err)
	defer robot.Close(ctx)

	comp, err := camera.FromRobot(robot, "cam")
	handleErr(err)

	stream, err := comp.Stream(ctx)
	handleErr(err)

	defer stream.Close(ctx)

	opts := perf.CloudOptions{
		Context:      ctx,
		Logger:       logger,
		MetricPrefix: "canary",
		// TODO: don't trace every call. Do some sampling so we don't eat up memory and run out of money
	}
	// TODO: This is going to look for creds. Make sure those are setup
	// TODO: we do NOT export metrics from RDK, only in App.
	exporter, err := perf.NewCloudExporter(opts)
	handleErr(err)
	handleErr(exporter.Start())

	defer exporter.Stop()

	for {
		// TODO: use unique name for this so we can distinguish in GCP
		//  example: ctx, span := trace.StartSpan(ctx, "example.com/Run")
		ctx, span := trace.StartSpan(ctx, "Canary")

		// TODO: need to get distributed tracing working, so it nests the Read call inside of this call to Next.
		//  Otherwise, they look like two separate calls.
		// TODO: if guard code in hotspot areas... or maybe everywhere.
		// TODO: add telemetry to decodes and encodes
		_, _, err := stream.Next(ctx)
		logger.Info("got image")
		handleErr(err)
		span.End()
	}
}
