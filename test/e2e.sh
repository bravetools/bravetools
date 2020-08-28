#!/bin/bash

BRAVETOOLS_HOME=$HOME"/.bravetools"

if  which lxc >/dev/null 2>&1; then
    sudo snap remove lxd
fi

if [ -d "$BRAVETOOLS_HOME" ]; then rm -Rf $BRAVETOOLS_HOME; fi

cd ..
make ubuntu
brave version

echo ">>>"
echo "Bravetools installed"
echo ">>>"
echo "Initialising host with default settings ..."
echo ""
brave init

echo ">>>"
echo "Bravehost is runnnig"
echo ">>>"
brave info

echo ">>>"
echo "LXC parameters"
echo ">>>"
lxc profile show brave
lxc storage list

# echo ">>>"
# echo "Building and deploying a unit"
# echo ">>>"
# cd examples/go-service
# brave base ubuntu/bionic
# brave build
# brave deploy

