
BENCH=src/examples/bench.go

install:
	@go get github.com/go-sql-driver/mysql
	@go get github.com/nsf/termbox-go
	@go get github.com/skip2/go-qrcode
	@go get github.com/gomarkdown/markdown
	go install $(BENCH)
	@date
	bench web.code.counter nmake 1

install_all: install
	touch etc/local.shy
	touch etc/local_exit.shy
	touch etc/init.shy
	touch etc/exit.shy
	touch etc/login.txt
	touch etc/history.txt

build:
	go build $(BENCH)

run:
	etc/run.sh 2>var/error.log

win64:
	GOARCH=amd64 GOOS=windows go build $(BENCH)
	mv bench.exe bench_1.0_win64.exe
win32:
	GOARCH=386 GOOS=windows go build $(BENCH)
	mv bench.exe bench_1.0_win32.exe

linux64:
	GOARCH=amd64 GOOS=linux go build $(BENCH)
	mv bench bench_1.0_linux64
linux32:
	GOARCH=386 GOOS=linux go build $(BENCH)
	mv bench bench_1.0_linux32
linux_arm:
	GOARCH=arm GOOS=linux go build $(BENCH)
	mv bench bench_1.0_linux_arm


DOTS=etc/dotsfile
back_dotsfile:
	cp ~/.zshrc $(DOTS)
	cp ~/.tmux.conf $(DOTS)
	cp ~/context/.git/hooks/post-commit $(DOTS)/git_hooks
	cp ~/.vimrc $(DOTS)
	cp ~/.vim/syntax/shy.vim $(DOTS)

load_dotsfile:\
   	~/.zshrc\
   	~/.tmux.conf\
   	~/context/.git/hooks/post-commit\
   	~/.vimrc\
   	~/.vim/syntax/shy.vim

~/.zshrc: $(DOTS)/.zshrc
	cp $< $@
~/.tmux.conf: $(DOTS)/.tmux.conf
	cp $< $@
~/context/.git/hooks/post-commit: $(DOTS)/git_hooks/post-commit
	cp $< $@
~/.vimrc: $(DOTS)/.vimrc
	cp $< $@
~/.vim/syntax/shy.vim: $(DOTS)/shy.vim
	cp $< $@

