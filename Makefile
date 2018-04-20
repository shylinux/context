
BENCH=src/example/bench.go

install:
	go install $(BENCH)

build:
	go build $(BENCH)
