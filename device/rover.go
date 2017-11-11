package device

import (
	"fmt"
	"time"

	"gobot.io/x/gobot/drivers/gpio"
)

const (
	DRIVE_FWD = iota
	DRIVE_BWD
	DRIVE_LEFT
	DRIVE_RIGHT
)

type Ubiquity struct {
	motorRightFwd *gpio.DirectPinDriver
	motorRightBwd *gpio.DirectPinDriver
	motorLeftFwd  *gpio.DirectPinDriver
	motorLeftBwd  *gpio.DirectPinDriver
	Servo         *Servo
}

// Return a New initializaed ubiquity device.
func New(
	mRF *gpio.DirectPinDriver,
	mRB *gpio.DirectPinDriver,
	mLF *gpio.DirectPinDriver,
	mLB *gpio.DirectPinDriver,
	servo *Servo,
) *Ubiquity {
	return &Ubiquity{
		motorRightFwd: mRF,
		motorRightBwd: mRB,
		motorLeftFwd:  mLF,
		motorLeftBwd:  mLB,
		Servo:         servo,
	}
}

// AllMotorStop stops all motors.
func (s *Ubiquity) AllMotorStop() error {
	if s.motorRightFwd == nil || s.motorRightBwd == nil ||
		s.motorLeftFwd == nil || s.motorLeftBwd == nil {
		return fmt.Errorf("motors not initialized")
	}

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

	if s.motorRightFwd == nil || s.motorRightBwd == nil ||
		s.motorLeftFwd == nil || s.motorLeftBwd == nil {
		return fmt.Errorf("motors not initialized")
	}

	switch dir {
	case DRIVE_FWD:
		if err := s.motorRightFwd.DigitalWrite(1); err != nil {
			return err
		}
		if err := s.motorLeftFwd.DigitalWrite(1); err != nil {
			return err
		}

	case DRIVE_BWD:
		if err := s.motorRightBwd.DigitalWrite(1); err != nil {
			return err
		}
		if err := s.motorLeftBwd.DigitalWrite(1); err != nil {
			return err
		}

	case DRIVE_LEFT:
		if err := s.motorLeftFwd.DigitalWrite(1); err != nil {
			return err
		}
		if err := s.motorRightBwd.DigitalWrite(1); err != nil {
			return err
		}

	case DRIVE_RIGHT:
		if err := s.motorRightFwd.DigitalWrite(1); err != nil {
			return err
		}
		if err := s.motorLeftBwd.DigitalWrite(1); err != nil {
			return err
		}
	}

	time.Sleep(dur * time.Millisecond)
	return s.AllMotorStop()
}
