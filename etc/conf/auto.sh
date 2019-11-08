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
    if [ "$SHELL" = "/bin/zsh" ]; then
        ShyJSON "$@" SHELL "${SHELL}" pwd "${PWD}" sid "${ctx_sid}"|read data
    else
        local data=`ShyJSON "$@" SHELL "${SHELL}" pwd "${PWD}" sid "${ctx_sid}"`
    fi
    curl -s "${ctx_url}" -H "${ctx_head}" -d "${data}"
}
ShyUpload() {
    curl "${ctx_dev}/upload" -F "upload=@$1"
}
ShyDownload() {
    curl "${ctx_dev}/download/$1"
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
    wget -q "${ctx_url}?${data}"
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
        "historys")
            ctx_end=`history|tail -n1|awk '{print $1}'`
            ctx_tail=`expr $ctx_end - $ctx_begin`
            echo $ctx_begin - $ctx_end $ctx_tail
            history|tail -n $ctx_tail |while read line; do
                line=`ShyLine $line`
                ShyPost arg "$line" cmd historys FORMAT "$HISTTIMEFORMAT"
                echo $line
            done
            ctx_begin=$ctx_end
            ;;
        "history") tail -n0 -f $HISTFILE | while true; do read line
            line=`ShyLine $line`
            ShyPost arg "$line" cmd history FORMAT "$HISTTIMEFORMAT"
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
        *)
            ctx_begin=`history|tail -n1|awk '{print $1}'`
            echo "begin: $ctx_begin"
            bind -x '"\C-gl":ShySync input'
        ;;
    esac
}
ShyRecord() {
    script $1
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
    echo "begin: ${ctx_begin}"
}
ShyInit() {
    case "$SHELL" in
        "/bin/zsh");;
        *) PS1="\!-\t[\u@\h]\W\$ ";;
    esac

}

ShyInit && ShyLogin && trap ShyLogout EXIT

