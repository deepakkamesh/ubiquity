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
)

// Control Message.
type ControlMsg struct {
	CmdType int
	Data    interface{}
}

type Server struct {
	connCount int // number of connected http clients.
	dev       *device.Ubiquity
	audio     *device.Audio
}

func New(dev *device.Ubiquity, aud *device.Audio) *Server {
	return &Server{
		dev:   dev,
		audio: aud,
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

		case AUDIO_START:
			//s.play(msg.Data)
		}

		/*
			jsMsg, err := json.Marshal(m.data)
			if err != nil {
				glog.Errorf("Failed to unmarshall: %v", err)
				continue
			}
			m.data.Err = ""

			err = c.WriteMessage(websocket.TextMessage, jsMsg)
			if err != nil {
				glog.Errorf("Failed to write: %v", err)
				retur1
			}
		*/
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
	/*
		// Setup playback.
		if err := portaudio.Initialize(); err != nil {
			glog.Fatalf("Failed to start audio out: %v", err)
		}

		bufOut := make([]int16, 743)
		out, err := portaudio.OpenDefaultStream(0, 1, 4000, len(bufOut), bufOut)
		if err != nil {
			glog.Fatalf("Failed to start audio out: %v", err)
		}
		if err := out.Start(); err != nil {
			glog.Fatalf("Failed to start audio out: %v", err)
		}
	*/
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

		/*		err = binary.Read(b, binary.LittleEndian, &bufOut)
				if err != nil {
					glog.Errorf("%v", err)
				}
				glog.V(2).Infof("Got audio packet of sz:%v", len(bufOut))

				if err := out.Write(); err != nil {
					glog.Warningf("Failed to write to audio out: %v", err)
				}*/

	}
}
