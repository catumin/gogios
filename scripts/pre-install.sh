#!/bin/bash

if ! grep "^gogios:" /etc/group &>/dev/null; then
    groupadd -r gogios
fi

if ! id gogios &>/dev/null; then
    useradd -r -M gogios -s /bin/false -d /var/spool/gogios -g gogios
fi
