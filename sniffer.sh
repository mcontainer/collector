#!/bin/bash
echo $1
#nsenter --net="$1"/1-rm7o5uxjmo sudo ./../docker-sniffer/dist/sniffer-linux-amd64 -i any -n 1r5454sdd
./../docker-sniffer/dist/sniffer-linux-amd64 -i any -n 1r5454sdd
