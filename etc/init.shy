@lex lex
~cli
	@lex lex
	~aaa login root root
	~ssh dial chat.shylinux.com:9090 true
	sleep 1
~host1
	~aaa login root root
	~web serve
	~nfs load usr/sess.txt
	var sessid = $result
	if -n $sessid
		~host1 remote context mpa register $sessid
	else
		~host1 remote context mpa register terminal shhylinux term term term 1
		let sessid $result
	end

	~host1 cache sessid $sessid
	~host1 remote cache sessid $sessid
	~nfs save usr/sess.txt $sessid
	~nfs genqr usr/sess.png "terminal: " $sessid
return

	return
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


