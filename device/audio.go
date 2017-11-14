/// +build ignore

package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/gordonklaus/portaudio"
)

type Audio struct {
	In           chan bytes.Buffer
	Out          chan bytes.Buffer
	streamIn     *portaudio.Stream
	streamOut    *portaudio.Stream
	bufIn        []int16
	bufOut       []int16
	listenStop   chan struct{}
	playbackStop chan struct{}
}

func NewAudio() *Audio {
	return &Audio{
		In:           make(chan bytes.Buffer),
		Out:          make(chan bytes.Buffer, 100),
		listenStop:   make(chan struct{}),
		playbackStop: make(chan struct{}),
	}
}

// Init initializes the audio. If BufLen is set to zero it does not init the corresponding
// inp or out.
// oBufLen is 8 bits while bufOut is 16 bits.
func (s *Audio) Init(inBufLen, oBufLen int, inSampleRate, outSampleRate float64) error {
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("init failed:%v", err)
	}

	// Open Input stream.
	if inBufLen > 0 {
		bufIn := make([]int16, inBufLen)
		in, err := portaudio.OpenDefaultStream(1, 0, inSampleRate, len(bufIn), bufIn)
		if err != nil {
			return fmt.Errorf("failed to open input stream:%v", err)
		}

		s.streamIn = in
		s.bufIn = bufIn
	}

	// Open Output stream.
	if oBufLen > 0 {
		bufOut := make([]int16, oBufLen)
		out, err := portaudio.OpenDefaultStream(0, 1, outSampleRate, len(bufOut), bufOut)
		if err != nil {
			return fmt.Errorf("failed to open output stream:%v", err)
		}

		s.streamOut = out
		s.bufOut = bufOut
	}
	return nil
}

func (s *Audio) StartPlayback() {
	go s.playback()
}

func (s *Audio) StopPlayback() {
	s.playbackStop <- struct{}{}
}

func (s *Audio) StartListen() {
	go s.listen()
}

func (s *Audio) StopListen() {
	s.listenStop <- struct{}{}
}

// ResetPlayback resets the output stream (stop, start).
// Some hardware seems to need a reset between playback.
func (s *Audio) ResetPlayback() {
	s.streamOut.Abort()
	time.Sleep(1000 * time.Millisecond)
	s.streamOut.Start()
}

func (s *Audio) listen() {
	glog.Info("Started capturing audio from mic")

	if err := s.streamIn.Start(); err != nil {
		glog.Fatalf("Failed to start input stream: %v ", err)
	}

	listenFunc := func() {
		if err := s.streamIn.Read(); err != nil {
			glog.Errorf("Failed to read input stream: %v", err)
		}

		var bufWriter bytes.Buffer
		binary.Write(&bufWriter, binary.LittleEndian, s.bufIn)
		s.In <- bufWriter
		glog.V(2).Infof("Recorded audio chunk size: %v", bufWriter.Len())
	}

	for {
		select {
		case <-s.listenStop:
			if err := s.streamIn.Stop(); err != nil {
				glog.Errorf("Failed to stop input audio stream: %v", err)
			}
			glog.Info("Stopped capturing audio from mic")
			return

		default:
			listenFunc()
		}
	}
}

func (s *Audio) playback() {

	glog.Info("Started playback audio from browser")

	if err := s.streamOut.Start(); err != nil {
		glog.Fatalf("Failed to start audio out: %v", err)
	}

	for {
		select {
		case <-s.playbackStop:
			if err := s.streamOut.Abort(); err != nil {
				glog.Errorf("Failed to stop output audio stream: %v", err)
			}
			glog.Info("Stopped playback audio from browser")
			return

		case out := <-s.Out:
			glog.V(2).Infof("Playback audio chunk size: %v", out.Len())
			if err := binary.Read(&out, binary.LittleEndian, s.bufOut); err != nil {
				glog.Warningf("Failed to convert to binary %v", err)
				continue
			}
			if err := s.streamOut.Write(); err != nil {
				glog.Warningf("Failed to write to audio out: %v", err)
			}
		}
	}
}

func (s *Audio) Quit() {
	if err := s.streamOut.Close(); err != nil {
		glog.Errorf("Failed to close output audio stream: %v", err)
	}
	if err := s.streamIn.Close(); err != nil {
		glog.Errorf("Failed to close input audio stream: %v", err)
	}
}
