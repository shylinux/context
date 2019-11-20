
let ctx_url = (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9095") . "/code/vim"
let ctx_head = "Content-Type: application/json"
if !exists("g:ctx_sid") | let ctx_sid = "" | end

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

fun! ShyLogout()
    if g:ctx_sid != ""
        let g:ctx_sid = ShyPost({"cmd": "logout"})
    endif
endfun
fun! ShyLogin()
    if g:ctx_sid == ""
        let g:ctx_sid = ShyPost({"cmd": "login", "share": $ctx_share, "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER})
    endif
endfun

fun! ShyDream(target)
    call ShyPost({"cmd": "dream", "arg": a:target})
endfun
fun! ShySync(target)
    if bufname("%") == "ControlP" | return | end

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
    if a:target == "cache"
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

fun! ShyFavors()
    let msg = json_decode(ShyPost({"cmd": "favors"}))
    let i = 0
    for i in range(len(msg["tab"]))
        if msg["tab"][i] == ""
            continue
        endif
        if exists(":TabooOpen")
            execute "TabooOpen " . msg["tab"][i]
        else
            tabnew
        endif
        lexpr msg["fix"][i]
        lopen
    endfor
endfun
fun! ShyFavor(note)
    if !exists("g:favor_tab") | let g:favor_tab = "" | endif
    if !exists("g:favor_note") | let g:favor_note = "" | endif
    if a:note != "" 
        let g:favor_tab = input("tab: ", g:favor_tab)
        let g:favor_note = input("note: ", g:favor_note)
    endif
    call ShyPost({"cmd": "favor", "tab": g:favor_tab, "note": g:favor_note, "arg": getline("."), "line": getpos(".")[1], "col": getpos(".")[2]})
endfun
fun! ShyTask()
    call ShyPost({"cmd": "tasklet", "arg": input("target: "), "sub": input("detail: ")})
endfun
fun! ShyGrep(word)
    if !exists("g:grep_dir") | let g:grep_dir = "./" | endif
    let g:grep_dir = input("dir: ", g:grep_dir, "file")
    execute "grep -rn --exclude tags --exclude '\..*' --exclude '*.tags' " . a:word . " " . g:grep_dir
endfun
fun! ShyTag(word)
    execute "tag " . a:word
endfun

fun! ShyHelp()
    echo ShyPost({"cmd": "help"})
endfun

call ShyLogin()
" call ShySync("bufs")
call ShySync("regs")
call ShySync("marks")
call ShySync("tags")
" call ShySync("fixs")

autocmd VimLeave * call ShyLogout()
autocmd BufReadPost * call ShySync("bufs")
autocmd BufReadPost * call ShySync("read")
autocmd BufWritePre * call ShySync("write")
autocmd CmdlineLeave * call ShyCheck("exec")
autocmd QuickFixCmdPost * call ShyCheck("fixs")
autocmd InsertLeave * call ShySync("insert")

command ShyHelp echo ShyPost({"cmd": "help"})

nnoremap <C-g><C-g> :call ShyGrep(expand("<cword>"))<CR>
" nnoremap <C-g><C-t> :call ShyTag(expand("<cword>"))<CR>
nnoremap <C-g><C-t> :call ShyTask()<CR>
nnoremap <C-g><C-r> :call ShyCheck("cache")<CR>
nnoremap <C-g><C-f> :call ShyFavor("note")<CR>
nnoremap <C-g>f :call ShyFavor("")<CR>
nnoremap <C-g>F :call ShyFavors()<CR>

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

