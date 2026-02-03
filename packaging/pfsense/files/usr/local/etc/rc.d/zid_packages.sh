#!/bin/sh

# PROVIDE: zid_packages
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name="zid_packages"
rcvar="zid_packages_enable"

load_rc_config $name

: ${zid_packages_enable:="NO"}
: ${zid_packages_pid:="/var/run/zid-packages.pid"}

pidfile="${zid_packages_pid}"
procname="/usr/local/sbin/zid-packages"
command="/usr/sbin/daemon"
command_args="-f -p ${pidfile} ${procname} daemon"

zid_packages_stop() {
	if [ -f "${pidfile}" ]; then
		PID=$(cat "${pidfile}" 2>/dev/null || true)
		if [ -n "${PID}" ]; then
			kill "${PID}" 2>/dev/null || true
		fi
		rm -f "${pidfile}" 2>/dev/null || true
	fi
	pkill -f "${command} ${command_args}" 2>/dev/null || true
}

stop_postcmd="zid_packages_stop"

run_rc_command "$1"
