
let ctx_url = (len($ctx_dev) > 1? $ctx_dev: "http://127.0.0.1:9095") . "/code/vim"
if !exists("g:ctx_sid") | let ctx_sid = "" | end

fun! ShySend(arg)
    if has_key(a:arg, "sub") && a:arg["sub"] != ""
        let temp = tempname()
        call writefile(split(a:arg["sub"], "\n"), temp, "b")
        let a:arg["sub"] = "@" . temp
    endif

    let a:arg["buf"] = bufname("%")
    let a:arg["pwd"] = getcwd()
    let a:arg["sid"] = g:ctx_sid
    let args = ""
    for k in keys(a:arg)
        let args = args . " -F '" . k . "=" . a:arg[k] . "' "
    endfor
    return system("curl -s " . g:ctx_url . args . " 2>/dev/null")
endfun

fun! ShyLogout()
    if g:ctx_sid == "" | return | endif
    call ShySend({"cmd": "logout"})
endfun
fun! ShyLogin()
    let g:ctx_sid = ShySend({"cmd": "login", "share": $ctx_share, "pid": getpid(), "pane": $TMUX_PANE, "hostname": hostname(), "username": $USER})
endfun
fun! ShyFavor()
    if !exists("g:favor_tab") | let g:favor_tab = "" | endif
    if !exists("g:favor_note") | let g:favor_note = "" | endif
    let g:favor_tab = input("tab: ", g:favor_tab)
    let g:favor_note = input("note: ", g:favor_note)
    call ShySend({"cmd": "favor", "tab": g:favor_tab, "note": g:favor_note, "arg": getline("."), "line": getpos(".")[1], "col": getpos(".")[2]})
endfun
fun! ShyFavors()
    let res = split(ShySend({"cmd": "favor", "tab": input("tab: ")}), "\n")
    let page = "" | let note = ""
    for i in range(0, len(res)-1, 2)
        if res[i] != page
            if note != "" | lexpr note | lopen | let note = "" | endif
            execute exists(":TabooOpen")? "TabooOpen " . res[i]: "tabnew"
        endif
        let page = res[i] | let note .= res[i+1] . "\n"
    endfor
    if note != "" | lexpr note | lopen | let note = "" | endif
endfun

fun! ShySync(target)
    if bufname("%") == "ControlP" | return | end

    if a:target == "read" || a:target == "write"
        call ShySend({"cmd": a:target, "arg": expand("<afile>")})
    elseif a:target == "exec"
        call ShySend({"cmd": a:target, "sub": getcmdline()})
    elseif a:target == "insert"
        call ShySend({"cmd": a:target, "sub": getreg("."), "row": line("."), "col": col(".")})
    else
        let cmd = {"bufs": "buffers", "regs": "registers", "marks": "marks", "tags": "tags", "fixs": "clist"}
        call ShySend({"cmd": "sync", "arg": a:target, "sub": execute(cmd[a:target])})
    endif
endfun
fun! ShyCheck(target)
    if a:target == "cache"
        call ShySync("bufs")
        call ShySync("regs")
        call ShySync("marks")
        call ShySync("tags")
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

fun! ShyTask()
    call ShySend({"cmd": "tasklet", "arg": input("target: "), "sub": input("detail: ")})
endfun
fun! ShyGrep(word)
    if !exists("g:grep_dir") | let g:grep_dir = "./" | endif
    let g:grep_dir = input("dir: ", g:grep_dir, "file")
    execute "grep -rn --exclude tags --exclude '*.tags' " . a:word . " " . g:grep_dir
endfun
fun! ShyTag(word)
    execute "tag " . a:word
endfun
fun! ShyHelp()
    echo ShySend({"cmd": "help"})
endfun

call ShyLogin()
autocmd VimLeave * call ShyLogout()
autocmd BufReadPost * call ShySync("bufs")
autocmd BufReadPost * call ShySync("read")
autocmd BufWritePre * call ShySync("write")
autocmd CmdlineLeave * call ShySync("exec")
autocmd QuickFixCmdPost * call ShyCheck("fixs")
autocmd InsertLeave * call ShySync("insert")
nnoremap <C-G><C-G> :call ShyGrep(expand("<cword>"))<CR>
nnoremap <C-G><C-R> :call ShyCheck("cache")<CR>
nnoremap <C-G><C-F> :call ShyFavor()<CR>
nnoremap <C-G>f :call ShyFavors()<CR>
nnoremap <C-G><C-T> :call ShyTask()<CR>

