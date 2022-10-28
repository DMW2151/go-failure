#! /bin/bash

apt-get -y update && apt-get -y install docker.io docker-compose 

git clone https://github.com/DMW2151/go-failure.git

cd ./go-failure && docker build . -f ./examples/server/Dockerfile -t dmw2151/phi-failure-server

cd ./demo/ && docker-compose up