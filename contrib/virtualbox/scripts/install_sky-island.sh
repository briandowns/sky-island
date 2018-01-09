#!/bin/sh

set -ex

GO_VERSION="1.9.2"

echo "Installing required software..."
pkg install -y influxdb telegraf grafana4 git curl

echo "Installing Go..."
curl https://dl.google.com/go/go${GO_VERSION}.freebsd-amd64.tar.gz -o go${GO_VERSION}.freebsd-amd64.tar.gz
tar -C /usr/local -xzf go${GO_VERSION}.freebsd-amd64.tar.gz
echo 'setenv PATH $PATH\:/usr/local/bin\:$HOME/bin\:/usr/local/go/bin\:.'  >> ~/.cshrc
PATH=$PATH:/usr/local/go/bin export PATH

echo "Installing Sky Island..."
mkdir -p go/src/briandowns go/bin go/pkg
cd go/src/briandowns
git clone https://github.com/briandowns/sky-island.git
cd sky-island
go get ./...
make install

echo "Starting supplimentary services..."
/usr/local/etc/rc.d/influxd start
/usr/local/etc/rc.d/telegraf start
/usr/local/etc/rc.d/grafana start

echo "Setting Sky Island configuration..."
echo '{
    "release": "/tmp/11.1-RELEASE",
    "http_port": 3280,
    "admin_api_token": "asdfasdfasdfasdf",
    "admin_token_header": "X-Sky-Island-Token",
    "go_version": "1.9.2",
    "base_sys_pkg_dir": "/tmp/11.1-RELEASE",
    "filesystem": {
        "zfs_dataset": "zroot",
        "compression": false
    },
    "network": {
        "ip4": {
            "interface": "em0",
            "start_addr": "192.168.0.20",
            "mask": "255.255.255.0",
            "range": 220,
            "gateway": "192.168.0.1",
            "dns": [
                "4.2.2.1",
                "4.2.2.2"
            ]
        }
    },
    "jails": {
        "base_jail_dir": "/zroot/jails",
        "cache_default_expiration": "8h",
        "cache_purge_after": "24h",
        "children_max": 0,
        "monitoring_addr": "127.0.0.1",
        "build_timeout": "10s",
        "exec_timeout": "5s"
    }
}
' > /usr/local/etc/sky-island.json
echo "Edit /usr/local/etc/sky-island.json as necessary..."

echo 'sky-island_enable="YES"' >> /etc/rc.conf

echo "Starting Sky Island..."
/usr/local/etc/rc.d/sky-island start

exit 0
