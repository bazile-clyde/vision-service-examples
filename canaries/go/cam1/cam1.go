// set up credentials
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

// For data see "Trace Explorer" in GCP (https://console.cloud.google.com/traces/list?referrer=search&project=engineering-tools-310515)
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

	// STACKDRIVER_PROJECT_ID must be set before this step!
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

	// TODO: use unique name for this so we can distinguish in GCP. See examples of StartSpan in RDK
	ctx, span := trace.StartSpan(ctx, "Canary")
	for {

		// TODO: need to get distributed tracing working, so it nests the Read call inside of this call to Next.
		//  Otherwise, they look like two separate calls.
		// 	For a working solution see how https://github.com/viamrobotics/rdk/blob/5c7649e6e51a52f5f7dc36f7547ec4e9a18ff056/rimage/image_file.go#L266
		// 	appears nested under https://github.com/viamrobotics/rdk/blob/5c7649e6e51a52f5f7dc36f7547ec4e9a18ff056/components/camera/client.go#L61
		// TODO: ensure Next calls StartSpan. See https://github.com/viamrobotics/gostream/blob/44932aa9195421119c27661c6d88afa89c75c1b9/media.go#L379
		// 	for USB cameras.
		_, _, err := stream.Next(ctx)
		logger.Info("got image")
		handleErr(err)
		span.End()
	}
}
