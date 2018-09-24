#!/usr/bin/env bash

echo "Start registry"
docker run -d -p 5000:5000 --name registry -v $(pwd)/registry.yaml:/etc/docker/registry/config.yml registry:2
echo "Start webhook"
docker build -t registry-webhook:latest . ; docker run --name=registry-webhook -p 80:8089 registry-webhook:latest

docker kill registry
docker rm registry registry-webhook
