package httphandler

import (
	"bytes"
	"encoding/json"
	"net/http"

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
	VIDEO_ENABLE
	VIDEO_DISABLE
	AUDIO_ENABLE
	AUDIO_DISABLE
	MASTER_ENABLE
	MASTER_DISABLE
)

// Control Message.
type ControlMsg struct {
	CmdType int
	Data    interface{}
}

type Server struct {
	dev   *device.Ubiquity
	audio *device.Audio
	video *device.Video

	connCount  int // number of connected http clients.
	servoStep  int // Servo step for each click.
	servoAngle int // Current Angle for servo.

	pauseRec bool
}

func New(dev *device.Ubiquity, aud *device.Audio, vid *device.Video) *Server {
	return &Server{
		dev:        dev,
		audio:      aud,
		video:      vid,
		servoAngle: 90,
		servoStep:  30,
		pauseRec:   false,
	}
}

func (s *Server) Start(hostPort string, resPath string, cert string, privkey string) error {

	// http routers.
	http.HandleFunc("/audiostream", s.audioSock)
	http.HandleFunc("/control", s.controlSock)
	if s.video != nil {
		http.Handle("/videostream", s.video.Stream)
	}

	// Serve static content from resources dir.
	fs := http.FileServer(http.Dir(resPath))
	http.Handle("/", fs)

	//return http.ListenAndServeTLS(hostPort, resPath+"/server.crt", resPath+"/server.key", nil)
	return http.ListenAndServeTLS(hostPort, resPath+"/"+cert, resPath+"/"+privkey, nil)
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
			glog.Errorf("Control websocket read error: %v", err)
			return
		}
		var msg ControlMsg
		json.Unmarshal(data, &msg)
		glog.V(2).Infof("Got control message type payload:%v", msg)

		switch msg.CmdType {
		case DRIVE_FWD:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_FWD, int(dur)); err != nil {
				glog.Errorf("Failed to move motor: %v", err)
				sendError(err.Error(), c)
			}

		case DRIVE_BWD:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_BWD, int(dur)); err != nil {
				glog.Errorf("Failed to move motor: %v", err)
				sendError(err.Error(), c)
			}

		case DRIVE_LEFT:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_LEFT, int(dur)); err != nil {
				glog.Errorf("Failed to move motor: %v", err)
				sendError(err.Error(), c)
			}

		case DRIVE_RIGHT:
			dur := msg.Data.(float64)
			if err := s.dev.MotorControl(device.DRIVE_RIGHT, int(dur)); err != nil {
				glog.Errorf("Failed to move motor: %v", err)
				sendError(err.Error(), c)
			}

		case SERVO_STEP:
			s.servoStep = int(msg.Data.(float64))

		case SERVO_UP:
			if err := s.dev.Servo.SetAngle(s.servoAngle - s.servoStep); err != nil {
				glog.Errorf("Failed to move servo: %v", err)
				sendError(err.Error(), c)
				continue
			}
			s.servoAngle -= s.servoStep

		case SERVO_DOWN:
			if err := s.dev.Servo.SetAngle(s.servoAngle + s.servoStep); err != nil {
				glog.Errorf("Failed to move servo: %v", err)
				sendError(err.Error(), c)
				continue
			}
			s.servoAngle += s.servoStep

		case AUDIO_START:
			if s.audio.IsRec() {
				s.pauseRec = true
				s.audio.StopRec()
			}
			s.audio.StartPlayback()

		case AUDIO_STOP:
			s.audio.StopPlayback()
			if s.pauseRec {
				s.pauseRec = false
				s.audio.StartRec()
			}

		case VIDEO_ENABLE:
			fps := msg.Data.(float64)
			s.video.SetFPS(uint(fps))
			s.video.StartVideoStream()

		case VIDEO_DISABLE:
			s.video.StopVideoStream()

		case AUDIO_ENABLE:
			s.audio.StartRec()

		case AUDIO_DISABLE:
			s.audio.StopRec()

		case MASTER_DISABLE:
			if err := s.dev.Lock(false); err != nil {
				glog.Errorf("Failed to lock: %v", err)
				sendError(err.Error(), c)
			}

		case MASTER_ENABLE:
			if err := s.dev.Lock(true); err != nil {
				glog.Errorf("Failed to lock: %v", err)
				sendError(err.Error(), c)
			}

		}

	}
}

// sendError sends an error packet on control socket to the browser.
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
		glog.Errorf("failed to upgrade conn:%v", err)
		return
	}
	defer c.Close()

	// Playback audio from browser.
	go func() {
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				glog.Errorf("Audio websocket read error: %v", err)
				return
			}
			b := bytes.NewBuffer(data)
			s.audio.Out <- *b
		}
	}()

	// Send audio packets to browser.
	for {
		audData := <-s.audio.In
		if err := c.WriteMessage(websocket.BinaryMessage, audData.Bytes()); err != nil {
			glog.Warningf("Websocket write error:%v", err)
			return
		}
	}

}
