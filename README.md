# Event collector

## Author 

Paul Boutes

## Abstract

Events collectors are agents which are deployed on
host machines. Their primary role are to collect network and
docker data, and forward them to the graph processor, in order to
generate a representation of your system, by building the container network topology.

![architecture](docs/event-collector.png)

### What type of data are collected ?

As mentioned, we collect 2 types of data:
* docker event
* network traffic

Docker events are collected in order to say: "Hey, a new container appears, let's monitor it".
Then, we collect network traffic from in-use containers.

In order to monitor our system, we need to sniff containers network traffic. As 
we want to operate as well in single host as in multi host, we can not listen to 
the `docker0` interface. We have to sniff what we called `overlay network`.
For more information, see this [issue](https://github.com/PaulBoutes/docker-event-collector/issues/6).

### Packet sniffer

Network traffic is sniffed by using [gopacket](https://github.com/google/gopacket). At this time, we are collecting packet from the
network layer (3 in OSI model), for capturing source and destination ip, packet size.

![OSI model](https://upload.wikimedia.org/wikipedia/commons/thumb/8/8d/OSI_Model_v1.svg/langfr-408px-OSI_Model_v1.svg.png?uselang=fr)

Sniffing performance is very important. For our case, better sniffing performance means less overhead on the application server.
We are using `af_packet` which is faster than `libpcap`.
`af_packet` uses memory mapped sniffing (mmap). Both the user space program and the kernel
map the same memory zone where mmap provides a configurable circular buffer. The kernel
writes packets into it, and the user space program reads from it. So as we are using a shared buffer between the kernel and the user space program,
we are minimizing packet copies.