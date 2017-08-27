#!/usr/bin/env bash
mkdir -p /opt/collector
mv collector /opt/collector

mv collector.service /etc/systemd/system/

systemctl enable collector.service