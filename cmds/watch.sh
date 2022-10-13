#!/usr/bin/env bash

docker stop plantuml
docker run -d --rm -p 45678:8080 --name plantuml plantuml/plantuml-server:tomcat

BASEDIR=$(dirname "$0")
echo "$BASEDIR"
cd $BASEDIR
cd ../docs

go install github.com/xh-dev-go/plantuml-watcher@latest

plantuml-watcher -url http://localhost:45678 -dir .