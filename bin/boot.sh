#! /bin/bash

export dev="https://shylinux.com"
export box_root="/usr/local/context"
export box_home="~/context"
bench="bench"

log() {
    echo -e $*
}

prepare() {
    mkdir -p bin etc usr
    mkdir -p var/log var/tmp var/run
}

install() {
    [ -n "$1" ] && dev=$1 && shift
    case `uname -s` in
        "Darwin") GOOS=darwin GOARCH=amd64;;
        *) GOOS=linux GOARCH=386;;
    esac
    case `uname -m` in
        "x86_64") GOARCH=amd64;;
        "armv7l") GOARCH=arm;;
    esac
    log "dev: $dev\nGOOS: $GOOS\nGOARCH: $GOARCH"

    dev=$dev/code/upgrade
    wget -O etc/exit.shy $dev/exit_shy
    wget -O etc/init.shy $dev/init_shy
    wget -O etc/common.shy $dev/common_shy
    wget -O bin/bench "$dev/bench?GOOS=$GOOS&GOARCH=$GOARCH" && chmod u+x bin/bench
    wget -O bin/boot.sh $dev/boot_sh && chmod u+x bin/boot.sh
    wget -O bin/node.sh $dev/node_sh && chmod u+x bin/node.sh
}

state() {
    md=md5sum && [ `uname -s` = "Darwin" ] && md=md5
    for file in bin/node.sh bin/boot.sh bin/bench etc/init.shy etc/common.shy etc/exit.shy; do
        echo `$md $file`
    done
}

action() {
    log "kill" $1 && kill -$1 $pid
}

main() {
    while true; do
        $bench "$@" 2>var/log/boot.log && break
        log "restarting..." && sleep 3
    done
}

dir=$box_root
[ -d "$1" ] && dir=$1 && shift
[ -d "$dir" ] && cd $dir
[ -f bin/bench ] && bench=bin/bench
pid=`cat var/run/bench.pid`
log "dir: $dir\nbench: $bench\npid: $pid"

case $1 in
    install) shift; prepare && install "$@";;
    start|"") shift; prepare && main "$@";;
    state) state;;
    stop) action QUIT;;
    restart) action USR1;;
    upgrade) action USR2;;
esac

