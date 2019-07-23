function! Keys(which, keys)
    for key in a:keys
        if a:which == "Statment"
            let key = "\"\\(^\\|\\t\\|  \\|$(\\)" . key . "\\>\""
        else
            let key = "\"\\<" . key . "\\>\""
        end
        exec  "syn match shy" . a:which. " " . l:key
    endfor
endfunction

highlight shyArgument   ctermfg=cyan
highlight shySubCommand ctermfg=yellow

highlight shyCache      ctermfg=yellow
highlight shyConfig     ctermfg=yellow
highlight shyCommand    ctermfg=green
highlight shyContext    ctermfg=red
highlight shyComment    ctermfg=blue

highlight shyStatment   ctermfg=yellow
highlight shyOperator   ctermfg=yellow
highlight shyVariable   ctermfg=magenta
highlight shyNumber     ctermfg=magenta
highlight shyString     ctermfg=magenta

syn match shyString     "'[^']*'"
syn match shyString	    "\"[^\"]*\""
syn match shyNumber	    "-\=\<\d\+\>#\="
syn match shyNumber	    "false\|true"
syn match shVariable    "\$[_a-zA-Z0-9]\+\>"
syn match shVariable    "@[_a-zA-Z0-9]\+\>"

syn match shyComment    "#.*$"
syn match shyContext    "^\~[a-zA-Z0-9_\.]\+\>"
syn match shyCommand    "\(^\|\t\|  \|$(\)[a-zA-Z0-9_\.]\+\>"

call Keys("Operator", ["new"])
call Keys("Statment", ["config", "cache"])
call Keys("Statment", ["return", "source"])
call Keys("Statment", ["if", "else", "else if", "for", "fun", "end"])
call Keys("Statment", ["let", "var"])
" context nfs
call Keys("SubCommand", ["import", "export", "load", "save"])

" context ctx
call Keys("Argument", ["list", "map"])
" context mdb
call Keys("Argument", ["dbname", "dbhelp"])
" context aaa
call Keys("Argument", ["user", "componet", "command"])
" context web
call Keys("Argument", ["client", "cookie", "header"])

