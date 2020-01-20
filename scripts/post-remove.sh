#!/bin/bash

function disable_systemd {
    systemctl disable gogios
    rm -f $1
}

function disable_update_rcd {
    update-rc.d -f gogios remove
    rm -f /etc/init.d/gogios
}

function disable_chkconfig {
    chkconfig --del gogios
    rm -f /etc/init.d/gogios
}

if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    if [[ "$1" = "0" ]]; then
        rm -f /etc/default/gogios

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /usr/lib/systemd/system/gogios.service
        else
            disable_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    if [ "$1" == "remove" -o "$1" == "purge" ]; then
        rm -f /etc/default/gogios

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /lib/systemd/system/gogios.service
        else
            if which update-rc.d &>/dev/null; then
                disable_update_rcd
            else
                disable_chkconfig
            fi
        fi
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ "$NAME" = "Arch Linux" ]]; then
        disable_systemd /usr/lib/systemd/system/gogios.service
    fi
fi
