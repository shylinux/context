# @debug on
# ~root aaa
# login root root
#

source etc/lex.sh

~root mdb
open chat chat "chat:chat@/chat" mysql

~root aaa
login root root

# ~root
# $nserver

#
# login root 94ca7394d007fa189cc4be0a2625d716 root

# ~root tcp
# listen ":9393"
# listen ":9394"
~root cli
remote slaver listen ":9393" tcp

# ~aaa
# login shy shy
# userinfo add context hi hello nice
# userinfo add command hi context
# ~web
# listen
# ~demo
# listen
# ~home spawn test
