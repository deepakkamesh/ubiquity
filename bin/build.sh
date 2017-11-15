#!/bin/bash
# $1 = arm or linux
# $2 binary to build: main or cli.
# $3 host to push binary to. eg. 10.0.0.20

LOC=$(dirname "$0")
HOSTNAME=$(hostname)
if [ $HOSTNAME == "dkg-macbookpro.roam.corp.google.com" ]; then
	PROJECT_ROOT="/Users/dkg/Projects/"
else
	PROJECT_ROOT="/home/dkg/Projects/"
fi
GOROOT="/usr/local/go"
export GOPATH="$PROJECT_ROOT/golang"
BUILDTIME="`date '+%Y-%m-%d_%I:%M:%S%p'`"
GITHASH="`git -C $LOC rev-parse --short=7 HEAD`"
VER="-X main.buildtime=$BUILDTIME -X main.githash=$GITHASH"

if [ $# -lt 1 ]; then
	echo "build.sh < arm | noarm | src> < ip address > <all | res | bin >"
	exit
fi

if [ "$1" == "src" ]; then
  	echo "Pushing source to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress ../  $2:~/Projects/ubiquity
		exit 0
fi

# Path to Raspberry pi compile tools from https://github.com/raspberrypi/tools
CCHOME=/home/dkg/tools/arm-bcm2708/arm-rpi-4.9.3-linux-gnueabihf/bin

# Compile binary if not res only.
if [ "$3" != "res" ]; then 
	if [ $1 == "arm" ]; then
		echo "Compiling for ARM $BUILDTIME $GITHASH"
   GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 CC="$CCHOME/arm-linux-gnueabihf-gcc" \
   CGO_LDFLAGS="-L/home/dkg/arm-lib -Wl,-rpath,/home/dkg/arm-lib" \
   PKG_CONFIG_PATH=/home/dkg/arm-lib \
   $GOROOT/bin/go build -ldflags "$VER" -o $LOC/main $LOC/../main.go

	else
		echo "Compiling on local machine $BUILDTIME $GITHASH"
	PKG_CONFIG_PATH=/usr/local/Cellar/portaudio/19.6.0/lib/pkgconfig/ \
	go build  -ldflags "$VER" ../main.go
	fi
fi

# Push binary to remote if previous step completed.
if ! [ -z "$2" ]; then
 	if [ $3 == "all" ]; then
  	echo "Pushing binary to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress main $2:~/ubiquity
  	echo "Pushing resources to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress ../resources $2:~/ubiquity
	elif [ $3 == "res" ]; then
  	echo "Pushing resources to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress ../resources $2:~/ubiquity
	elif [ $3 == "bin" ]; then
  	echo "Pushing binary to machine $2"
		rsync -avz -e "ssh -o StrictHostKeyChecking=no" --progress main $2:~/ubiquity
	fi
fi

