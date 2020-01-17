#!/usr/bin/env bash

# chkconfig: 2345 99 01
# description: gogios daemon

### BEGIN INIT INFO
# Provides:             gogios
# Required-Start:       $all
# Required-Stop:        $remote_fs $syslog
# Default-Start:        2 3 4 5
# Default-Stop:         0 1 6
# Short-Description:    Start gogios at boot time
### END INIT INFO

GOGIOS_OPTS=

USER=gogios
GROUP=gogios

if [ -r /lib/lsb/init-functions ]; then
    source /lib/lsb/init-functions
fi

DEFAULT=/etc/default/gogios

if [ -r $DEFAULT ]; then
    set -o allexport
    source $DEFAULT
    set +o allexport
fi

if [ -z "$STDOUT" ]; then
    STDOUT=/dev/null
fi
if [ ! -f "$STDOUT" ]; then
    mkdir -p `dirname $STDOUT`
fi

if [ ! -f "$STDERR" ]; then
    STDERR=/var/log/gogios/gogios.log
fi
if [ ! -f "$STDERR" ]; then
    mkdir -p `dirname $STDERR`
fi

OPEN_FILE_LIMIT=65536

function pidofproc() {
    if [ $# -ne 3 ]; then
        echo "Expected three arguments, e.g. $0 -p pidfile daemon-name"
    fi

    if [ ! -f "$2" ]; then
        return 1
    fi

    local pidfile=`cat $2`

    if [ "x$pidfile" == "x" ]; then
        return 0
    fi

    if ps --pid "$pidfile" | grep -q $(basename $3); then
        return 0
    fi

    return 1
}

function killproc() {
    if [ $# -ne 3 ]; then
        echo "Expected three arguments, e.g. $0 -p pidfile signal"
    fi

    pid=`cat $2`

    kill -s $3 $pid
}

function log_failure_msg() {
    echo "$@" "[ FAILED ]"
}

function log_success_msg() {
    echo "$@" "[ OK ]"
}

# Process name
name=gogios

# Daemon name
daemon=/usr/bin/gogios

# pid file for daemon
pidfile=/var/run/gogios/gogios.pid
piddir=`dirname $pidfile`

if [ ! -d "$piddir" ]; then
    mkdir -p $piddir
    chown $USER:$GROUP $piddir
fi

# Configuration file
config=/etc/gogios/gogios.toml

# If the daemon is not there, exit
[ -x $daemon ] || exit 5

case $1 in
    start)
        if [ -e "$pidfile" ]; then
            if pidofproc -p $pidfile $daemon >/dev/null; then
                log_failure_msg "$name process is running"
            else
                log_failure_msg "$name pidfile has no corresponding process; ensure $name is stopped and remove $pidfile"
            fi
            exit 0
        fi

        ulimit -n $OPEN_FILE_LIMIT
        if [ $? -ne 0 ]; then
            log_failure_msg "set open limit to $OPEN_FILE_LIMIT"
        fi

        log_success_msg "Starting the process" "$name"
        if command -v startproc >/dev/null; then
            startproc -u "$USER" -g "$GROUP" -p "$pidfile" -q -- "$daemon" -config "$config" $GOGIOS_OPTS
        elif which start-stop-daemon >/dev/null 2>&1; then
            start-stop-daemon --chuid $USER:$GROUP --start --quiet --pidfile $pidfile --exec $daemon -- -config $config $GOGIOS_OPTS >>$STDOUT 2>>$STDERR &
        else
            su -s /bin/sh -c "nohup $daemon -config $config $GOGIOS_OPTS >>$STDOUT 2>>$STDERR &" $USER
        fi
        log_success_msg "$name process was started"
        ;;

    stop)
        if [ -e $pidfile ]; then
            if pidofproc -p $pidfile $daemon >/dev/null; then
                while true; do
                    if ! pidofproc -p $pidfile $daemon >/dev/null; then
                        break
                    fi
                    killproc -p $pidfile SIGTERM 2>&1 >/dev/null
                    sleep 2
                done

                log_success_msg "$name process was stopped"
                rm -f $pidfile
            fi
        else
            log_failure_msg "$name process is not running"
        fi
        ;;

    reload)
        if [ -e $pidfile ]; then
            if pidofproc -p $pidfile $daemon >/dev/null; then
                if killproc -p $pidfile SIGHUP; then
                    log_success_msg "$name process was reloaded"
                else
                    log_failure_msg "$name failed to reload service"
                fi
            fi
        else
            log_failure_msg "$name process is not running"
        fi
        ;;

    restart)
        $0 stop && sleep 2 && $0 start
        ;;

    status)
        if [ -e $pidfile ]; then
            if pidofproc -p $pidfile $daemon >/dev/null; then
                log_success_msg "$name Process is running"
                exit 0
            else
                log_failure_msg "$name Process is not running"
                exit 1
            fi
        else
            log_failure_msg "$name Process is not running"
            exit 3
        fi
        ;;

    version)
        $daemon -version
        ;;

    *)
        echo "Usage: $0 {start|stop|restart|status|version}"
        exit 2
        ;;

esac
