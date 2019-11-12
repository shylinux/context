#!/bin/sh

if [ "${ctx_dev}" = "" ] || [ "${ctx_dev}" = "-" ]; then
    ctx_dev="http://localhost:9095"
fi

ctx_url=$ctx_dev"/code/zsh"
ctx_head=${ctx_head:="Content-Type: application/json"}
ctx_sync=${ctx_sync:=""}
ctx_sid=${ctx_sid:=""}

ctx_welcome=${ctx_welcome:="^_^  Welcome to Context world  ^_^"}
ctx_goodbye=${ctx_goodbye:="^_^  Goodbye to Context world  ^_^"}

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
ShyDownload() {
    curl "${ctx_dev}/download/$1"
}
ShyUpdate() {
    curl "${ctx_dev}/publish/$1" > $1
}
ShyUpload() {
    curl "${ctx_dev}/upload" -F "upload=@$1"
}
ShyFuck() {
    local TEMP=`mktemp /tmp/tmp.XXXXXX` && $@ > $TEMP
    local data=`ShyForm cmd sync arg $1 SHELL "${SHELL}" pwd "${PWD}" sid "${ctx_sid}"`
    curl "${ctx_url}?$data" -F "sub=@$TEMP"
}
ShySync() {
    case "$1" in
        "df")
            ShyFuck df
            ;;
        "ps")
            ShyFuck ps
            ;;
        "env")
            ShyFuck env
            ;;
        "free")
            ShyFuck free
            ;;
        "historys")
            ctx_end=`history|tail -n1|awk '{print $1}'`
            ctx_tail=`expr $ctx_end - $ctx_begin`
            echo 
            echo "upload history $ctx_begin-$ctx_end count $ctx_tail to $ctx_dev"
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

ShySyncs() {
    case "$1" in
        "base")
            ShySync df
            ShySync env
            ShySync free
            ShySync historys
            ;;
        *)
    esac

}
ShyHistory() {
    case "$SHELL" in
        "/bin/zsh")
            ShySync history &>/dev/null &
            ctx_sync=$!
            ;;
        *) ;;
    esac
}
ShyRecord() {
    script $1
}
Shy() {
    local ctx_res=`ShyPost cmd "$1" arg "$2"`
    case "$ctx_res" in
        "PS1");;
        *) [ -n "${ctx_res}" ] && ShyPost cmd "$1" arg "$2" res `sh -c ${ctx_res}`
    esac
}

ShyLogout() {
    echo ${ctx_goodbye}
    sleep 1
    Shy logout
}
ShyLogin() {
    HOST=`hostname` ctx_sid=`ShyPost cmd login pid "$$" pane "${TMUX_PANE}" hostname "${HOST}" username "${USER}"`
    echo ${ctx_welcome}
    echo "url: ${ctx_url}"
    echo "sid: ${ctx_sid:0:6}"
    echo "pid: $$"
    echo "pane: $TMUX_PANE"
    echo "begin: ${ctx_begin}"
}
ShyInit() {
    ctx_begin=`cat $HISTFILE|wc -l`
    case "$SHELL" in
        "/bin/zsh");;
        *)
            ctx_begin=`history|tail -n1|awk '{print $1}'`
            bind -x '"\C-t":ShySyncs base'
            # bind -x '"\C-gl":ShySync input'
            PS1="\!-\t[\u@\h]\W\$ "
            # PS1="\!-\t\$ "
            ;;
    esac
}

ShyInit && ShyLogin && trap ShyLogout EXIT

