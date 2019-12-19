syntax match Comment "Name: \"[^\"]*\""
syntax match Comment "Help: \"[^\"]*\""

highlight kitConst    ctermfg=yellow
syntax match kitConst "kit\.[a-zA-Z_.]*"

highlight msgConst    ctermfg=cyan
syntax match msgConst "m\.[a-zA-Z_.]*"
syntax match msgConst "msg\.[a-zA-Z_.]*"
syntax match msgConst "sub\.[a-zA-Z_.]*"

" highlight iceConst    ctermfg=darkgreen
" syntax match iceConst "ice\.[a-zA-Z_.]*"

