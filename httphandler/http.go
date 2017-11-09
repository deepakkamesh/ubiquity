package httphandler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/deepakkamesh/ubiquity/device"
	"github.com/golang/glog"
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

	// Startup data collection routine.
	return http.ListenAndServe(hostPort, nil)
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

	for {

		mt, data, err := c.ReadMessage()
		if err != nil {
			glog.Errorf("Websocket read error: %v", err)
			return
		}
		var msg Message
		json.Unmarshal(data, &msg)
		glog.Infof("Got message type %v payload %v", mt, msg)

		switch msg.MsgType {
		case CMD:
			s.execute(msg.Data)
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
func (s *Server) execute(c interface{}) {
	cmd := c.(map[string]interface{})
	dir := cmd["CmdType"].(float64)
	dur := cmd["Param"].(float64)

	if err := s.dev.MotorControl(int(dir), time.Duration(dur)); err != nil {
		glog.Errorf("Failed to move motor %v", err)
	}
}
