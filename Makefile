
BENCH=src/example/bench.go

install:
	@go get github.com/go-sql-driver/mysql
	@go get github.com/nsf/termbox-go
	@go get github.com/skip2/go-qrcode
	@cp etc/go.snippets ~/.vim/snippets/
	@cp etc/shy.vim ~/.vim/syntax/
	@touch etc/local.shy
	go install $(BENCH)
	@[ `uname` = "Darwin" ] && md5 `which bench`
	@date

build:
	go build $(BENCH)
