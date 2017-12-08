# @debug on
~lex start
source etc/lex.sh
~cli @lex lex

~mdb
open chat chat "chat:chat@/chat" mysql
~chat
query "select * from userinfo"

return
~aaa
login root root
login shy shy


~root cli
remote slaver listen ":9393" tcp
~root aaa
