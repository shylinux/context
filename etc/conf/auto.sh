#!/bin/sh

if [ "$ctx_dev" = "" ] || [ "$ctx_dev" = "-" ]; then
    ctx_dev="http://localhost:9095"
fi

ctx_url=$ctx_dev"/code/zsh"
ctx_head=${ctx_head:="Content-Type: application/json"}
ctx_sync=${ctx_sync:=""}
ctx_sid=${ctx_sid:=""}
ctx_welcome=${ctx_welcome:="^_^  Welcome to Context world  ^_^"}
ctx_goodbye=${ctx_goodbye:="^_^  Welcome to Context world  ^_^"}

ShyJSON() {
    echo -n "{"
    [ -n "$1" ] && echo -n \"$1\"\:\"$2\" && shift 2 && while [ -n "$1" ]; do
        echo -n \, && echo -n \"$1\"\:\"$2\" && shift 2
    done
    echo -n "}"
}
ShyPost() {
    ShyJSON "$@" pwd "$(pwd)" sid "${ctx_sid}"| xargs -d'\n' -n1 curl -s "${ctx_url}" -H "${ctx_head}" -d 2>/dev/null
}
ShySync() {
    case "$1" in
        "history") tail -n0 -f $HISTFILE | while true; do read line
            ShyPost arg "$line" cmd history SHELL $SHELL
            echo $line
        done;;
    "input")
        curl -s "${ctx_url}?cmd=input&arg=$READLINE_LINE" &>/dev/null
        ;;
    esac
}
Shy() {
    local ctx_res=`ShyPost cmd "$1" arg "$2"`
    case "$ctx_res" in
        "PS1");;
        *) [ -n "${ctx_res}" ] && ShyPost cmd "$1" arg "$2" res `sh -c ${ctx_res}`
    esac
}


ShyHistory() {
    case "$SHELL" in
        "/bin/zsh")
            ShySync history &>/dev/null &
            ctx_sync=$!
            ;;
        *) bind -x '"\C-gl":ShySync input'
    esac
}
ShyLogout() {
    echo ${ctx_goodbye}
    ShyPost cmd logout
}
ShyLogin() {
    ctx_sid=`ShyPost cmd login pid "$$" pane "${TMUX_PANE}" hostname "$(hostname)" username "${USER}"`
    echo ${ctx_welcome}
    echo "sid: ${ctx_sid}"
    echo "pid: $$"
}

ShyLogin && trap ShyLogout EXIT

