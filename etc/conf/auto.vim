
let ctx_url = (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9095") . "/code/vim"
let ctx_head = "Content-Type: application/json"
let ctx_sid = ""

fun! ShyPost(arg)
    let a:arg["buf"] = bufname("%")
    let a:arg["pwd"] = getcwd()
    let a:arg["sid"] = g:ctx_sid
    for k in keys(a:arg)
        let a:arg[k] = substitute(a:arg[k], "'", "XXXXXsingleXXXXX", "g")
    endfor

    let data = json_encode(a:arg)
    return system("curl -s '" . g:ctx_url . "' -H '" . g:ctx_head . "' -d '" . data . "' 2>/dev/null")
endfun

fun! ShySync(target)
    if bufname("%") == "ControlP"
        return
    end
    if a:target == "exec"
        call ShyPost({"cmd": "exec", "arg": getcmdline()})
    elseif a:target == "insert"
        call ShyPost({"cmd": "insert", "arg": getreg("."), "row": line("."), "col": col(".")})
    else
        let cmd = {"marks": "marks", "tags": "tags", "fixs": "clist", "bufs": "buffers", "regs": "registers"}
        call ShyPost({"cmd": "sync", "arg": a:target, "sub": execute(cmd[a:target])})
    endif
endfun

fun! Shy(action, target)
    let arg = {"arg": a:target, "cmd": a:action}
    let cmd = ShyPost(arg)
    if cmd != ""
        let arg["res"] = execute(cmd)
        let res = ShyPost(arg)
    endif
endfun

fun! ShyCheck(target)
    if a:target == "exec"
        let cmd = getcmdline()
        if cmd != ""
            call ShySync("exec")
            if getcmdline() == "w"
                call ShySync("tags")
                call ShySync("regs")
                call ShySync("marks")
            endif
        endif
    elseif a:target == "fixs"
        if len(getqflist()) > 1
            copen
            call ShySync("fixs")
		else
            cclose
        end
    end
endfun
fun! ShyLogout()
    call Shy("logout", "")
endfun
fun! ShyLogin()
    let arg = {"cmd": "login", "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER}
    let g:ctx_sid = ShyPost(arg)
endfun

call ShyLogin()
autocmd VimLeave * call ShyLogout()
autocmd InsertLeave * call ShySync("insert")
autocmd CmdlineLeave * call ShyCheck("exec")
autocmd QuickFixCmdPost * call ShyCheck("fixs")

autocmd BufReadPost * call Shy("read", expand("<afile>")) | call ShySync("bufs")
autocmd BufWritePre * call Shy("write", expand("<afile>"))
" autocmd BufUnload * call Shy("close", expand("<afile>")) | call ShySync("bufs")
" autocmd CmdlineLeave * 
call ShySync("tags")


" autocmd CompleteDone * call Shy("sync", "regs")
" autocmd InsertEnter * call Shy("sync", "regs")
" autocmd CmdlineEnter * call Shy("sync", "regs")
" autocmd BufWinEnter * call Shy("enter", expand("<afile>"))
" autocmd WinEnter * call Shy("enter", expand("<afile>"))
" autocmd WinLeave * call Shy("leave", expand("<afile>"))
" autocmd CursorMoved * call Shy("line", getcurpos()[1])
" autocmd InsertCharPre * call Shy("char", v:char)
"
" let g:colorscheme=1
" let g:colorlist = [ "ron", "torte", "darkblue", "peachpuff" ]
" function! ColorNext()
"     if g:colorscheme >= len(g:colorlist)
"         let g:colorscheme = 0
"     endif
"     let g:scheme = g:colorlist[g:colorscheme]
"     exec "colorscheme " . g:scheme
"     let g:colorscheme = g:colorscheme+1
" endfunction
" call ColorNext()
" command! NN call ColorNext()<CR>
" command! RR wa | source ~/.vimrc |e
" command! SS mksession! etc/session.vim

