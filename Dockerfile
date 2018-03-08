FROM google/debian:wheezy
MAINTAINER Luis Mora Medina <luismoramedina@gmail.com>

RUN apt-get update
RUN apt-get install -y iptables curl procps adduser

ADD gomezh gomezh
ADD docker/start.sh start.sh
RUN chmod +x start.sh
EXPOSE 8080

ENTRYPOINT ["/start.sh"]