package main

import (
	"flag"
	"time"

	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"

	"github.com/deepakkamesh/ubiquity/device"
	"github.com/deepakkamesh/ubiquity/httphandler"
	"github.com/golang/glog"
)

var (
	buildtime string // Compiler Flags
	githash   string // Compiler Flags
)

func main() {
	var (
		res          = flag.String("resources", "../resources", "resources directory")
		httpHostPort = flag.String("http_port", ":8080", "host:port number for http")
		mlfwd        = flag.String("left_motor_fwd_pin", "11", "Motor controller")
		mlbwd        = flag.String("left_motor_bwd_pin", "7", "Motor controller")
		mrfwd        = flag.String("right_motor_fwd_pin", "13", "Motor controller")
		mrbwd        = flag.String("right_motor_bwd_pin", "15", "Motor controller")

		ssl        = flag.Bool("serve_ssl", true, "Serve HTTP over ssl")
		sslCert    = flag.String("ssl_cert", "cert.pem", "The SSL certificate in resources dir")
		sslPrivKey = flag.String("ssl_priv_key", "privkey.pem", "SSL private Keyname in resources dir")
		enPi       = flag.Bool("enable_pi_gpio", false, "Enable PI GPIO, I2C etc")

		enVid     = flag.Bool("enable_video", false, "Enable Video")
		vidHeight = flag.Uint("vid_height", 480, "Video Height")
		vidWidth  = flag.Uint("vid_width", 640, "Video Width")

		enAud = flag.Bool("enable_audio", false, "Enable Audio")
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

	var (
		motorRightBwd, motorRightFwd *gpio.DirectPinDriver
		motorLeftBwd, motorLeftFwd   *gpio.DirectPinDriver
		servo                        *device.Servo
	)

	if *enPi {
		// Initialize PI Adaptor.
		pi := raspi.NewAdaptor()
		if err := pi.Connect(); err != nil {
			glog.Fatalf("Failed to initialize Adapter:%v", err)
		}

		// Initialize motor devices.
		motorRightFwd = gpio.NewDirectPinDriver(pi, *mrfwd)
		if err := motorRightFwd.Start(); err != nil {
			glog.Fatalf("Failed to setup GPIO: %v", err)
		}

		motorRightBwd = gpio.NewDirectPinDriver(pi, *mrbwd)
		if err := motorRightBwd.Start(); err != nil {
			glog.Fatalf("Failed to setup GPIO: %v", err)
		}

		motorLeftFwd = gpio.NewDirectPinDriver(pi, *mlfwd)
		if err := motorRightFwd.Start(); err != nil {
			glog.Fatalf("Failed to setup GPIO: %v", err)
		}

		motorLeftBwd = gpio.NewDirectPinDriver(pi, *mlbwd)
		if err := motorRightBwd.Start(); err != nil {
			glog.Fatalf("Failed to setup GPIO: %v", err)
		}

		// Initialize Servo.
		servo = device.NewServo(20000, "23", pi)
		servo.SetAngle(90)
	}

	// Initialize new Ubiquity Device.
	dev := device.New(motorRightFwd, motorRightBwd, motorLeftFwd, motorLeftBwd, servo)

	// Initialize audio device.
	var aud *device.Audio
	if *enAud {
		aud = device.NewAudio()
		if err := aud.Init(512, 186, 8000, 4000); err != nil {
			glog.Fatalf("Unable to initialize audio:%v", err)
		}
	}

	// Initialize video device.
	var vid *device.Video
	if *enVid {
		vid = device.NewVideo(device.MJPEG, uint32(*vidWidth), uint32(*vidHeight), 2)
	}

	// Startup HTTP service.
	h := httphandler.New(dev, aud, vid)
	if err := h.Start(*httpHostPort, *res, *sslCert, *sslPrivKey, *ssl); err != nil {
		glog.Fatalf("Failed to start HTTP: %v", err)
	}
}
