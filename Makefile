
BENCH=src/example/bench.go

install:
	@go get github.com/go-sql-driver/mysql
	@go get github.com/nsf/termbox-go
	@go get github.com/skip2/go-qrcode
	@go get github.com/gomarkdown/markdown
	@cp etc/shy.vim ~/.vim/syntax/
	@touch etc/local.shy
	@touch etc/local_exit.shy
	@touch etc/init.shy
	@touch etc/exit.shy
	@touch etc/login.txt
	@touch etc/history.txt
	go install $(BENCH)
	# @[ `uname` = "Darwin" ] && md5 `which bench`
	@date

build:
	go build $(BENCH)

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


DOCS=etc/dotsfile
back_dotsfile:
	cp ~/.zshrc $(DOCS)
	cp ~/.tmux.conf $(DOCS)
	cp ~/.vimrc $(DOCS)

load_dotsfile: ~/.zshrc ~/.tmux.conf ~/.vimrc
~/.zshrc: $(DOCS)/.zshrc
	cp $(DOCS)/.zshrc ~/
~/.tmux.conf: $(DOCS)/.tmux.conf
	cp $(DOCS)/.tmux.conf ~/
~/.vimrc: $(DOCS)/.vimrc
	cp $(DOCS)/.vimrc ~/

