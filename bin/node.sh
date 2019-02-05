#! /bin/bash

export ctx_box=${ctx_box:="http://localhost:9094"}
export ctx_root="/usr/local/context"
export ctx_home=~/context
export ctx_bin="bench"

export user_cert=etc/user/cert.pem
export user_key=etc/user/key.pem

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
    create) mkdir $2; cd $2 && shift && shift && prepare && main "$@";;
    init) shift; prepare && main "$@";;
    *) mkdir -p var/run var/log && main "$@";;
esac

