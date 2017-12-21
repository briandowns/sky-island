#!/bin/sh

set -ex

pkg autoremove -y && pkg clean -y

rm -rf /var/db/freebsd-update/files
mkdir /var/db/freebsd-update/files

rm -f /var/db/freebsd-update/*-rollback
rm -rf /var/db/freebsd-update/install.*
rm -f /var/db/dhclient.leases.*
rm -rf /boot/kernel.old
rm -f /*.core
rm -rf /tmp/*