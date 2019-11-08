
let ctx_url = (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9095") . "/code/vim"
let ctx_head = "Content-Type: application/json"
let ctx_sid = ""

fun! ShyPost(arg)
    let a:arg["pwd"] = getcwd()
    let a:arg["sid"] = g:ctx_sid
    return system("curl -s '" . g:ctx_url . "' -H '" . g:ctx_head . "' -d '" . json_encode(a:arg) . "' 2>/dev/null")
endfun

fun! Shy(action, target)
    if g:ctx_sid == ""
        call ShyLogin()
    endif
    let arg = {"arg": a:target, "cmd": a:action}
    if a:action == "sync"
        let cmd = {"tags": "tags", "fixs": "clist", "bufs": "buffers", "regs": "registers", "marks": "marks"}
        let arg[a:target] = execute(cmd[a:target])
    endif

    let cmd = ShyPost(arg)
    if cmd != ""
        let arg["res"] = execute(cmd)
        let res = ShyPost(arg)
    endif
endfun

fun! ShyLogout()
    call Shy("logout", "")
endfun
fun! ShyLogin()
    let arg = {"cmd": "login", "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER}
    let g:ctx_sid = ShyPost(arg)
endfun
autocmd VimEnter * call ShyLogin()
autocmd VimLeave * call ShyLogout()

autocmd BufReadPost * call Shy("read", expand("<afile>"))
autocmd BufWritePre * call Shy("write", expand("<afile>"))
autocmd BufUnload * call Shy("close", expand("<afile>"))

autocmd BufWritePost * call Shy("sync", "tags")
autocmd BufWritePost * call Shy("sync", "fixs")
autocmd BufWritePost * call Shy("sync", "bufs")
autocmd BufWritePost * call Shy("sync", "regs")

" autocmd BufWinEnter * call Shy("enter", expand("<afile>"))
" autocmd WinEnter * call Shy("enter", expand("<afile>"))
" autocmd WinLeave * call Shy("leave", expand("<afile>"))
"
" autocmd InsertEnter * call Shy("line", getcurpos()[1])
" autocmd CursorMoved * call Shy("line", getcurpos()[1])

" autocmd InsertCharPre * call Shy("char", v:char)
