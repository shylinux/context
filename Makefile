
PUBLISH=usr/publish
BENCH=src/extend/shy.go
BUILD=go build -o $(PUBLISH)/
TARGET=shy

install:prepare
	GOPATH=$(PWD):$(GOPATH) go install $(BENCH) && date && echo
	@bin/boot.sh restart && date

prepare:
	@go get github.com/shylinux/icebergs
	@go get github.com/shylinux/toolkits
	@go get github.com/nsf/termbox-go
	@go get github.com/gorilla/websocket
	@go get github.com/go-sql-driver/mysql
	@go get github.com/gomodule/redigo/redis
	@go get github.com/gomarkdown/markdown
	@go get github.com/skip2/go-qrcode
	@go get gopkg.in/gomail.v2

gotags:
	gotags -f golang.tags -R $(GOROOT)/src
tags:
	gotags -f ctx.tags -R src
tool:
	go get github.com/nsf/gocode
	go get github.com/jstemmer/gotags
	# go get github.com/bradfitz/goimports
	go get github.com/Go-zh/tools/cmd/gopls

linux:
	GOPATH=$(PWD):$(GOPATH) GOOS=linux $(BUILD)$(TARGET).linux.$(shell go env GOARCH) $(BENCH)
linux_arm:
	GOARCH=arm GOOS=linux $(BUILD)$(TARGET).linux.arm $(BENCH)
linux32:
	GOARCH=386 GOOS=linux $(BUILD)$(TARGET).linux.386 $(BENCH)
linux64:
	GOARCH=amd64 GOOS=linux $(BUILD)$(TARGET).linux.amd64 $(BENCH)
darwin:
	GOARCH=amd64 GOOS=darwin $(BUILD)$(TARGET).darwin.amd64 $(BENCH)
win64:
	GOARCH=amd64 GOOS=windows $(BUILD)$(TARGET).win64.exe $(BENCH)
win32:
	GOARCH=386 GOOS=windows $(BUILD)$(TARGET).win32.exe $(BENCH)

