package device

import (
	"time"

	"github.com/blackjack/webcam"
	"github.com/golang/glog"
	"github.com/saljam/mjpeg"
)

// V4L format identifiers from /usr/include/linux/videodev2.h.
const (
	MJPEG   webcam.PixelFormat = 1196444237
	YUYV422 webcam.PixelFormat = 1448695129
)

type Video struct {
	Stream      *mjpeg.Stream
	cam         *webcam.Webcam
	height      uint32
	width       uint32
	pixelFormat webcam.PixelFormat
	stop        chan struct{}
	fps         uint
	capStatus   bool
}

func NewVideo(pixelFormat webcam.PixelFormat, h, w uint32, fps uint) *Video {
	return &Video{
		pixelFormat: pixelFormat,
		height:      h,
		width:       w,
		stop:        make(chan struct{}),
		fps:         fps,
		capStatus:   false,
	}
}

func (s *Video) Init() error {
	cam, err := webcam.Open("/dev/video0")
	if err != nil {
		return err
	}

	// Initial image size.
	if _, _, _, err := cam.SetImageFormat(s.pixelFormat, s.width, s.height); err != nil {
		return err
	}

	s.cam = cam
	s.Stream = mjpeg.NewStream()

	return nil
}

func (s *Video) SetFormat(pixelFormat webcam.PixelFormat, h, w uint32) error {
	if _, _, _, err := s.cam.SetImageFormat(pixelFormat, w, h); err != nil {
		return err
	}
	return nil
}

func (s *Video) SetFPS(fps uint) {
	s.fps = fps
}

func (s *Video) StartVideoStream() {
	go s.startStreamer()

}

func (s *Video) StopVideoStream() {
	if s.capStatus {
		s.stop <- struct{}{}
		if err := s.cam.StopStreaming(); err != nil {
			glog.Errorf("Failed to start stream:%v", err)
		}
	}
}

func (s *Video) startStreamer() {

	// Since the ReadFrame is buffered, trying to read at FPS results in delay.
	fpsTicker := time.NewTicker(time.Duration(1000/s.fps) * time.Millisecond)

	if err := s.cam.StartStreaming(); err != nil {
		glog.Errorf("Failed to start stream:%v", err)
		return
	}

	s.capStatus = true
	glog.Infof("Started Video Capture")

	frame := []byte{}
	for {
		select {
		case <-s.stop:
			glog.Info("Stopped Video Capture")
			s.capStatus = false
			return

		default:
			if err := s.cam.WaitForFrame(5); err != nil {
				glog.Errorf("Failed to read webcam:%v", err)
			}
			var err error
			frame, err = s.cam.ReadFrame()
			if err != nil || len(frame) == 0 {
				glog.Errorf("Failed tp read webcam frame:%v or frame size 0", err)
			}

		case <-fpsTicker.C:
			s.Stream.UpdateJPEG(frame)
		}
	}
}
