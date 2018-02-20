FROM google/debian:wheezy
MAINTAINER Luis Mora Medina <luismoramedina@gmail.com>

ADD gomesh gomesh
EXPOSE 8080
EXPOSE 8082
ENTRYPOINT ["/gomesh"]