package device

import (
	"time"

	"gobot.io/x/gobot/drivers/gpio"
)

type Ubiquity struct {
	motorRightFwd *gpio.DirectPinDriver
	motorRightBwd *gpio.DirectPinDriver
	motorLeftFwd  *gpio.DirectPinDriver
	motorLeftBwd  *gpio.DirectPinDriver
}

func New(
	mRF *gpio.DirectPinDriver,
	mRB *gpio.DirectPinDriver,
	mLF *gpio.DirectPinDriver,
	mLB *gpio.DirectPinDriver,
) *Ubiquity {
	return &Ubiquity{
		motorRightFwd: mRF,
		motorRightBwd: mRB,
		motorLeftFwd:  mLF,
		motorLeftBwd:  mLB,
	}
}

func (s *Ubiquity) AllMotorStop() error {
	if err := s.motorLeftBwd.DigitalWrite(0); err != nil {
		return err
	}
	if err := s.motorLeftFwd.DigitalWrite(0); err != nil {
		return err
	}

	if err := s.motorRightBwd.DigitalWrite(0); err != nil {
		return err
	}

	if err := s.motorRightFwd.DigitalWrite(0); err != nil {
		return err
	}
	return nil
}

// Move moves the rover for dur milliseconds in a specific direction.
// dir 0 = fwd, 1 = bwd, 2 = left, 3 = right
func (s *Ubiquity) MotorControl(dir int, dur time.Duration) error {

	switch dir {
	case 0:
		s.motorRightFwd.DigitalWrite(1)
		s.motorLeftFwd.DigitalWrite(1)
	case 1:
		s.motorRightBwd.DigitalWrite(1)
		s.motorLeftBwd.DigitalWrite(1)
	case 2:
		s.motorLeftFwd.DigitalWrite(1)
		s.motorRightBwd.DigitalWrite(1)
	case 3:
		s.motorRightFwd.DigitalWrite(1)
		s.motorLeftBwd.DigitalWrite(1)
	}

	time.Sleep(dur * time.Millisecond)
	return s.AllMotorStop()
}
