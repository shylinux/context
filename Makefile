
BENCH=src/extend/bench.go
upgrade=usr/upgrade/

install:
	@go get github.com/nsf/termbox-go
	@go get github.com/skip2/go-qrcode
	@go get github.com/go-sql-driver/mysql
	@go get github.com/gomarkdown/markdown
	@go get github.com/PuerkitoBio/goquery
	@go get github.com/go-cas/cas
	GOPATH=$(PWD):$(GOPATH) go install $(BENCH)
	@date
	# bench web.code.counter nmake 1

install_all: install
	touch etc/init.shy
	touch etc/exit.shy
	touch etc/local.shy
	touch etc/local_exit.shy

run:
	etc/bootstrap.sh
shy:
	cp -r src/toolkit ~/context/src/
	cp -r src/contexts ~/context/src/
	cp -r src/examples ~/context/src/
	cp -r usr/template ~/context/usr/
	cp -r usr/librarys/ ~/context/usr/
	cp -r bin ~/context

tar:
	[ -e tar ] || mkdir tar
	[ -e tar/bin ] || mkdir tar/bin
	[ -e tar/etc ] || mkdir tar/etc
	cp etc/bootstrap.sh tar/
	cp etc/init.shy tar/etc/
	cp etc/exit.shy tar/etc/
	touch tar/etc/local.shy
	touch tar/etc/exit_local.shy
	[ -e tar/usr ] || mkdir tar/usr
	cp -r usr/template tar/usr
	cp -r usr/librarys tar/usr
	[ -e tar/var ] || mkdir tar/var

tar_all: tar linux64 darwin win64
	cp etc/local.shy tar/etc/
	cp etc/exit_local.shy tar/etc/
	mv bench.darwin tar/bin/
	mv bench.linux64 tar/bin/
	mv bench.win64.exe tar/bin/
	tar zcvf tar.tgz tar

linux_arm:
	GOARCH=arm GOOS=linux go build -o $(upgrade)bench.linux.arm $(BENCH)
linux32:
	GOARCH=386 GOOS=linux go build -o $(upgrade)bench.linux.386 $(BENCH)
linux64:
	GOARCH=amd64 GOOS=linux go build -o $(upgrade)bench.linux.amd64 $(BENCH)
darwin:
	GOARCH=amd64 GOOS=darwin go build -o $(upgrade)bench.darwin.amd64 $(BENCH)
win32:
	GOARCH=386 GOOS=windows go build -o $(upgrade)bench.win32.exe $(BENCH)
win64:
	GOARCH=amd64 GOOS=windows go build -o $(upgrade)bench.win64.exe $(BENCH)


DOTS=etc/dotsfile
back_dotsfile:
	cp ~/.zshrc $(DOTS)
	cp ~/.tmux.conf $(DOTS)
	cp ~/.vimrc $(DOTS)
	cp ~/.vim/syntax/go.vim $(DOTS)
	cp ~/.vim/syntax/shy.vim $(DOTS)

load_dotsfile:\
   	~/.zshrc\
   	~/.tmux.conf\
   	~/.vimrc\
   	~/.vim/syntax/go.vim\
   	~/.vim/syntax/shy.vim

~/.zshrc: $(DOTS)/.zshrc
	cp $< $@
~/.tmux.conf: $(DOTS)/.tmux.conf
	cp $< $@
~/context/.git/hooks/post-commit: $(DOTS)/git_hooks/post-commit
	cp $< $@
~/.vimrc: $(DOTS)/.vimrc
	cp $< $@
~/.vim/syntax/go.vim: $(DOTS)/go.vim
	cp $< $@
~/.vim/syntax/shy.vim: $(DOTS)/shy.vim
	cp $< $@

.PHONY: tar run

