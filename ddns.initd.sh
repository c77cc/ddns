#!/bin/bash
#
# Provides:	 ddns_deamon
# Default-Start: 	2 3 4 5
# Default-Stop: 	0 1 6
# Description: 	This file should be used to construct scripts to be placed in /etc/init.d.

PROG="ddns"
PROG_PATH="/usr/local/ddns/bin"
PID_PATH="/var/run/"

start() {
    if [ -e "$PID_PATH/$PROG.pid" ]; then
        echo "Error! $PROG is currently running!" 1>&2
        exit 1
    else
        $PROG_PATH/$PROG 2>&1 >> /var/log/$PROG &
        exit 0
    fi
}

stop() {
    echo "begin stop"
    if [ -e "$PID_PATH/$PROG.pid" ]; then
        kill $(cat "$PID_PATH/$PROG.pid")
        rm "$PID_PATH/$PROG.pid"
        echo "$PROG stopped"
        exit 0
    else
        ## Program is not running, exit with error.
        echo "Error! $PROG not started!" 1>&2
        exit 1
    fi
}

if [ "$(id -u)" != "0" ]; then
    echo "This script must be run as root" 1>&2
    exit 1
fi

case "$1" in
    start)
        start
        exit 0
    ;;
    stop)
        stop
        exit 0
    ;;
    reload|restart|force-reload)
        stop
        start
        exit 0
    ;;
    **)
        echo "Usage: $0 {start|stop|reload}" 1>&2
        exit 1
    ;;
esac
