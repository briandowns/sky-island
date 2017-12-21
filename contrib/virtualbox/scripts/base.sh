#!/bin/sh

set -ex

freebsd-update --not-running-from-cron fetch install || :

echo WITH_PKGNG=yes >> /etc/make.conf

env ASSUME_ALWAYS_YES=YES pkg bootstrap

pkg update && pkg upgrade -y

echo 'autoboot_delay="0"' >> /boot/loader.conf