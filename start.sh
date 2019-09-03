#!/bin/sh
# This is a comment!
echo Starting script

docker rm -f $(docker ps -aq)
sudo rm -rf build
make build-linux
make build-docker-colordnode
make localnet-start

