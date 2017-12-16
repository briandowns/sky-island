#!/usr/local/bin/bash

JAIL=$(which jail)
ZFS=$(which zfs)
JAIL_ROOT="zroot/jails"
FILESYSTEMS=("build" "monitoring" "releases")

# destory_filesystems destroys the created
# filesystems in use with sky-island
function destory_filesystems() {
    for fs in ${FILESYSTEMS[*]}; do 
        ${ZFS} destroy -rf ${JAIL_ROOT}/${fs}
    done
}

# kill_all_jails kills all running jails 
# found on the system
function kill_all_jails() {
    JAIL_IDS=$(jls | grep -v JID | awk '{print $1}')
    if [ ! -z ${JAIL_IDS} ]; then
        for jid in ${JAIL_IDS}; do
            ${JAIL} -r ${jid}
        done
    fi
}

# destroy_snapshot
function destroy_snapshot() {
    ${ZFS} destroy -rf ${JAIL_ROOT}/releases/11.1-RELEASE@p1
}

echo "Removing sky-island filesystems..."
destory_filesystems

echo "Removing sky-island jails..."
kill_all_jails

echo "Removing sky-island snapshot..."
destroy_snapshot

exit 0