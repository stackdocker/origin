#!/bin/bash
#/usr/lib/coreos/dockerd daemon --host=fd:// $DOCKER_OPTS $DOCKER_OPT_BIP $DOCKER_OPT_MTU $DOCKER_OPT_IPMASQ -/run/flannel_docker_opts.env
sudo cp docker /etc/sysconfig/
#sudo docker daemon --host=fd:// --bip=172.17.0.1/22 --mtu=1472 --ipmasq=true -/run/flannel_docker_opts.env
sudo systemctl restart docker.service
sudo docker network create --gateway=10.3.0.1 --subnet=10.3.0.0/24 cbr0
