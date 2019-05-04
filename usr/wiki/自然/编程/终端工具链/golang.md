## golang

- 官网: <https://golang.org/>
- 文档: <https://golang.org/doc/>
- 源码: <https://dl.google.com/go/go1.11.1.src.tar.gz>
- 开源: <https://github.com/golang>

## 基本命令
```
run clean build install
fmt fix vet bug
mod get doc list
env help version
test tool generate
```
## 编译过程
```
main() // cmd/compile/main.go:40
    gc.Main() // cmd/compile/internal/gc/main.go:130
        parseFiles() // cmd/compile/internal/gc/noder.go:26
            syntax.Parse() // cmd/compile/internal/syntax/syntax.go:58
                p.fileOrNil() // cmd/compile/internal/syntax/parser.go:58
                    p.funcDeclOrNil()
                        p.funcBody()
                            p.blockStmt()
                                p.stmtList()
                                    p.stmtOrNil()
                                        p.simpleStmt()
                                            p.exprList()
                                                p.expr()
                                                    p.binaryExpr()
                                                        p.unaryExpr()

```

fmt
mime
text
html
image
unicode
strings
strconv
encoding

hash
math
sort
index
container
compress
archive
crypto
regexp

os
flag
path
time
errors
syscall

io
log
net
bytes
bufio
database

go
cmd
sync
debug
plugin
vendor
unsafe
expvar
runtime
context
testing
reflect
builtin
internal

