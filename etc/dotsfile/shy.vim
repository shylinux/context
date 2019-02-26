syntax match   shyComment	"#.*$"
syntax match   shyContext   "^\~[a-zA-Z0-9_\.]\+\>"
syntax match   shyCommand   "\(^\|\t\|  \|$(\)[a-zA-Z0-9_\.]\+\>"
syntax match   shyConfig    "\(^\|\t\|  \|$(\)config\>"
syntax match   shyCache     "\(^\|\t\|  \|$(\)cache\>"

syntax match   shyStmt	"return"

syntax match   shyString	"'[^']*'"
syntax match   shyString	"\"[^\"]*\""
syntax match   shyNumber	"-\=\<\d\+\>#\="
syntax match   shVariable	"\$[_a-zA-Z0-9]\+\>"
syntax match   shVariable	"@[_a-zA-Z0-9]\+\>"

syn match   shySubCommand    "\<\(load\|save\)\>"
" context nfs
syn match   shySubCommand    "\<\(import\|export\)\>"
" context mdb
syn match   shyArgument    "\<\(dbname\|dbhelp\)\>"
" context aaa
syn match   shyArgument    "\<\(componet\|command\)\>"
" context web
syn match   shyArgument    "\<\(client\|cookie\|header\)\>"
syn match   shyOperator    "\<\(new\)\>"

highlight shyComment        ctermfg=blue
highlight shyContext        ctermfg=red
highlight shyCommand        ctermfg=green
highlight shyConfig         ctermfg=yellow
highlight shyCache          ctermfg=yellow

highlight shyStmt           ctermfg=yellow

highlight shyString         ctermfg=magenta
highlight shyNumber         ctermfg=magenta
highlight shyVariable       ctermfg=magenta
highlight shyOperator       ctermfg=yellow
highlight shyArgument       ctermfg=cyan
highlight shySubCommand     ctermfg=yellow


" syn match   shNumber			"-\=\<\d\+\>#\="
"
" syn match   shOperator			"=\|+\|-\|*\|/"
" syn match   shOperator			"<\|<=\|>\|>=\|!=\|=="
" syn match   shOperator			"\\"
"
" " syn keyword shStatement break cd chdir continue eval exec exit kill newgrp pwd read readonly shift trap ulimit umask wait
"
" " syn keyword shStatement if else elif end for
"
" " ctx command
"
" " ctx command
" syn match   shCommand "\(^\|\t\|  \|$(\)command"
" " cli command
" syn match   shStatement "\(^\|\t\|  \|$(\)let"
" syn match   shStatement "\(^\|\t\|  \|$(\)var"
" syn match   shStatement "\(^\|\t\|  \|$(\)return"
" syn match   shStatement "\(^\|\t\|  \|$(\)arguments"
" syn match   shStatement "\(^\|\t\|  \|$(\)source"
" syn match   shCommand "\(^\|\t\|  \|$(\)alias"
"
" " aaa command
" syn match   shCommand "\(^\|\t\|  \|$(\)hash"
" syn match   shCommand "\(^\|\t\|  \|$(\)auth"
" syn match   shCommand "\(^\|\t\|  \|$(\)role"
" syn match   shCommand "\(^\|\t\|  \|$(\)user"
" syn match   shSubCommand "\<\(componet\|command\)\>"
"
" " web command
" syn match   shCommand "\(^\|\t\|  \|$(\)serve"
" syn match   shCommand "\(^\|\t\|  \|$(\)route"
" syn match   shCommand "\(^\|\t\|  \|$(\)client"
" syn match   shCommand "\(^\|\t\|  \|$(\)cookie"
" syn match   shCommand "\(^\|\t\|  \|$(\)template"
"
" " mdb command
" syn match   shCommand "\(^\|\t\|  \|$(\)open"
"
"
