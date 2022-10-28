#! /bin/bash

apt-get -y update && apt-get -y install docker.io

sudo git clone https://github.com/DMW2151/go-failure.git

cd ./go-failure &&\
	docker build . -f ./examples/client/Dockerfile -t dmw2151/phi-failure-client

DO_REGION=`(curl -s http://169.254.169.254/metadata/v1/region)`

docker run --name phi-failure-client \
	-e FAILURE_DETECTOR_SERVER_HOST=${FAILURE_DETECTOR_SERVER_HOST}\
	-e FAILURE_DETECTOR_SERVER_PORT="52151"\
	-e DIGITALOCEAN_REGION=$DO_REGION \
	dmw2151/phi-failure-client ./client