~cli
	@lex lex
# ~cli
# 	remote slaver listen ":9393" tcp
~aaa
	login root root
# ~ssh
# 	listen :9191
~tcp
	listen :9393
~web
	listen

# ~tcp dial ":9393"
# @debug on
# ~aaa
# 	login shy shy
# ~mdb
# 	open chat:chat@/chat mysql
# ~web listen
# @debug on
# ~nfs
# 	open hi.txt
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


