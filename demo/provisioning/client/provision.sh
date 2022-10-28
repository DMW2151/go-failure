#! /bin/bash

apt-get -y update && apt-get -y install docker.io

git clone https://github.com/DMW2151/go-failure.git

cd ./go-failure &&\
	docker build . -f ./examples/client/Dockerfile -t dmw2151/phi-failure-client &&\

cd ./demo/ && docker-compose up