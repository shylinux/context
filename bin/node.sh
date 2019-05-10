#! /bin/bash

export ctx_box=${ctx_box:="http://localhost:9094"}
export ctx_root="/usr/local/context"
export ctx_home=~/context

ctx_bin="bench" && [ -e bin/bench ] && ctx_bin=bin/bench
export ctx_bin

log() {
    echo -e $*
}

prepare() {
    mkdir -p bin etc usr
    mkdir -p var/run var/log var/tmp
}

main() {
    while true; do
        $ctx_bin "$@" 2>var/log/boot.log && break
        log "restarting..." && sleep 3
    done
}

case $1 in
    create) mkdir -p $2; cd $2 && shift && shift && prepare && main "$@";;
    init) shift; prepare && main "$@";;
    *) mkdir -p var/run var/log && main "$@";;
esac

