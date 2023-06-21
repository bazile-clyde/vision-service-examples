package main

import (
	"context"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
	"image"

	"github.com/edaniels/golog"
	"github.com/webview/webview"
	"go.viam.com/rdk/components/camera"
	// registers all cameras.
	_ "go.viam.com/rdk/components/camera/register"
	// registers the vision service
	_ "go.viam.com/rdk/services/vision/register"
	// registers the mlmodel service
	_ "go.viam.com/rdk/services/mlmodel/register"
)

const (
	pathHere = "/full/path/to/vision-service-examples/go/"
	saveTo   = "output.jpg"
)

func main() {
	logger := golog.NewDevelopmentLogger("client")
	r, err := client.New(
		context.Background(),
		"starter-bot-main.k4xl69bmso.viam.cloud",
		logger,
		client.WithDialOptions(rpc.WithCredentials(rpc.Credentials{
			Type:    utils.CredentialsTypeRobotLocationSecret,
			Payload: "e7ztx7s67d4pnnnor54qkn2wlhf58cd7erukcldonha0cc62",
		})),
	)
	if err != nil {
		logger.Fatal(err)
	}

	// Print available resources
	logger.Info("Resources:")
	logger.Info(r.ResourceNames())

	// grab the camera from the robot
	cameraName := "cam1" // make sure to use the same name as in the json/APP
	cam, err := camera.FromRobot(r, cameraName)
	if err != nil {
		logger.Fatalf("cannot get camera: %v", err)
	}
	camStream, err := cam.Stream(context.Background())
	if err != nil {
		logger.Fatalf("cannot get camera stream: %v", err)
	}

	for {
		// get image
		img, release, err := camStream.Next(context.Background())
		if err != nil {
			logger.Fatalf("cannot get image: %v", err)
		}
		defer release()

		displayImage(img) // once this is called, you're done.
	}

}

// displayImage will display the image stored in pathHere+saveTo
func displayImage(img image.Image) {
	webView := webview.New(true)
	defer webView.Destroy()
	webView.SetSize(img.Bounds().Dx()+100, img.Bounds().Dy()+100, webview.HintFixed)
	webView.Run()
}
