#! /bin/bash -i

# 日志配置
export ctx_err=${ctx_err:="/dev/null"}
export ctx_log=${ctx_log:="var/log"}
export ctx_log_debug=${ctx_log_debug:=false}
export ctx_log_disable=${ctx_log_disable:=false}
export ctx_gdb_enable=${ctx_gdb_enable:=false}

# 目录配置
export ctx_app=${ctx_app:="shy"}
export ctx_bin=${ctx_bin:=$(pwd)/bin/${ctx_app}}; [ -e $ctx_bin ] || export ctx_bin=`which ${ctx_app}`
export ctx_root=${ctx_root:=/usr/local/context}
export ctx_home=${ctx_home:=~/context}

# 网络配置
export ctx_cas=${ctx_cas:=""}
export ctx_ups=${ctx_ups:=""}
export ctx_box=${ctx_box:=""}
export ctx_dev=${ctx_dev:="https://shylinux.com"}

# 服务配置
export ctx_type=${ctx_type:="node"}
export ssh_port=${ssh_port:=":9090"}
export web_port=${web_port:=":9095"}

# 用户配置
export node_cert=${node_cert:=""}
export node_key=${node_key:=""}
export HOSTNAME=${HOSTNAME:=""}
export USER=${USER:=""}
export PWD=${PWD:=""}

install() {
    if [ -n "$1" ]; then mkdir $1; cd $1; shift; fi

    md5=md5sum
    case `uname -s` in
        "Darwin") GOOS=darwin md5=md5;;
        "Linux") GOOS=linux;;
        *) GOOS=windows;;
    esac
    case `uname -m` in
        "x86_64") GOARCH=amd64;;
        "armv7l") GOARCH=arm;;
        *) GOARCH=386;;
    esac

    echo
    echo
    curl -o ${ctx_app} "$ctx_dev/publish/${ctx_app}?GOOS=$GOOS&GOARCH=$GOARCH" && chmod a+x ${ctx_app} || return

    target=install && [ -n "$1" ] && target=$1
    ${md5} ${ctx_app} && ./${ctx_app} upgrade ${target} || return

    mv ${ctx_app} bin/${ctx_app} && bin/boot.sh
}
main() {
    trap HUP hup
    log "\nstarting..."
    while true; do
        date && ${ctx_bin} "$@" 2>${ctx_err} && break
        log "\n\nrestarting..." && sleep 1
    done
}
action() {
    pid=$(cat var/run/bench.pid)
    log "kill" $1 $pid && kill -$1 ${pid}
}
hup() { echo "term hup"; }
log() { echo -e $*; }

log "bin: $ctx_bin"
log "box: $ctx_box"
log "dev: $ctx_dev"
log "ups: $ctx_ups"

case $1 in
    help) echo
        echo " >>>>  welcome context world!  <<<<"
        echo
        echo "more info see https://github.com/shylinux/context"
        echo
        echo "  install [dir [type]] 安装并启动服务"
        echo "  [start] 启动服务"
        echo "  create dir 创建并启动服务"
        echo "  restart 重启服务"
        echo "  upgrade 升级服务"
        echo "  quit 保存并停止服务"
        echo "  term 停止服务"
    ;;
    install) shift && install "$@";;
    start) shift && main "$@";;
    "") main "$@";;
    create) mkdir -p $2; cd $2 && shift && shift && main "$@";;
    restart) action 30;;
    upgrade) action 31;;
    quit) action QUIT;;
    term) action TERM
esac
