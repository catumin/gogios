#!/bin/bash

BIN_DIR=/usr/bin

if [[ -f /etc/debian_verison ]]; then
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        deb-systemd-invoke stop gogios.service
    else
        # SysVinit
        invoke-rc.d gogios stop
    fi
fi
