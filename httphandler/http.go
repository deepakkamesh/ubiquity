package httphandler

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"time"

	"github.com/deepakkamesh/ubiquity/device"
	"github.com/golang/glog"
	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

// Message Types.
const (
	ERR = iota
	CMD
	AUDIO
)

// Command Types.
const (
	FWD = iota
	BWD
	LEFT
	RIGHT
)

// Message.
type Message struct {
	MsgType int
	Data    interface{}
}

type Server struct {
	connCount int
	dev       *device.Ubiquity
}

func New(dev *device.Ubiquity) *Server {

	return &Server{
		dev: dev,
	}
}

func (s *Server) Start(hostPort string, resPath string) error {

	// http routers.
	http.HandleFunc("/datastream", s.controlSocket)

	// Serve static content from resources dir.
	fs := http.FileServer(http.Dir(resPath))
	http.Handle("/", fs)

	//return http.ListenAndServe(hostPort, nil)
	return http.ListenAndServeTLS(hostPort, resPath+"/server.crt", resPath+"/server.key", nil)
}

// controlSocket is the websocket server that streams rover stats.
func (s *Server) controlSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Warningf("failed to upgrade conn:%v", err)
		return
	}

	s.connCount++

	defer func() {
		c.Close()
		s.connCount--
	}()

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

	for {

		mt, data, err := c.ReadMessage()
		if err != nil {
			glog.Errorf("Websocket read error: %v", err)
			return
		}
		//var msg Message
		//json.Unmarshal(data, &msg)

		b := bytes.NewReader(data)
		err = binary.Read(b, binary.LittleEndian, &bufOut)
		if err != nil {
			glog.Errorf("%v", err)
		}
		glog.Infof("Got message type %v payload %v", mt, len(bufOut))

		if err := out.Write(); err != nil {
			glog.Warningf("Failed to write to audio out: %v", err)
		}

		/*
			switch msg.MsgType {
			case CMD:
				s.execute(msg.Data)

			case AUDIO:
				s.play(msg.Data)
			} */

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
func (s *Server) execute(c interface{}) {
	cmd := c.(map[string]interface{})
	dir := cmd["CmdType"].(float64)
	dur := cmd["Param"].(float64)

	if err := s.dev.MotorControl(int(dir), time.Duration(dur)); err != nil {
		glog.Errorf("Failed to move motor %v", err)
	}
}

func (s *Server) play(d interface{}) {

	data := d.([]float32)
	glog.Infof("Got audio chunk size %v", len(data))
}
