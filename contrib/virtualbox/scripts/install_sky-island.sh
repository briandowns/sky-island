#!/bin/sh

set -ex

echo "Installing monitoring software"
pkg install -y influxdb telegraf grafana4

echo "Installing Sky Island"
# git clone
# make install
# /usr/local/etc/rc.d/influxd start
# /usr/local/etc/rc.d/telegraf start
# /usr/local/etc/rc.d/grafana start

exit 0
