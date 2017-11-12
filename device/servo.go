package device

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang/glog"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/sysfs"
)

type Servo struct {
	pin       string
	pwmPeriod uint32 // PWM period in ms.
	adaptor   gobot.Adaptor
}

func NewServo(piBlasterPeriod uint32, pin string, a gobot.Adaptor) *Servo {
	return &Servo{
		pin:       pin,
		pwmPeriod: piBlasterPeriod,
		adaptor:   a,
	}
}

// Unexport unexports the pin and releases the pin from the operating system
func (p *Servo) Unexport() error {
	return p.piBlaster(fmt.Sprintf("release %v\n", p.pin))
}

// SetAngle moves the servo to the appropriate angle.
func (p *Servo) SetAngle(angle int) error {
	if angle < 0 || angle > 180 {
		return fmt.Errorf("Angle needs to be 0 to 180, got %v", angle)
	}

	val := 500 + 1500*angle/180

	glog.V(2).Infof("Setting angle:%v -> duty cycle:%v microsecs ", angle, val)
	p.SetDutyCycle(uint32(val))
	return nil
}

// SetDutyCycle sets the duty cycle of the PWM.
func (p *Servo) SetDutyCycle(duty uint32) error {
	if duty > 20000 { //p.pwmPeriod {
		return errors.New("Duty cycle exceeds period.")
	}

	val := gobot.FromScale(float64(duty), 0, float64(p.pwmPeriod))

	glog.V(2).Infof("Setting PWM duty cycle:%v, period:%v, piBlasterDuty:%v pin:%v", duty, p.pwmPeriod, val, p.pin)
	return nil
	//TODO:remove comment	return p.piBlaster(fmt.Sprintf("%v=%v\n", p.pin, val))
}

func (p *Servo) piBlaster(data string) (err error) {
	fi, err := sysfs.OpenFile("/dev/pi-blaster", os.O_WRONLY|os.O_APPEND, 0644)
	defer fi.Close()

	if err != nil {
		return err
	}

	_, err = fi.WriteString(data)
	return
}
