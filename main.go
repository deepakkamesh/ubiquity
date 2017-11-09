package main

import (
	"flag"
	"time"
	"ubiquity/device"
	"ubiquity/httphandler"

	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"

	"github.com/golang/glog"
)

var (
	buildtime string // Compiler Flags
	githash   string // Compiler Flags
)

func main() {
	var (
		res          = flag.String("resources", "./resources", "resources directory")
		httpHostPort = flag.String("http_port", ":8080", "host:port number for http")
		mlfwd        = flag.String("left_motor_fwd_pin", "11", "Motor controller")
		mlbwd        = flag.String("left_motor_bwd_pin", "7", "Motor controller")
		mrfwd        = flag.String("right_motor_fwd_pin", "13", "Motor controller")
		mrbwd        = flag.String("right_motor_bwd_pin", "15", "Motor controller")
	)
	flag.Parse()
	glog.Infof("Starting Ubiquity ver %s build on %s", githash, buildtime)

	// Log flush Routine.
	go func() {
		for {
			glog.Flush()
			time.Sleep(300 * time.Millisecond)
		}
	}()

	// Initialize PI Adaptor.
	pi := raspi.NewAdaptor()
	if err := pi.Connect(); err != nil {
		glog.Fatalf("Failed to initialize Adapter:%v", err)
	}

	// Initialize devices.
	motorRightFwd := gpio.NewDirectPinDriver(pi, *mrfwd)
	if err := motorRightFwd.Start(); err != nil {
		glog.Fatalf("Failed to setup GPIO: %v", err)
	}

	motorRightBwd := gpio.NewDirectPinDriver(pi, *mrbwd)
	if err := motorRightBwd.Start(); err != nil {
		glog.Fatalf("Failed to setup GPIO: %v", err)
	}

	motorLeftFwd := gpio.NewDirectPinDriver(pi, *mlfwd)
	if err := motorRightFwd.Start(); err != nil {
		glog.Fatalf("Failed to setup GPIO: %v", err)
	}

	motorLeftBwd := gpio.NewDirectPinDriver(pi, *mlbwd)
	if err := motorRightBwd.Start(); err != nil {
		glog.Fatalf("Failed to setup GPIO: %v", err)
	}

	dev := device.New(motorRightFwd, motorRightBwd, motorLeftFwd, motorLeftBwd)

	// Startup HTTP service.
	h := httphandler.New(dev)
	if err := h.Start(*httpHostPort, *res); err != nil {
		glog.Fatalf("Failed to start HTTP: %v", err)
	}

}
