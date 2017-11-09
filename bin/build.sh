#!/bin/bash
# $1 = arm or linux
# $2 binary to build: main or cli.
# $3 host to push binary to. eg. 10.0.0.20

BUILDTIME="`date '+%Y-%m-%d_%I:%M:%S%p'`"
GITHASH="`git rev-parse --short=7 HEAD`"
VER="-X main.buildtime=$BUILDTIME -X main.githash=$GITHASH"

if [ $# -lt 2 ]; then
#if [ "$1" == "help" ]; then
	echo "build.sh < arm | noarm > < ip address > <all | res | bin >"
	exit
fi


# Compile binary if not res only.
if [ "$3" != "res" ]; then 
	if [ $1 == "arm" ]; then
		echo "Compiling for ARM $BUILDTIME $GITHASH"
		GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -ldflags "$VER" ../main.go 
	else
		echo "Compiling on local machine $BUILDTIME $GITHASH"
		go build -ldflags "$VER" $BINARY
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

