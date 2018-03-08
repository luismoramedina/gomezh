#!/bin/bash

# all 8081 input traffic is intercepted and redirected to 8080, except local requests
iptables -t nat -A PREROUTING  ! -d 127.0.0.1 -p tcp --dport 8081 -j REDIRECT --to 8080
# all output traffic in 8081 is intercepted and redirected to 8082, except proxy requests
iptables -t nat -A OUTPUT      -m owner ! --uid-owner 666 -p tcp --dport 8081 -j REDIRECT --to 8082
iptables-save

useradd -m --uid 666 proxyuser
adduser proxyuser sudo
echo "proxyuser ALL=NOPASSWD: ALL" >> /etc/sudoers

su -m proxyuser -c /gomezh