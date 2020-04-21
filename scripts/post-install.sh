#!/bin/bash

BIN_DIR=/usr/bin
LOG_DIR=/var/log/gogios
DATABASE_DIR=/var/lib/gogios
SCRIPT_DIR=/usr/lib/gogios/scripts

function install_init() {
    cp -f $SCRIPT_DIR/init.sh /etc/init.d/gogios
    chmod +x /etc/init.d/gogios
}

function install_systemd() {
    cp -f $SCRIPT_DIR/gogios.service $1
    systemctl enable gogios || true
    systemctl daemon-reload || true
}

function install_update_rcd() {
    update-rc.d gogios defaults
}

function install_chkconfig() {
    chkconfig --add gogios
}

# Add defaults file if it doesn't exist
if [[ ! -d /etc/default/gogios ]]; then
    touch /etc/default/gogios
fi

# If the user has no checks yet, give them the example file
if [ ! -f /etc/gogios/checks.json ]; then
    echo "Making default check file"
    cp /etc/gogios/example.json /etc/gogios/checks.json
fi

# Distribution specific
if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /usr/lib/systemd/system/gogios.service
    else
        # SysVinit
        install_init
        # Try update-rc.d then fallback to chkconfig
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    test -d $LOG_DIR || mkdir -p $LOG_DIR
    chown -R -L gogios:gogios $LOG_DIR
    chmod 755 $LOG_DIR
    test -d $DATABASE_DIR || mkdir -p $DATABASE_DIR
    chown -R -L gogios:gogios $DATABASE_DIR
    chmod 755 $DATABASE_DIR

    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /lib/systemd/system/gogios.service
        deb-systemd-invoke restart gogios.service || echo "WARNING: systemd not running."
    else
        # SysVinit
        install_init
        # Try update-rc.d then fallback to chkconfig
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
        invoke-rc.d gogios restart
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ "$NAME" = "Arch Linux" ]]; then
        test -d $LOG_DIR || mkdir -p $LOG_DIR
        chown -R -L gogios:gogios $LOG_DIR
        chmod 755 $LOG_DIR
        test -d $DATABASE_DIR || mkdir -p $DATABASE_DIR
        chown -R -L gogios:gogios $DATABASE_DIR
        chmod 755 $DATABASE_DIR
        install_systemd /usr/lib/systemd/system/gogios.service
    fi
fi
