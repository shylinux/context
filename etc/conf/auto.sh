#!/bin/sh

if [ "${ctx_dev}" = "" ] || [ "${ctx_dev}" = "-" ]; then
    ctx_dev="http://localhost:9095"
fi

ctx_url=$ctx_dev"/code/zsh"
ctx_head=${ctx_head:="Content-Type: application/json"}
ctx_sync=${ctx_sync:=""}
ctx_sid=${ctx_sid:=""}
ctx_welcome=${ctx_welcome:="^_^  Welcome to Context world  ^_^"}
ctx_goodbye=${ctx_goodbye:="^_^  Welcome to Context world  ^_^"}

ShyLine() {
    echo "$*"|sed -e 's/\"/\\\"/g' -e 's/\n/\\n/g'
}
ShyJSON() {
    [ $# -eq 1 ] && echo \"`ShyLine "$1"`\" && return
    echo -n "{"
    while [ $# -gt 1 ]; do
        echo -n \"`ShyLine "$1"`\"\:\"`ShyLine "$2"`\"
        shift 2 && [ $# -gt 1 ] && echo -n ","
    done
    echo -n "}"
}
ShyPost() {
    local data=`ShyJSON "$@" SHELL "${SHELL}" pwd "${PWD}" sid "${ctx_sid}"`
    curl -s "${ctx_url}" -H "${ctx_head}" -d "${data}"
}
ShyWord() {
    echo "$*"|sed -e 's/\ /%20/g' -e 's/\n/\\n/g'
}
ShyForm() {
    while [ $# -gt 1 ]; do
        echo -n "`ShyWord "$1"`=`ShyWord "$2"`"
        shift 2 && [ $# -gt 1 ] && echo -n "&"
    done
}
ShyGet() {
    local data=`ShyForm "$@" SHELL "${SHELL}" pwd "${PWD}" sid "${ctx_sid}"`
    curl -s "${ctx_url}?${data}"
}
Shy() {
    local ctx_res=`ShyPost cmd "$1" arg "$2"`
    case "$ctx_res" in
        "PS1");;
        *) [ -n "${ctx_res}" ] && ShyPost cmd "$1" arg "$2" res `sh -c ${ctx_res}`
    esac
}

ShySync() {
    case "$1" in
        "history") tail -n0 -f $HISTFILE | while true; do read line
            line=`ShyLine $line`
            Shy history "$line"
            echo $line
        done;;
        "input")
            ShyGet arg "$READLINE_LINE" cmd "input" SHELL "$SHELL"
        ;;
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
    Shy logout
}
ShyLogin() {
    HOST=`hostname` ctx_sid=`ShyPost cmd login pid "$$" pane "${TMUX_PANE}" hostname "${HOST}" username "${USER}"`
    echo ${ctx_welcome}
    echo "url: ${ctx_url}"
    echo "sid: ${ctx_sid:0:6}"
    echo "pid: $$"
}

ShyLogin && trap ShyLogout EXIT

