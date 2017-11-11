package device

import (
	"bytes"
	"encoding/binary"
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
		In:           make(chan bytes.Buffer, 10),
		Out:          make(chan bytes.Buffer, 1000),
		listenStop:   make(chan struct{}),
		playbackStop: make(chan struct{}),
	}
}

func (s *Audio) Init(inBufLen, oBufLen int, sampleRate float64) error {
	if err := portaudio.Initialize(); err != nil {
		return err
	}

	// Open Input stream.
	bufIn := make([]int16, inBufLen)
	in, err := portaudio.OpenDefaultStream(1, 0, sampleRate, len(bufIn), bufIn)
	if err != nil {
		return err
	}

	s.streamIn = in
	s.bufIn = bufIn

	// Open Output stream.
	bufOut := make([]int16, oBufLen)
	out, err := portaudio.OpenDefaultStream(0, 1, sampleRate, len(bufOut), bufOut)
	if err != nil {
		return err
	}

	s.streamOut = out
	s.bufOut = bufOut

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
	s.StopPlayback()
	time.Sleep(50 * time.Millisecond)
	s.StartPlayback()
}

func (s *Audio) listen() {

	// TODO: Get a cleaner solution to by removing the buffered channel and
	// handling the race condition. Write to s.In in a go func.
	if len(s.In) > 0 {
		glog.Warningf("Audio input channel is non zero: %v", len(s.In))
	}
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
	}

	for {
		select {
		case <-s.listenStop:
			if err := s.streamIn.Stop(); err != nil {
				glog.Errorf("Failed to stop input audio stream: %v", err)
			}
			return

		default:
			listenFunc()
		}
	}
}

func (s *Audio) playback() {

	if err := s.streamOut.Start(); err != nil {
		glog.Fatalf("Failed to start audio out: %v", err)
	}

	for {
		select {
		case <-s.playbackStop:
			if err := s.streamOut.Stop(); err != nil {
				glog.Errorf("Failed to stop output audio stream: %v", err)
			}
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
