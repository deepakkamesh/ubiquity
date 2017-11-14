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
}

func NewAudio() *Audio {
	return &Audio{
		In:       make(chan bytes.Buffer),
		Out:      make(chan bytes.Buffer, 100),
		recStop:  make(chan struct{}),
		playStop: make(chan struct{}),
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

	return nil
}

func (s *Audio) StartPlayback() {
	go s.playback()
}

func (s *Audio) StopPlayback() {
	s.playStop <- struct{}{}
}

func (s *Audio) StartRec() {
	go s.rec()
}

func (s *Audio) StopRec() {
	s.recStop <- struct{}{}
}

func (s *Audio) rec() {
	glog.Info("Started capturing audio from mic")

	stream, err := portaudio.OpenDefaultStream(1, 0, s.recSampleRate, len(s.recBuf), s.recBuf)
	if err != nil {
		glog.Errorf("failed to open input stream:%v", err)
	}

	if err := stream.Start(); err != nil {
		glog.Fatalf("Failed to start input stream: %v ", err)
	}

	for {
		select {
		case <-s.recStop:
			if err := stream.Stop(); err != nil {
				glog.Errorf("Failed to stop input audio stream: %v", err)
			}
			if err := stream.Close(); err != nil {
				glog.Errorf("Failed to close output audio stream: %v", err)
			}
			glog.Info("Stopped capturing audio from mic")
			return

		default:
			if err := stream.Read(); err != nil {
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

	glog.Info("Started playback audio from browser")

	stream, err := portaudio.OpenDefaultStream(0, 1, s.playSampleRate, len(s.playBuf), s.playBuf)
	if err != nil {
		glog.Errorf("failed to open output stream:%v", err)
	}

	if err := stream.Start(); err != nil {
		glog.Fatalf("Failed to start audio out: %v", err)
	}

	for {
		select {
		case <-s.playStop:
			if err := stream.Stop(); err != nil {
				glog.Errorf("Failed to stop output audio stream: %v", err)
			}
			if err := stream.Close(); err != nil {
				glog.Errorf("Failed to close output audio stream: %v", err)
			}
			glog.Info("Stopped playback audio from browser")
			return

		case out := <-s.Out:
			glog.V(2).Infof("Playback audio chunk size: %v", out.Len())
			if err := binary.Read(&out, binary.LittleEndian, s.playBuf); err != nil {
				glog.Warningf("Failed to convert to binary %v", err)
				continue
			}
			if err := stream.Write(); err != nil {
				glog.Warningf("Failed to write to audio out: %v", err)
			}
		}
	}

}
