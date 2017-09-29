#!/usr/bin/env bash
mkdir -p /opt/collector
mkdir -p /opt/processor
mv collector /opt/collector
mv processor /opt/processor

mv collector.service /etc/systemd/system/
mv processor.service /etc/systemd/system/

#systemctl enable processor.service
#systemctl enable collector.service
