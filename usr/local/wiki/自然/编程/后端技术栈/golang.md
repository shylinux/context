## golang

- 官网: <https://golang.google.cn>
- 下载: <https://golang.google.cn/dl>
- 文档: <https://golang.google.cn/doc/>
- 源码: <https://dl.google.com/go/go1.11.1.src.tar.gz>
- 开源: <https://github.com/golang/go>

bash tmux golang git vim

## 命令

```
env help version
run test install
get list
```

## 文件
```
package
import
const
type
func
var
```

## 语句
```
if else for range break continue
switch case default fallthrough
defer recover panic
goto return
go select
```

## 表达式
```
//
```
```
0 iota true false nil "" '' ``

int float
bool error
rune string
byte uintptr

interface struct
map chan

make len cap
append copy delete close
new complex real imag
```

## 官方包
```
os flag path time
io bufio
fmt
sync

math
bytes
image
unicode
strings
strconv

net
log
```

io fmt log net bufio bytes database
os flag time path errors syscall plugin
runtime context sync expvar testing debug reflect unsafe
math hash crypto sort container index
unicode strings strconv regexp
encoding archive compress
mime text html image

go/ast
go/build
go/constant
go/doc
go/format
go/importer
go/internal/gccgoimporter
go/internal/gcimporter
go/internal/srcimporter
go/parser
go/printer
go/scanner
go/token
go/types
internal/bytealg
internal/cpu
internal/nettrace
internal/poll
internal/race
internal/singleflight
internal/syscall/unix
internal/syscall/windows
internal/syscall/windows/registry
internal/syscall/windows/sysdll
internal/testenv
internal/testlog
internal/trace
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

