syntax match Comment "Name: \"[^\"]*\""
syntax match Comment "Help: \"[^\"]*\""

highlight kitConst    ctermfg=yellow
syntax match kitConst "kit\.[a-z0-9A-Z_.]*"
syntax match kitConst "app\.[a-z0-9A-Z_.]*"


highlight msgConst    ctermfg=cyan
syntax match msgConst "m\.[a-z0-9A-Z_.]*"
syntax match msgConst "msg\.[a-z0-9A-Z_.]*"
syntax match msgConst "sub\.[a-z0-9A-Z_.]*"

