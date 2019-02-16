#! /bin/bash

export ctx_dev=${ctx_dev:="https://shylinux.com"}
export ctx_root="/usr/local/context"
export ctx_home=~/context
export ctx_bin="bench"

log() {
    echo -e $*
}

prepare() {
    mkdir -p bin etc usr
    mkdir -p var/log var/tmp var/run
}

install() {
    [ -n "$1" ] && ctx_dev=$1 && shift
    case `uname -s` in
        "Darwin") GOOS=darwin GOARCH=amd64;;
        *) GOOS=linux GOARCH=386;;
    esac
    case `uname -m` in
        "x86_64") GOARCH=amd64;;
        "armv7l") GOARCH=arm;;
    esac
    log "ctx_dev: $ctx_dev\nGOOS: $GOOS\nGOARCH: $GOARCH"

    ctx_dev=$ctx_dev/code/upgrade
    wget -O etc/exit.shy $ctx_dev/exit_shy
    wget -O etc/init.shy $ctx_dev/init_shy
    wget -O etc/common.shy $ctx_dev/common_shy
    wget -O bin/bench.new "$ctx_dev/bench?GOOS=$GOOS&GOARCH=$GOARCH" && chmod u+x bin/bench.new && mv bin/bench.new bin/bench
    wget -O bin/boot.sh $ctx_dev/boot_sh && chmod u+x bin/boot.sh
    wget -O bin/node.sh $ctx_dev/node_sh && chmod u+x bin/node.sh
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
        $ctx_bin "$@" 2>var/log/boot.log && break
        log "restarting..." && sleep 3
    done
}

dir=$ctx_root
[ -d "$1" ] && dir=$1 && shift
[ -d "$dir" ] && cd $dir
[ -f bin/bench ] && ctx_bin=bin/bench
pid=`cat var/run/bench.pid`
log "dir: $dir\nbench: $ctx_bin\npid: $pid"

case $1 in
    install) shift; prepare && install "$@";;
    start|"") shift; prepare && main "$@";;
    state) state;;
    stop) action QUIT;;
    restart) action USR1;;
    upgrade) action USR2;;
esac

