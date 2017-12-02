/// +build ignore

package device

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/golang/glog"
	"github.com/gordonklaus/portaudio"
)

type Audio struct {
	In             chan bytes.Buffer
	Out            chan bytes.Buffer
	playSampleRate float64
	recSampleRate  float64
	recBuf         []int16
	playBuf        []int16
	recStop        chan struct{}
	playStop       chan struct{}
	playStatus     bool // True is currently in playback loop.
	recStatus      bool // True is current in record loop.
	playStream     *portaudio.Stream
	recStream      *portaudio.Stream
}

func NewAudio() *Audio {
	return &Audio{
		In:         make(chan bytes.Buffer),
		Out:        make(chan bytes.Buffer),
		recStop:    make(chan struct{}),
		playStop:   make(chan struct{}),
		playStatus: false,
		recStatus:  false,
	}
}

// Init initializes the audio.
// oBufLen is 8 bits while bufOut is 16 bits.
func (s *Audio) Init(recBufLen, playBufLen int, recSampleRate, playSampleRate float64) error {
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("init failed:%v", err)
	}

	s.playSampleRate = playSampleRate
	s.recSampleRate = recSampleRate

	buf := make([]int16, playBufLen)
	s.playBuf = buf

	buf = make([]int16, recBufLen)
	s.recBuf = buf

	stream, err := portaudio.OpenDefaultStream(0, 1, s.playSampleRate, len(s.playBuf), s.playBuf)
	if err != nil {
		glog.Errorf("failed to open output stream:%v", err)
		return err
	}
	s.playStream = stream

	stream, err = portaudio.OpenDefaultStream(1, 0, s.recSampleRate, len(s.recBuf), s.recBuf)
	if err != nil {
		glog.Errorf("Failed to open input stream:%v", err)
		return err
	}
	s.recStream = stream

	return nil
}

func (s *Audio) Close() {
	if err := s.playStream.Close(); err != nil {
		glog.Errorf("Failed to close output audio stream: %v", err)
	}
	if err := s.recStream.Close(); err != nil {
		glog.Errorf("Failed to close output audio stream: %v", err)
	}
	if err := portaudio.Terminate(); err != nil {
		glog.Errorf("Failed to terminate portaudio: %v", err)
	}
}

// IsRec returns true if currently recording.
func (s *Audio) IsRec() bool {
	return s.recStatus
}

func (s *Audio) StartPlayback() {
	if s.playStatus {
		return
	}
	go s.playback()
}

func (s *Audio) StopPlayback() {
	if !s.playStatus {
		return
	}
	s.playStop <- struct{}{}

}

func (s *Audio) StartRec() {
	if s.recStatus {
		return
	}
	go s.rec()
}

func (s *Audio) StopRec() {
	if !s.recStatus {
		return
	}
	s.recStop <- struct{}{}
}

func (s *Audio) rec() {

	if err := s.recStream.Start(); err != nil {
		glog.Errorf("Failed to start input stream: %v ", err)
		return
	}

	glog.Info("Started capturing audio from mic")
	s.recStatus = true

	for {
		select {
		case <-s.recStop:
			if err := s.recStream.Stop(); err != nil {
				glog.Errorf("Failed to stop input audio stream: %v", err)
			}
			glog.Info("Stopped capturing audio from mic")
			s.recStatus = false
			return

		default:
			if err := s.recStream.Read(); err != nil {
				glog.Errorf("Failed to read input stream: %v", err)
			}
			var bufWriter bytes.Buffer
			binary.Write(&bufWriter, binary.LittleEndian, s.recBuf)
			s.In <- bufWriter
			glog.V(2).Infof("Recorded audio chunk size: %v", bufWriter.Len())
		}
	}
}

func (s *Audio) playback() {

	if err := s.playStream.Start(); err != nil {
		glog.Errorf("Failed to start audio out: %v", err)
		return
	}

	glog.Info("Started playback audio from browser")
	s.playStatus = true

	for {
		select {
		case <-s.playStop:
			if err := s.playStream.Stop(); err != nil {
				glog.Errorf("Failed to stop output audio stream: %v", err)
			}
			glog.Info("Stopped playback audio from browser")
			s.playStatus = false
			return

		case out := <-s.Out:
			glog.V(2).Infof("Playback audio chunk size: %v", out.Len())
			if err := binary.Read(&out, binary.LittleEndian, s.playBuf); err != nil {
				glog.Warningf("Failed to convert to binary %v", err)
				continue
			}
			if err := s.playStream.Write(); err != nil {
				glog.Errorf("Failed to write to audio out: %v", err)
			}
		}
	}

}
