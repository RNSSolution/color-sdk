#!/bin/bash
# -------------------------------------------------
 

echo "starting color start script"
make localnet-stop

echo "removing build"
sudo rm -rf build

echo "making linux build"
make build-linux

echo "staring localnet"
make localnet-start

docker logs -f colordnode1
