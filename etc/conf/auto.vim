let ctx_dev = (len($ctx_dev) > 1? $ctx_dev: "http://localhost:9095") . "/code/vim"

fun! ShyPost(arg)
    return system("curl -s '" . g:ctx_dev . "' -H 'Content-Type: application/json' -d '" . json_encode(a:arg) . "' 2>/dev/null")
endfun

fun! Shy(action, target)
    let arg = {"arg": a:target, "cmd": a:action, "pwd": getcwd(), "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER}
    if a:action == "sync"
        let cmd = {"tags": "tags", "bufs": "buffers", "regs": "registers", "marks": "marks"}
        let arg[a:target] = execute(cmd[a:target])
    endif

    let cmd = ShyPost(arg)
    if cmd != ""
        let arg["res"] = execute(cmd)
        let res = ShyPost(arg)
    endif
endfun

autocmd BufReadPost * call Shy("read", expand("<afile>"))
autocmd BufWritePre * call Shy("write", expand("<afile>"))
autocmd BufUnload * call Shy("close", expand("<afile>"))

autocmd BufWritePost * call Shy("sync", "bufs")
autocmd BufWritePost * call Shy("sync", "tags")
autocmd BufWritePost * call Shy("sync", "regs")

" autocmd BufWinEnter * call Shy("enter", expand("<afile>"))
" autocmd WinEnter * call Shy("enter", expand("<afile>"))
" autocmd WinLeave * call Shy("leave", expand("<afile>"))
"
" autocmd InsertEnter * call Shy("line", getcurpos()[1])
" autocmd CursorMoved * call Shy("line", getcurpos()[1])

" autocmd InsertCharPre * call Shy("char", v:char)
