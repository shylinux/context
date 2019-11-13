
let ctx_url = (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9095") . "/code/vim"
let ctx_head = "Content-Type: application/json"
if !exists("g:ctx_sid")
    let ctx_sid = ""
end

fun! ShyPost(arg)
    let a:arg["buf"] = bufname("%")
    let a:arg["buf"] = bufname("%")
    let a:arg["pwd"] = getcwd()
    let a:arg["sid"] = g:ctx_sid
    for k in keys(a:arg)
        let a:arg[k] = substitute(a:arg[k], "'", "XXXXXsingleXXXXX", "g")
    endfor
    return system("curl -s '" . g:ctx_url . "' -H '" . g:ctx_head . "' -d '" .  json_encode(a:arg) . "' 2>/dev/null")
endfun

fun! ShySync(target)
    if bufname("%") == "ControlP"
        return
    end

    if a:target == "read" || a:target == "write"
        call ShyPost({"cmd": a:target, "arg": expand("<afile>")})
    elseif a:target == "exec"
        call ShyPost({"cmd": a:target, "arg": getcmdline()})
    elseif a:target == "insert"
        call ShyPost({"cmd": a:target, "arg": getreg("."), "row": line("."), "col": col(".")})
    else
        let cmd = {"bufs": "buffers", "regs": "registers", "marks": "marks", "tags": "tags", "fixs": "clist"}
        call ShyPost({"cmd": "sync", "arg": a:target, "sub": execute(cmd[a:target])})
    endif
endfun

fun! ShyCheck(target)
    if a:target == "login"
        if g:ctx_sid == ""
            let arg = {"cmd": "login", "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER}
            let g:ctx_sid = ShyPost(arg)
        endif
    elseif a:target == "favor"
        cexpr ShyPost({"cmd": "favor"})
    elseif a:target == "favors"
        let msg = json_decode(ShyPost({"cmd": "favors"}))
        let i = 0
        for i in range(len(msg["tab"]))
            tabnew
            lexpr msg["fix"][i]
            lopen
        endfor
    elseif a:target == "cache"
        call ShySync("bufs")
        call ShySync("regs")
        call ShySync("marks")
        call ShySync("tags")
    elseif a:target == "exec"
        let cmd = getcmdline()
        if cmd != ""
            call ShySync("exec")
            if getcmdline() == "w"
                call ShySync("regs")
                call ShySync("marks")
                call ShySync("tags")
            endif
        endif
    elseif a:target == "fixs"
        let l = len(getqflist())
        if l > 0
            execute "copen " . (l > 10? 10: l + 1)
            call ShySync("fixs")
		else
            cclose
        end
    end
endfun

fun! Shy(action, target)
    let arg = {"arg": a:target, "cmd": a:action}
    let cmd = ShyPost(arg)
    if cmd != ""
        let arg["res"] = execute(cmd)
        let res = ShyPost(arg)
    endif
endfun

let favor_tab = ""
let favor_note = ""
fun! ShyFavor(note)
    if a:note == "" 
        call ShyPost({"cmd": "favor", "arg": getline("."), "line": getpos(".")[1], "col": getpos(".")[2]})
    else
        let g:favor_tab = input("tab: ", g:favor_tab)
        let g:favor_note = input("note: ", g:favor_note)
        call ShyPost({"cmd": "favor", "tab": g:favor_tab, "note": g:favor_note, "arg": getline("."), "line": getpos(".")[1], "col": getpos(".")[2]})
    endif
endfun

fun! ShyLogout()
    call Shy("logout", "")
    let g:ctx_sid = ""
endfun

call ShyCheck("login")
autocmd VimLeave * call ShyLogout()

autocmd InsertLeave * call ShySync("insert")
autocmd CmdlineLeave * call ShyCheck("exec")
autocmd BufReadPost * call Shy("read", expand("<afile>"))
autocmd BufReadPost * call ShySync("bufs")
autocmd BufWritePre * call Shy("write", expand("<afile>"))

autocmd QuickFixCmdPost * call ShyCheck("fixs")
" call ShySync("bufs")
call ShySync("regs")
call ShySync("marks")
call ShySync("tags")
" call ShySync("fixs")
"
nnoremap <C-R><C-R> :call ShyCheck("cache")<CR>
" nnoremap <C-R><C-F> :call ShyCheck("favor")<CR>
nnoremap <C-R><C-F> :call ShyCheck("favors")<CR>
nnoremap <C-R>F :call ShyFavor("note")<CR>
nnoremap <C-R>f :call ShyFavor("")<CR>

" autocmd BufUnload * call Shy("close", expand("<afile>")) | call ShySync("bufs")
" autocmd CmdlineLeave * 
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
" command! SS mksession! etc/session.vim

