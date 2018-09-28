syn match   shComment			"#.*$"
syn match   shString			"'[^']*'"
syn match   shString			"\"[^\"]*\""
syn match   shNumber			"-\=\<\d\+\>#\="

syn match   shOperator			"=\|+\|-\|*\|/"
syn match   shOperator			"<\|<=\|>\|>=\|!=\|=="
syn match   shOperator			"\\"
syn match   shOperator			"\~[-_a-zA-Z0-9]\+\>"
syn match   shShellVariables	"\$[-_a-zA-Z0-9]\+\>"
syn match   shShellVariables	"@[-_a-zA-Z0-9]\+\>"

syn keyword shStatement break cd chdir continue eval exec exit kill newgrp pwd read readonly return shift test trap ulimit umask wait

syn keyword shStatement source return function
syn keyword shStatement if else elif end for
syn keyword shStatement let var

" ctx command
syn match   shStatement "\(^\|\t\|$(\|\ \ \)cache"
syn match   shStatement "\(^\|\t\|$(\|\ \ \)config"

syn match   shStatement "\(^\|\t\|$(\|\ \ \)detail"
syn match   shStatement "\(^\|\t\|$(\|\ \ \)option"
syn match   shStatement "\(^\|\t\|$(\|\ \ \)append"
syn match   shStatement "\(^\|\t\|$(\|\ \ \)result"

" ctx command
syn match   shCommand "\(^\|\t\|$(\|\ \ \)command"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)context"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)message"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)session"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)server"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)right"

" tcp command
syn match   shCommand "\(^\|\t\|$(\|\ \ \)listen"

" web command
syn match   shCommand "\(^\|\t\|$(\|\ \ \)client"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)serve\$"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)route"

syn match   shCommand "\(^\|\t\|$(\|\ \ \)open"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)cookie"
syn match   shCommand "\(^\|\t\|$(\|\ \ \)login"

hi def link shComment			Comment
hi def link shString			String
hi def link shNumber			Number
hi def link shOperator			Operator
hi def link shShellVariables	PreProc
hi def link shStatement			Statement
hi def link shCommand 	 		Identifier

hi def link shArithmetic		Special
hi def link shCharClass  		Identifier
hi def link shSnglCase	 	  	Statement
hi def link shCommandSub		Special
hi def link shConditional		Conditional
hi def link shCtrlSeq			Special
hi def link shExprRegion		Delimiter
hi def link shFunctionKey		Function
hi def link shFunctionName		Function
hi def link shRepeat			Repeat
hi def link shSet				Statement
hi def link shSetList			Identifier
hi def link shSpecial			Special
hi def link shTodo				Todo
hi def link shAlias				Identifier

