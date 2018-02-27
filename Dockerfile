FROM google/debian:wheezy
MAINTAINER Luis Mora Medina <luismoramedina@gmail.com>

ADD gomezh gomezh
EXPOSE 8080
EXPOSE 8082
ENTRYPOINT ["/gomezh"]