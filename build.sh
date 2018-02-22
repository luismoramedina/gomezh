#/bin/bash

go build
docker build . -t gomesh
docker tag gomesh luismoramedina/gomesh
docker push luismoramedina/gomesh

cd ./samples/books
mvn clean package
docker build . -t books
docker tag books luismoramedina/books
docker push luismoramedina/books


cd ./samples/stars/
mvn clean package
docker build . -t stars
docker tag stars luismoramedina/stars
docker push luismoramedina/stars

cd ..