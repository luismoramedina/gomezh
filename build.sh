#!/bin/bash
set -e

go build
docker build . -t gomezh

cd ./samples/books
mvn clean package
docker build . -t books

cd ../stars/
mvn clean package
docker build . -t stars

cd ../..