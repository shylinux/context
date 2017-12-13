# ~lex source etc/lex.sh
~cli @lex lex
~root aaa login root root
# ~tcp listen ":9393"
# ~tcp dial ":9393"
~cli remote slaver listen ":9393" tcp
# @debug on

~mdb open chat chat "chat:chat@/chat" mysql
~web listen
return


@debug
~web spawn hi he ./
route template /tpl ./usr/msg.tpl
route script /php ./usr/msg.php
route script /who who
~hi listen ./ ":9494"
master nice
return

login shy shy


