@debug on
~root aaa
login root root

~root lex
server start
~root cli
@lex lex

# login root 94ca7394d007fa189cc4be0a2625d716 root

# ~cli
# remote slaver listen :9393 tcp

# ~aaa
# login shy shy
# userinfo add context hi hello nice
# userinfo add command hi context
# ~web
# listen
# ~demo
# listen
# ~home spawn test
