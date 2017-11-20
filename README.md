# ubiquity
Mini Rover that currently streams audio and video on a controllable platform.

## Hardware Setup
### Raspberry PI Zero W Setup

#### Audio Setup.
Follow instructions on [blog](https://www.tinkernut.com/2017/04/adding-audio-output-raspberry-pi-zero-tinkernut-workbench/) to send audio via PWM.

*TL;DR*
* echo "overlay=pwm-2chan,pin=18,func=2,pin2=13,func2=4" >> /boot/config.txt
* or echo for one channel  "overlay=pwm,pin=18,func=2" >> /boot/config.txt
* Force audio via 3.5mm jack in raspi-config
* Build a low pass filter as shown in the above link.
* Test by "aplay /usr/share/sounds/alsa/Front_Center.wav"
 

#### PWM setup.
* Setup [pi-blaster](https://github.com/sarfata/pi-blaster) for PWM support if there is a servo mount.

#### References
* [webcam lib](https://github.com/blackjack/webcam)


#### TODO
* Restart process from web ui

#### Current Profile
* H bridge 30ma (rest)
* Servo - negligible (rest)
* Speaker - 10ma (rest)
* Display - negligible (rest)
* All devices powered on - 170ma
* Running binary - 220ma (total)
* Streaming video - 260ma
* Audio Streams - 300ma

