#! /bin/bash

export box="http://localhost:9094"
bench="bench"

log() {
    echo -e $*
}

prepare() {
    mkdir -p bin etc usr
    mkdir -p var/run var/log var/tmp
}

main() {
    while true; do
        $bench "$@" 2>var/log/boot.log && break
        log "restarting..." && sleep 3
    done
}

case $1 in
    create) mkdir $2 && cd $2 && shift && shift && prepare && main "$@";;
    init) shift; prepare && main "$@";;
    *) mkdir -p var/run var/log && main "$@";;
esac

