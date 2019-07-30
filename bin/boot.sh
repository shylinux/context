#! /bin/bash

ctx_log=${ctx_log:="var/log"}
ctx_app=${ctx_app:="bench"}
ctx_bin=${ctx_app} && [ -f bin/${ctx_app} ] && ctx_bin=$(pwd)/bin/${ctx_app}
# ctx_cas=
# ctx_ups=
# ctx_box=
ctx_dev=${ctx_dev:="https://shylinux.com"}
ctx_root=${ctx_root:=/usr/local/context}
ctx_home=${ctx_home:=~/context}
# ctx_type=
# node_cert=
# node_key=
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

    target=system && [ -n "$2" ] && target=$2

    wget -O ${ctx_app} "$ctx_dev/publish/${ctx_app}?GOOS=$GOOS&GOARCH=$GOARCH" && chmod a+x ${ctx_app} \
        && ${md5} ${ctx_app} && ./${ctx_app} upgrade ${target} && ./${ctx_app} upgrade portal \
        && mv ${ctx_app} bin/${ctx_app}

    mkdir -p usr/script && touch usr/script/local.shy && cd etc && ln -s ../usr/script/local.shy .
}
hup() {
    echo "term hup"
}
main() {
    trap HUP hup
    log "\nstarting..."
    while true; do
        date && ${ctx_bin} "$@" && break
        log "\n\nrestarting..." && sleep 1
    done
}
action() {
    pid=$(cat var/run/bench.pid)
    log "kill" $1 $pid && kill -$1 ${pid}
}


dir=./ && [ -d "$1" ] && dir=$1 && shift
[ -d "${dir}" ] && cd ${dir}
log "ups:$ctx_ups"
log "box:$ctx_box"
log "dev:$ctx_dev"
log "bin:$ctx_bin"
log "dir:$dir"

case $1 in
    install) shift && install "$@";;
    start) shift && main "$@";;
    "") main "$@";;
    create) mkdir -p $2; cd $2 && shift && shift && main "$@";;
    restart) action 30;;
    upgrade) action 31;;
    quit) action QUIT;;
    term) action TERM
esac

