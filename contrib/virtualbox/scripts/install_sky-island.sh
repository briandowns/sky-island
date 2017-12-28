#!/bin/sh

set -ex

echo "Installing monitoring software"
pkg install -y influxdb telegraf grafana4

echo "Installing Sky Island..."
git clone https://github.com/briandowns/sky-island.git
cd sky-island
make install

echo "Starting supplimentary services..."
/usr/local/etc/rc.d/influxd start
/usr/local/etc/rc.d/telegraf start
/usr/local/etc/rc.d/grafana start

echo "Starting Sky Island..."
/usr/local/etc/rc.d/sky-island start

exit 0
