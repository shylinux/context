#! /bin/bash

log() {
    echo $*
}
prepare() {
    log "prepare dir"
    mkdir -p bin etc usr
    mkdir -p var/log var/tmp var/run
}

dir=/usr/local/context
[ -d "$1" ] && dir=$1 && shift
[ -d "$dir" ] && cd $dir

bench=bench
[ -f bin/bench ] && bench=bin/bench

pid=`cat var/run/bench.pid`

case $1 in
    help)
    cat<<END
$0: context boot script
install: install context
restart: restart context
start: start context
stop: stop context
END
    ;;
    install)
        dev=$2
        prepare
        wget -O etc/exit.shy $2/code/upgrade/exit_shy
        wget -O etc/init.shy $2/code/upgrade/init_shy
        wget -O etc/common.shy $2/code/upgrade/common_shy
        wget -O bin/bench $2/code/upgrade/bench && chmod u+x bin/bench
        wget -O bin/boot.sh $2/code/upgrade/boot_sh && chmod u+x bin/boot.sh
    ;;
    upgrade)
        log "kill" usr1
        kill -USR2 $pid
    ;;
    start|"")
        prepare
        while true; do
            pwd
            log $bench
            $bench 2>var/log/error.log && break
            log "restarting..."
            sleep 3
        done
    ;;
    stop)
        log "kill" quit
        kill -QUIT $pid
    ;;
    restart)
        log "kill" usr1
        kill -USR1 $pid
    ;;
esac
