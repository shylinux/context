~cli
	@lex lex
	~aaa login root root
	~cli var a = 10
if $username == root
	echo welcome root user: $username
end
~web serve

function nice
	echo who
end

return hello hello

sleep 1
~host1
	remote context mpa register terminal shhylinux term term term 1
	$sessid $result
	remote cache sessid $sessid
	~nfs save usr/sess.txt "terminal: " $sessid
	~nfs genqr usr/sess.png "terminal: " $sessid
return
# ~ssh dial chat.shylinux.com:9090 true
# ~cli
# 	remote slaver listen ":9393" tcp
# ~ssh
# 	listen :9191
~tcp
	listen :9393

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


@debug
~web spawn hi he ./
route template /tpl ./usr/msg.tpl
route script /php ./usr/msg.php
route script /who who
~hi listen ./ ":9494"
master nice
return

login shy shy


