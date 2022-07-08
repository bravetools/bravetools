#!/bin/bash
set -e

cd python-multi-service

echo ">> Building and deploying with compose ..."
brave compose

echo ">> Showing units and images ..."
brave units
brave images

echo ">> Deleting units and images ..."
brave remove brave-test-api
brave remove brave-test-auth
brave remove brave-test-log

brave remove -i brave-test-api-1.0
brave remove -i brave-test-auth-1.0
brave remove -i brave-test-log-1.0

echo ">> Showing units and images ..."
brave units
brave images

cd ..
