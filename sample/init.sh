#!/usr/bin/env bash

PREFIX_INFO="[INFO] ::"

function fetchIP() {
    vagrant ssh $1 -c "ip addr show eth1 | grep 'inet ' | sed -e 's/^.*inet //' -e 's/\/.*$//'" 2> /dev/null
}
vagrant up


echo "$PREFIX_INFO Fetching master and worker ip"

MASTER_IP=`fetchIP "core-01" | tr -d '[:space:]'`
WORKER_IP=`fetchIP "core-02" | tr -d '[:space:]'`
GRAPH_IP=`fetchIP "core-03" | tr  -d '[:space:]'`

echo "$PREFIX_INFO master ip -> $MASTER_IP"
echo "$PREFIX_INFO worker ip -> $WORKER_IP"
echo "$PREFIX_INFO graph_ip -> $GRAPH_IP"

echo "$PREFIX_INFO Creating swarm context"

INIT_SWARM="docker swarm init --advertise-addr $MASTER_IP"

echo "$PREFIX_INFO executing -> $INIT_SWARM"
vagrant ssh core-01 -c "$INIT_SWARM" 2> /dev/null

echo "$PREFIX_INFO Fetching swarm join worker token"
WORKER_TOKEN=`vagrant ssh core-01 -c "docker swarm join-token worker -q" | tr -d '[:space:]' 2> /dev/null`

echo "$PREFIX_INFO TOKEN = $WORKER_TOKEN"

echo "$PREFIX_INFO Adding worker node (core-02) to swarm cluster"
vagrant ssh core-02 -c "docker swarm join --token $WORKER_TOKEN $MASTER_IP:2377"

vagrant ssh "core-03" -c "mkdir -p dgraph && cd dgraph && docker run -d -p 8080:8080 -p 9080:9080 -v ~/dgraph:/dgraph --name dgraph dgraph/dgraph dgraph --bindall=true --memory_mb 2048"

#for i in 1 2
#do
#    echo "$PREFIX_INFO starting processor binary on core-0$i"
#    vagrant ssh "core-0$i" -c "sudo systemctl start processor.service"
#
#    echo "$PREFIX_INFO starting collector binary on core-0$i"
#    vagrant ssh "core-0$i" -c "sudo systemctl start collector.service"
#done

echo "$PREFIX_INFO starting docker services"
vagrant ssh core-01 -c "docker network create --driver overlay custom-network"
vagrant ssh core-01 -c "docker service create --name server --endpoint-mode dnsrr --network custom-network paulboutes/server:latest"
vagrant ssh core-01 -c "docker service create --name client --network custom-network -p 9090:9090 paulboutes/client:1"

vagrant ssh core-01
