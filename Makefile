
BENCH=src/example/bench.go

install:
	@cp etc/go.snippets ~/.vim/snippets/
	@cp etc/shy.vim ~/.vim/syntax/
	@touch etc/local.shy
	go install $(BENCH)
	@md5 `which bench`
	@date

build:
	go build $(BENCH)
