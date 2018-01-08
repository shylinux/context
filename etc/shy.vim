syn keyword shStatement break cd chdir continue eval exec exit kill newgrp pwd read readonly return shift test trap ulimit umask wait

syn keyword shStatement source return function
syn keyword shStatement if else elif end for
syn keyword shStatement let var

syn match   shOperator		"\~[-_a-zA-Z0-9]\+\>"
syn match   shShellVariables	"\$[-_a-zA-Z0-9]\+\>"
syn match   shShellVariables	"@[-_a-zA-Z0-9]\+\>"

hi def link shArithRegion	shShellVariables
hi def link shBeginHere	shRedir
hi def link shCaseBar	shConditional
hi def link shCaseCommandSub	shCommandSub
hi def link shCaseDoubleQuote	shDoubleQuote
hi def link shCaseIn	shConditional
hi def link shQuote	shOperator
hi def link shCaseSingleQuote	shSingleQuote
hi def link shCaseStart	shConditional
hi def link shCmdSubRegion	shShellVariables
hi def link shColon	shComment
hi def link shDerefOp	shOperator
hi def link shDerefPOL	shDerefOp
hi def link shDerefPPS	shDerefOp
hi def link shDeref	shShellVariables
hi def link shDerefDelim	shOperator
hi def link shDerefSimple	shDeref
hi def link shDerefSpecial	shDeref
hi def link shDerefString	shDoubleQuote
hi def link shDerefVar	shDeref
hi def link shDoubleQuote	shString
hi def link shEcho	shString
hi def link shEchoDelim	shOperator
hi def link shEchoQuote	shString
hi def link shEmbeddedEcho	shString
hi def link shEscape	shCommandSub
hi def link shExDoubleQuote	shDoubleQuote
hi def link shExSingleQuote	shSingleQuote
hi def link shFunction	Function
hi def link shHereDoc	shString
hi def link shHerePayload	shHereDoc
hi def link shLoop	shStatement
hi def link shMoreSpecial	shSpecial
hi def link shOption	shCommandSub
hi def link shPattern	shString
hi def link shParen	shArithmetic
hi def link shPosnParm	shShellVariables
hi def link shQuickComment	shComment
hi def link shRange	shOperator
hi def link shRedir	shOperator
hi def link shSetListDelim	shOperator
hi def link shSetOption	shOption
hi def link shSingleQuote	shString
hi def link shSource	shOperator
hi def link shStringSpecial	shSpecial
hi def link shSubShRegion	shOperator
hi def link shTestOpr	shConditional
hi def link shTestPattern	shString
hi def link shTestDoubleQuote	shString
hi def link shTestSingleQuote	shString
hi def link shVariable	shSetList
hi def link shWrapLineOperator	shOperator

hi def link shArithmetic		Special
hi def link shCharClass		Identifier
hi def link shSnglCase		Statement
hi def link shCommandSub		Special
hi def link shComment		Comment
hi def link shConditional		Conditional
hi def link shCtrlSeq		Special
hi def link shExprRegion		Delimiter
hi def link shFunctionKey		Function
hi def link shFunctionName		Function
hi def link shNumber		Number
hi def link shOperator		Operator
hi def link shRepeat		Repeat
hi def link shSet		Statement
hi def link shSetList		Identifier
hi def link shShellVariables		PreProc
hi def link shSpecial		Special
hi def link shStatement		Statement
hi def link shString		String
hi def link shTodo		Todo
hi def link shAlias		Identifier

