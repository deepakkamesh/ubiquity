package httphandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/deepakkamesh/ubiquity/device"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

// Control Message Types.
const (
	ERR = iota
	CMD
	AUDIO_START
	AUDIO_STOP
	DRIVE_FWD
	DRIVE_BWD
	DRIVE_LEFT
	DRIVE_RIGHT
	SERVO_UP
	SERVO_DOWN
	SERVO_STEP
)

// Control Message.
type ControlMsg struct {
	CmdType int
	Data    interface{}
}

type Server struct {
	dev   *device.Ubiquity
	audio *device.Audio

	connCount  int // number of connected http clients.
	servoStep  int // Servo step for each click.
	servoAngle int // Current Angle for servo.
}

func New(dev *device.Ubiquity, aud *device.Audio) *Server {
	return &Server{
		dev:        dev,
		audio:      aud,
		servoAngle: 90,
		servoStep:  30,
	}
}

func (s *Server) Start(hostPort string, resPath string) error {

	// http routers.
	http.HandleFunc("/audiostream", s.audioSock)
	http.HandleFunc("/control", s.controlSock)

	// Serve static content from resources dir.
	fs := http.FileServer(http.Dir(resPath))
	http.Handle("/", fs)

	return http.ListenAndServeTLS(hostPort, resPath+"/server.crt", resPath+"/server.key", nil)
}

// controlSock handles the control messages from the http client.
func (s *Server) controlSock(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("Failed to upgrade conn:%v", err)
		return
	}

	s.connCount++

	defer func() {
		c.Close()
		s.connCount--
	}()

	for {
		_, data, err := c.ReadMessage()
		if err != nil {
			glog.Errorf("Websocket read error: %v", err)
			return
		}
		var msg ControlMsg
		json.Unmarshal(data, &msg)
		glog.V(2).Infof("Got control message type payload:%v", msg)

		switch msg.CmdType {
		case DRIVE_FWD:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_FWD, time.Duration(dur)); err != nil {
				glog.Errorf("Failed to move motor %v", err)
			}

		case DRIVE_BWD:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_BWD, time.Duration(dur)); err != nil {
				glog.Errorf("Failed to move motor %v", err)
			}

		case DRIVE_LEFT:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_LEFT, time.Duration(dur)); err != nil {
				glog.Errorf("Failed to move motor %v", err)
			}

		case DRIVE_RIGHT:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_RIGHT, time.Duration(dur)); err != nil {
				glog.Errorf("Failed to move motor %v", err)
			}

		case SERVO_STEP:
			s.servoStep = int(msg.Data.(float64))

		case SERVO_UP:
			if err := s.dev.Servo.SetAngle(s.servoAngle - s.servoStep); err != nil {
				sendError(err.Error(), c)
				continue
			}
			s.servoAngle -= s.servoStep

		case SERVO_DOWN:
			if err := s.dev.Servo.SetAngle(s.servoAngle + s.servoStep); err != nil {
				sendError(err.Error(), c)
				continue
			}
			s.servoAngle += s.servoStep

		case AUDIO_START:
		}

	}
}

// sendError sends an error packet to the browser.
func sendError(errorString string, c *websocket.Conn) {
	msg := ControlMsg{
		CmdType: ERR,
		Data:    errorString,
	}

	jsMsg, err := json.Marshal(msg)
	if err != nil {
		glog.Errorf("Failed to unmarshall: %v", err)
	}

	err = c.WriteMessage(websocket.TextMessage, jsMsg)
	if err != nil {
		glog.Errorf("Failed to write: %v", err)
	}
}

// audioSock handles audio playback from browser.
func (s *Server) audioSock(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Warningf("failed to upgrade conn:%v", err)
		return
	}

	defer c.Close()

	for {

		mt, data, err := c.ReadMessage()
		if err != nil {
			glog.Errorf("Websocket read error: %v", err)
			return
		}
		if mt != 2 {
			glog.Errorf("Audio packet should be binary. Instead got text message type.")
			return
		}

		b := bytes.NewBuffer(data)
		s.audio.Out <- *b

	}
}
