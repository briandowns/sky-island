#!/bin/sh

set -ex

pkg install -y virtualbox-ose-additions
sysrc vboxnet_enable="YES"
sysrc vboxguest_enable="YES"
sysrc vboxservice_enable="YES"