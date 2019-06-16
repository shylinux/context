#! /bin/bash

ctx_log=${ctx_log:="var/log"}
ctx_app=${ctx_app:="bench"}
ctx_bin=${ctx_app} && [ -f bin/${ctx_app} ] && ctx_bin=$(pwd)/bin/${ctx_app}
# ctx_box=
# ctx_cas=
ctx_dev=${ctx_dev:="https://shylinux.com"}
ctx_root=${ctx_root:=/usr/local/context}
ctx_home=${ctx_home:=~/context}
# web_port=
# ssh_port=
# HOSTNAME=
# USER=
# PWD=
export ctx_log ctx_app ctx_bin ctx_box ctx_cas ctx_dev
export ctx_root ctx_home web_port ssh_port

log() {
    echo -e $*
}
install() {
    if [ -n "$1" ]; then
        mkdir $1; cd $1
    fi

    md5=md5sum
    case `uname -s` in
        "Darwin") GOOS=darwin GOARCH=amd64 md5=md5;;
        *) GOOS=linux GOARCH=386;;
    esac
    case `uname -m` in
        "x86_64") GOARCH=amd64;;
        "armv7l") GOARCH=arm;;
    esac

    wget -O ${ctx_app} "$ctx_dev/publish/${ctx_app}?GOOS=$GOOS&GOARCH=$GOARCH" && chmod a+x ${ctx_app} \
         && ./${ctx_app} upgrade system && ${md5} ${ctx_app} \
         && mv ${ctx_app} bin/${ctx_app}
}
main() {
    while true; do
        ${ctx_bin} "$@" && break
        log "restarting..." && sleep 3
    done
}
action() {
    pid=$(cat var/run/bench.pid)
    log "kill" $1 $pid && kill -$1 ${pid}
}


dir=./ && [ -d "$1" ] && dir=$1 && shift
[ -d "${dir}" ] && cd ${dir}
log "dev:$ctx_dev\ndir: $dir\nbin: $ctx_bin\n"

case $1 in
    install) shift && install "$@";;
    start|"") main "$@";;
    create) mkdir -p $2; cd $2 && shift && shift && main "$@";;
    restart) action 30;;
    upgrade) action 31;;
    quit) action QUIT;;
    term) action TERM;;
esac

