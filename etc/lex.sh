~root lex
server start
train [a-zA-Z][a-zA-Z0-9]* 2 2
train 0x[0-9]+ 3 2
train [0-9]+ 3 2
train "[^"]*" 4 2
train '[^']*' 4 2
train [~!@#$&*:] 4 2

~root cli
@lex lex

