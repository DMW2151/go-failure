#! /bin/bash

apt-get -y update && apt-get -y install docker.io

git clone https://github.com/DMW2151/go-failure.git

cd ./go-failure &&\
	docker build . -f ./examples/server/Dockerfile -t dmw2151/phi-failure-server

docker run --name phi-failure-client \
	-e FAILURE_DETECTOR_SERVER_HOST=${FAILURE_DETECTOR_SERVER_HOST}\
	-e FAILURE_DETECTOR_SERVER_PORT="52151"\
	dmw2151/phi-failure-client ./client
