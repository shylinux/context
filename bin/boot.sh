#! /bin/bash

ctx_log=${ctx_log:="var/log"}
ctx_app=${ctx_app:="bench"}
ctx_bin=${ctx_app} && [ -f bin/${ctx_app} ] && ctx_bin=$(pwd)/bin/${ctx_app}
# ctx_box=
ctx_dev=${ctx_dev:="https://shylinux.com"}
# ctx_cas=
ctx_root=${ctx_root:=/usr/local/context}
ctx_home=${ctx_home:=~/context}
# web_port=
# ssh_port=
# HOSTNAME=
# USER=
# PWD=
export ctx_log ctx_app ctx_bin ctx_dev ctx_root ctx_home

log() {
    echo -e $*
}
install() {
    case `uname -s` in
        "Darwin") GOOS=darwin GOARCH=amd64;;
        *) GOOS=linux GOARCH=386;;
    esac
    case `uname -m` in
        "x86_64") GOARCH=amd64;;
        "armv7l") GOARCH=arm;;
    esac
    wget -O ${ctx_app} "$ctx_dev/publish/${ctx_app}?GOOS=$GOOS&GOARCH=$GOARCH" && chmod a+x ${ctx_app} \
        && ./${ctx_app} upgrade system && md5sum ${ctx_app} \
        && mv ${ctx_app} bin/${ctx_app}
}
main() {
    while true; do
        ${ctx_bin} "$@" 2>${ctx_log}/boot.log && break
        log "restarting..." && sleep 3
    done
}
action() {
    pid=$(cat var/run/bench.log)
    log "kill" $1 && kill -$1 ${pid}
}


dir=./ && [ -d "$1" ] && dir=$1 && shift
[ -d "${dir}" ] && cd ${dir}
log "dev:$ctx_dev\ndir: $dir\nbin: $ctx_bin\n"

case $1 in
    install) install "$@";;
    start|"") main "$@";;
    create) mkdir -p $2; cd $2 && shift && shift && main "$@";;
    upgrade) action USR2;;
    restart) action USR1;;
    stop) action QUIT;;
esac

