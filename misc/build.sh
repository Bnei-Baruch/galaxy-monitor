#!/usr/bin/env bash
# run misc/build.sh from project root
set -e
set -x

docker image build -t galaxy-monitor:latest .
docker create --name dummy galaxy-monitor:latest
docker cp dummy:/app/galaxy-monitor ./galaxy-monitor
docker rm -f dummy
