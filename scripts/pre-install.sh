#!/bin/bash

if ! grep "^gogios:" /etc/group &>/dev/null; then
    groupadd -r gogios
fi

if ! id gogios &>/dev/null; then
    useradd -r -M gogios -s /bin/false -d /var/spool/gogios -g gogios
fi

if [ ! -f /etc/gogios/checks.json ]; then
    echo "Making default check file"
    cp /etc/gogios/example.json /etc/gogios/checks.json
fi

if [ ! -f /etc/gogios/gogios.toml ]; then
    echo "Making default config"
    cp /etc/gogios/gogios.sample.toml /etc/gogios/gogios.toml
fi
