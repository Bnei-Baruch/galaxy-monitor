#!/usr/bin/env bash
# run misc/deploy.sh from project root
set -e
set -x

scp galaxy-monitor root@gxydb.kli.one:/opt/galaxy-monitor/galaxy-monitor.new

ssh root@gxydb.kli.one "/bin/cp -f /opt/galaxy-monitor/galaxy-monitor /opt/galaxy-monitor/galaxy-monitor.old"
ssh root@gxydb.kli.one "systemctl stop galaxy-monitor"
ssh root@gxydb.kli.one "mv /opt/galaxy-monitor/galaxy-monitor.new /opt/galaxy-monitor/galaxy-monitor"
ssh root@gxydb.kli.one "systemctl start galaxy-monitor"
