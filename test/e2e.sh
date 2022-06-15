#!/bin/bash
set -e

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
brave units
sleep 10

echo ">> Stopping a unit ..."
brave stop go-service
brave units
sleep 10

echo ">> Deleteing a unit ..."
brave remove go-service
brave units
brave images
brave info