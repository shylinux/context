
BENCH=src/example/bench.go

install:
	go install $(BENCH)
	touch etc/local.shy
	cp etc/shy.vim ~/.vim/syntax/
	cp etc/go.snippets ~/.vim/snippets/

build:
	go build $(BENCH)
