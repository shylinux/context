#!/bin/sh

if [ "${ctx_dev}" = "" ] || [ "${ctx_dev}" = "-" ]; then
    ctx_dev="http://localhost:9095"
fi

ctx_url=$ctx_dev"/code/zsh"
ctx_get=${ctx_get:="wget -q"}
ctx_curl=${ctx_curl:="curl"}
ctx_head=${ctx_head:="Content-Type: application/json"}
ctx_sync=${ctx_sync:=""}
ctx_sid=${ctx_sid:=""}

ctx_silent=${ctx_silent:=""}
ctx_welcome=${ctx_welcome:="^_^  Welcome to Context world  ^_^"}
ctx_goodbye=${ctx_goodbye:="^_^  Goodbye to Context world  ^_^"}
ctx_bind=${ctx_bind:="bind -x"}
ctx_null=${ctx_null:="false"}

ShyRight() {
    [ "$1" = "" ] && return 1
    [ "$1" = "0" ] && return 1
    [ "$1" = "false" ] && return 1
    [ "$1" = "true" ] && return 0
    return 0
}
ShyEcho() {
    ShyRight "$ctx_silent" || echo "$@"
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
    ${ctx_get} "${ctx_url}?${data}"
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
    ${ctx_curl} -s "${ctx_url}" -H "${ctx_head}" -d "${data}"
}
ShyDownload() {
    ${ctx_curl} -s "${ctx_url}" -F "cmd=download" -F "arg=$1" -F "sid=$ctx_sid"
}
ShyUpdate() {
    ${ctx_curl} -s "${ctx_dev}/publish/$1" > $1
}
ShyUpload() {
    ${ctx_curl} -s "${ctx_url}" -F "cmd=upload" -F "sid=$ctx_sid" -F "upload=@$1"
}
ShyBench() {
    ${ctx_curl} -s "${ctx_dev}/publish/boot.sh" | sh -s installs context
}
ShySend() {
    local TEMP=`mktemp /tmp/tmp.XXXXXX` && "$@" > $TEMP
    ShyRight "$ctx_silent" || cat $TEMP
    ${ctx_curl} -s "${ctx_url}" -F "cmd=sync" -F "arg=$1" -F "args=$*" -F "sub=@$TEMP"\
        -F "SHELL=${SHELL}" -F "pwd=${PWD}" -F "sid=${ctx_sid}"
}
ShySends() {
    local cmd=$1 && shift
    local arg=$2 && shift
    local TEMP=`mktemp /tmp/tmp.XXXXXX` && echo "$@" > $TEMP
    ShyRight "$ctx_silent" || cat $TEMP
    ${ctx_curl} -s "${ctx_url}" -F "cmd=$cmd" -F "arg=$arg" -F "sub=@$TEMP" \
        -F "SHELL=${SHELL}" -F "pwd=${PWD}" -F "sid=${ctx_sid}"
}
ShyRun() {
    ctx_silent=false ShySend "$@"
}
Shy() {
    local ctx_res=`ShyPost cmd "$1" arg "$2"`
    case "$ctx_res" in
        "PS1");;
        *) [ -n "${ctx_res}" ] && ShyPost cmd "$1" arg "$2" res `sh -c ${ctx_res}`
    esac
}

ShyLogout() {
    echo ${ctx_goodbye} && [ "$ctx_sid" != "" ] && Shy logout
}
ShyLogin() {
    HOST=`hostname` ctx_sid=`ShyPost cmd login pid "$$" pane "${TMUX_PANE}" hostname "${HOST}" username "${USER}"`
    echo "sid: ${ctx_sid:0:6}"
}
ShyFavor() {
    [ "$READLINE_LINE" != "" ] && set $READLINE_LINE && READLINE_LINE=""
    [ "$1" != "" ] && ctx_tab=$1
    [ "$2" != "" ] && ctx_note=$2
    ShyPost cmd favor arg "`history|tail -n2|head -n1`" tab "${ctx_tab}" note "${ctx_note}"
}
ShyFavors() {
    [ "$READLINE_LINE" != "" ] && set $READLINE_LINE && READLINE_LINE=""
    ShyPost cmd favor tab "$1"
}
ShySync() {
    [ "$ctx_sid" = "" ] && ShyLogin

    case "$1" in
        "historys") tail -n0 -f $HISTFILE | while true; do read line
                line=`ShyLine $line`
                ShyPost arg "$line" cmd history FORMAT "$HISTTIMEFORMAT"
                echo $line
            done
            ;;
        "history")
            ctx_end=`history|tail -n1|awk '{print $1}'`
            ctx_tail=`expr $ctx_end - $ctx_begin`
            ShyEcho "sync $ctx_begin-$ctx_end count $ctx_tail to $ctx_dev"
            history|tail -n $ctx_tail |while read line; do
                ShySends historys sub "$line"
            done
            ctx_begin=$ctx_end
            ;;
        *) ShySend "$@"
    esac
}
ShySyncs() {
    case "$1" in
        "base")
            ShySync df
            ShySync env
            ShySync free
            ShySync history
            ;;
        *)
    esac
}
ShyHistory() {
    case "$SHELL" in
        "/bin/zsh")
            ShySync df
            ShySync env
            ShySync free
            ShySync historys &>/dev/null &
            ctx_sync=$!
            ;;
        *) ;;
    esac
}
ShyRecord() {
    script $1
}
ShyHelp() {
    ShyPost cmd help arg "$@"
}
ShyInit() {
    [ "$ctx_begin" = "" ] && ctx_begin=`history|tail -n1|awk '{print $1}'`

    case "$SHELL" in
        "/bin/zsh")
            ctx_bind=${ctx_null}
            PROMPT='%![%*]%c$ '
            ;;
        *)
            PS1="\!-$$-\t[\u@\h]\W\$ "
            PS1="\e[32m\!\e[0m-$$-\e[31m$SPY_OWNER\e[0m@\e[33m$SPY_ROLE\e[0m[\e[32m\t\e[0m]\W\$ "
            PS1="\!-$$-\t[\u@\h]\W\$ "
            PS1="\!-$$-\u@\h[\t]\W\$ "
            ;;
    esac

    ${ctx_bind} '"\C-g\C-r":ShySyncs base'
    ${ctx_bind} '"\C-g\C-f":ShyFavor'
    ${ctx_bind} '"\C-gf":ShyFavor'
    ${ctx_bind} '"\C-gF":ShyFavors'

    echo ${ctx_welcome}
    echo "url: ${ctx_url}"
    echo "pid: $$"
    echo "pane: $TMUX_PANE"
    echo "begin: ${ctx_begin}"
}

ShyInit && trap ShyLogout EXIT

