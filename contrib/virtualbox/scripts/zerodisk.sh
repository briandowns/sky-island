#!/bin/sh

set -ex

if zfs list -H -o name zroot; then
  zfs create -o compression=off -o sync=standard -o mountpoint=/var/tmp zroot/empty
  trap 'zfs destroy zroot/empty' 0
fi

dd if=/dev/zero of=/var/tmp/EMPTY bs=1M || :

rm /var/tmp/EMPTY
sync