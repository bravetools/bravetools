#!/bin/bash

BRAVETOOLS_HOME=$HOME"/.bravetools"

if  which lxc >/dev/null 2>&1; then
    lxc profile delete brave
    sudo snap remove lxd
fi

if [ -d "$BRAVETOOLS_HOME" ]; then rm -Rf $BRAVETOOLS_HOME; fi

echo ">> Installing Bravetools"
cd ..
make ubuntu
brave version
sleep 10

echo ">> Initialising host with default settings ..."
brave init
brave info
sleep 10

echo ">> Building and deploying a unit ..."
cd examples/go-service
brave build
brave deploy
brave unitssleep 10

echo ">> Stopping a unit ..."
brave stop go-service
brave units
sleep 10

echo ">> Deleteing a unit ..."
brave remove go-service
brave units
brave images
brave info