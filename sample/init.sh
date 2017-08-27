#!/usr/bin/env bash

PREFIX_INFO="[INFO] ::"

function fetchIP() {
    vagrant ssh $1 -c "ip addr show eth1 | grep 'inet ' | sed -e 's/^.*inet //' -e 's/\/.*$//'" 2> /dev/null
}
vagrant up


echo "$PREFIX_INFO Fetching master and worker ip"

MASTER_IP=`fetchIP "core-01" | tr -d '[:space:]'`
WORKER_IP=`fetchIP "core-02" | tr -d '[:space:]'`

echo "$PREFIX_INFO master ip -> $MASTER_IP"
echo "$PREFIX_INFO worker ip -> $WORKER_IP"

echo "$PREFIX_INFO Creating swarm context"

INIT_SWARM="docker swarm init --advertise-addr $MASTER_IP"

echo "$PREFIX_INFO executing -> $INIT_SWARM"
vagrant ssh core-01 -c "$INIT_SWARM" 2> /dev/null

echo "$PREFIX_INFO Fetching swarm join worker token"
WORKER_TOKEN=`vagrant ssh core-01 -c "docker swarm join-token worker -q" 2> /dev/null`

echo "$PREFIX_INFO Adding worker node (core-02) to swarm cluster"
vagrant ssh core-01 -c "docker swarm join --token $WORKER_TOKEN $MASTER_IP:2377"

echo "$PREFIX_INFO starting collector binary on master node"
vagrant ssh core-01 -c "sudo systemctl start collector.service"

echo "$PREFIX_INFO starting docker services"
vagrant ssh core-01 -c "docker network create --driver overlay yolo-network"
vagrant ssh core-01 -c "docker service create --name server --endpoint-mode dnsrr --network yolo-network paulboutes/server:1"
vagrant ssh core-01 -c "docker service create --name client --network yolo-network -p 9090:9090 paulboutes/client:1"

vagrant ssh core-01