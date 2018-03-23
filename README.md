## 0 context简介
context: 终端工具链，各种实用的功能模块，通过简洁的接口，自由的组合在一起。

作为一个工具箱，内置各种实用工具，通过灵活的配置，打造个性化的工具链。

作为一个框架，通过模块式开发，可以快速开发各种应用软件。

## 1 context安装
### 1.0 context下载
选择自己操作系统与处理器的类型对应的版本下载，直接运行即可。

https://github.com/shylinux/context-bin/raw/master/bench-linux-arm

https://github.com/shylinux/context-bin/raw/master/bench-linux-386

https://github.com/shylinux/context-bin/raw/master/bench-linux-amd64

https://github.com/shylinux/context-bin/raw/master/bench-windows-386.exe

https://github.com/shylinux/context-bin/raw/master/bench-windows-amd64.exe

https://github.com/shylinux/context-bin/raw/master/bench-darwin-amd64

### 1.1 context源码安装
#### 1.1.0 golang开发环境安装
* 下载：git clone https://github.com/shylinux/context-dev
* 安装：cd context-dev && ./install.sh

#### 1.1.1 context源码安装
* 下载：git clone https://github.com/shylinux/context
* 编译：cd context && go install src/example/bench.go

## 2 context使用
### 2.0 应用示例--启动WEB服务器
```sh
  $ bench
  > ~web
  > serve ./ ':9090'
```
在shell中，运行命令bench，启动应用，进入到一个类似于shell的环境中。

执行"~web"，切换到web模块，执行"serve ./ ':9090'"，在当前目录启动一个WEB服务器，监听地址为"0.0.0.0:9090"。

打开浏览器输入"http://localhost:9090" ，即可看一个静态WEB服务器已经启动。

### 2.1 常用命令
#### 2.1.1 cache: 缓存管理
```sh
web> cache
address(:9090): 服务地址
directory(./): 服务目录
protocol(http): 服务协议
```
输入"cache"命令，即可查看当前模块的缓存数据，通过这些缓存数据，就可以了解模块的当前各种运行状态。
如"address"，代表web模块监听的网络地址为"0.0.0.0:9090"。

#### 2.1.2 config: 配置管理
```sh
web> config
logheaders(yes): 日志输出报文头(yes/no)
```
输入"config"命令，即可查看当前模块的配置信息。通过修改这些配置信息，就可以控制模块的运行的环境，改变模块运行的状态。
如"logheaders"，代表web模块处理网络请求时，是否在日志中输出报文头，此时值为yes即打印报文头。
```sh
web> config logheaders no
web> config
logheaders(no): 日志输出报文头(yes/no)
```
输入"config logheaders no"命令，修改logheaders的值为no，即不在日志中输出报文头。

#### 2.1.3 command: 命令管理
```sh
web> command
serve: serve [directory [address [protocol]]]
route: route directory|template|script route content
/demo: /demo
```
输入"command"命令，即可查看当前模块的命令集合。通过调用这些命令使用模块提供的各种功能。
如"serve"，启动一个web服务器，参数都是可选的。
参数"directory"代表web服务器的响应路径，存放网页的各种文件。
参数"address"，代表服务器监听的网络地址。
参数"protocol"，代表服务器使用的网络协议。

#### 2.1.4 context: 模块管理
```sh
web> context
web(ctx:cli:aaa::): start(:9090) 应用中心
```
输入"context"命令，即可查看当前模块及子模块的基本信息。web模块没有子模块，所以这里只显示了web模块的基本信息。
```sh
web> context ctx
ctx> context
ctx(:cli:aaa:root:root): begin() 模块中心
lex(ctx::aaa:root:root): start(52,19,0) 词法中心
yac(ctx::aaa:root:root): start(26,13,21) 语法中心
cli(ctx:cli:aaa:root:root): start(stdio) 管理中心
ssh(ctx:cli:aaa:root:root): begin() 集群中心
mdb(ctx:cli:aaa:root:root): begin() 数据中心
tcp(ctx::aaa:root:root): begin() 网络中心
web(ctx:cli:aaa:root:root): begin() 应用中心
aaa(ctx::aaa:root:root): start(root) 认证中心
nfs(ctx::aaa:root:root): begin() 存储中心
log(ctx::aaa:root:root): begin() 日志中心
file1(nfs::aaa:root:root): start(var/bench.log) 打开文件
stdio(nfs::aaa:root:root): start(stdio) 扫描文件
file2(nfs::aaa:root:root): start(etc/init.shy) 扫描文件
```
输入"context ctx"命令，切换ctx模块为当前模块。输入"context"命令，即可查看当前模块及子模块的基本信息。
ctx为根模块，所以可以查看到所有模块的基本信息。

#### 2.1.5 message: 消息管理
```sh
tcp> message
requests:
0 9(ctx->tcp): 23:30:19 []
sessions:
historys:
0 9(ctx->tcp): 23:30:19 []
1 4358(cli->tcp): 23:30:22 [context tcp]
  0 4359(cli->log): 23:30:22 [log cmd 2 context [tcp] []]
  1 4361(tcp->log): 23:30:22 [log search 1 match [tcp]]
  2 4363(cli->tcp): 23:30:22 []
```
输入"message"命令，查看当前模块收发的所有消息。
其中"requests"是收到的长消息，如有一条ctx模块发送给tcp模块的编号为9消息。
其中"sessions"，是发送出的长消息，这里为空，所以tcp模块没有发送长消息。
其中"historys"是模块收到的所有消息，包括长消息和短消息。显示中除了显示当前消息，还会显示消息的子消息。
```sh
tcp> message 9
message: 0
9(ctx->tcp): 23:36:48 []
```
输入"message 9"命令，查看编号为9的消息的具体信息。

### 2.2 web模块的命令
web模块，提供web服务。目前有两条命令serve主机管理，route路由管理。
```sh
web> command
serve: serve [directory [address [protocol]]]
route: route directory|template|script route content
/demo: /demo
```

#### 2.2.1 serve主机管理
```sh
web> command serve
serve [directory [address [protocol]]]
    开启应用服务
```
* directory服务目录
* address服务地址(ip:port)
* protocol服务协议(http/https)

#### 2.2.2 route路由管理
```sh
web> command route
route directory|template|script route content
    添加应用内容
```
参数route代表http请求的uri地址，参数content代表响应回复的内容，不同类型的服务有不同的意义。
* directory静态服务
```sh
web> route directory /p pkg
```
命令"route diretory /p pkg"，当web模块接收到请求uri为"/p/"时把目录"pkg"中的内容作为响应回复。

content代表路径，即web服务请求此route路径时，回复的内容为content指定的目录或文件。
* template模板服务
```sh
web> route template /t LICENSE
```
命令"route template /t LICENSE"，当web模块接收到请求uri为"/t"时把文件"LICENSE"中的内容作为响应回复。

* script脚本服务

content代表脚本的文件名，即web服务请求此route路径时，回复的内容为content指定的脚本运行后输出的内容。

## 3 context开发
### 3.0 context模块开发入门
在context目录下，创建目录src/example/demo，然后打开src/example/demo/demo.go文件，并输入以下代码。
```go
package demo

import (
	"context"
)

var Index = &ctx.Context{Name: "demo", Help: "example demo",
	Caches: map[string]*ctx.Cache{
		"format": &ctx.Cache{Name: "format", Value: "hello %s world", Help: "output string"},
	},
	Configs: map[string]*ctx.Config{
		"default": &ctx.Config{Name: "default", Value: "go", Help: "output string"},
	},
	Commands: map[string]*ctx.Command{
		"echo": &ctx.Command{
			Name: "echo word",
			Help: "echo something",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Echo(m.Cap("format"), m.Conf("default"))
			},
		},
	},
}

func init() {
	ctx.Index.Register(Index, nil)
}
```
在context目录下，打开src/example/bench.go文件，添加一行 _ "example/demo"，引入新添加的模块 。
```go
package main

import (
	"context"
	_ "context/aaa"
	_ "context/cli"
	_ "context/ssh"

	_ "context/mdb"
	_ "context/nfs"
	_ "context/tcp"
	_ "context/web"

	_ "context/lex"
	_ "context/log"
	_ "context/yac"

	_ "example/demo"

	"os"
)

func main() {
	ctx.Start(os.Args[1:]...)
}
```
在context目录下，编译安装bench.go，启动bench进入新模块，执行新添加的命令。
```sh
$ go install src/example/bench.go
$ bench
> ~demo
> echo
hello go world
```

#### 3.0.0 代码解析
```go
func init() {
	ctx.Index.Register(Index, nil)
}
```
在模块初始化时，向ctx模块注册当前模块，即当前模块为ctx的子模块。
```go
var Index = &ctx.Context{Name: "demo", Help :example demo",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{},
	Commands: map[string]*ctx.Commands{},
}
```
Index即为模块的数据结构，Name为模块的名字，Help为模块的简介，Caches为模块的缓存项，
Configs为模块的配置项，Commands为命令项。
```go
type Cache struct {
	Name  string
	Value string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}

type Config struct {
	Name  string
	Value string
	Help  string
	Hand  func(m *Message, x *Config, arg ...string) string
}

type Command struct {
	Name string
	Help string

	Formats map[string]int
	Options map[string]string
	Appends map[string]string
	Hand    func(m *Message, c *Context, key string, arg ...string)
}
```
Cache为缓存项的定义，Name为缓存项的名字，Value为缓存项的值，Help为缓存项的帮助信息，Hand为缓存项读写函数，可选。
Config为配置项的定义，Name为配置项的名字，Value为配置项的值，Help为配置项的帮助信息，Hand为配置项读写函数，可选。
Command为命令项的定义，Name为命令项的名字，Help为命令项的帮助信息，Hand为命令项执行函数。
```go
Commands: map[string]*ctx.Command{
	"echo": &ctx.Command{Name: "echo word", Help: "echo something", Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
		m.Echo(m.Cap("format"), m.Conf("default"))
	}},
},
```
命令Hand函数，是消息驱动的。
m.Cap()读写当前模块的某个缓存项。
m.Conf()读写当前模块的某个配置项。
m.Echo()输出命令执行结果。

#### 3.0.1 缓存接口
```go
func (m *Message) Caps(key string, arg ...bool) bool
func (m *Message) Capi(key string, arg ...int) int
func (m *Message) Cap(key string, arg ...string) string
```
只有一个参数时，代表读取缓存项的值。有两个参数是会向缓存项中写入新值。

Cap()把缓存项，当成字符串来进行直接读写。
Capi()只有一个参数时，读取缓存项并转换成整型数值返回。有两个参数时，会把第二个整型参数转换成字符串写缓存项中。
Caps()只有一个参数时，读取缓存项并转换成布尔值返回。有两个参数时，会把第二个布尔参数转换成字符串写缓存项中。

#### 3.0.2 配置接口
```go
func (m *Message) Confs(key string, arg ...bool) bool
func (m *Message) Confi(key string, arg ...int) int
func (m *Message) Conf(key string, arg ...string) string
```
只有一个参数时，代表读取配置项的值。有两个参数是会向配置项中写入新值。

Cap()把缓存项，当成字符串来进行直接读写。
Capi()只有一个参数时，读取配置项并转换成整型数值返回。有两个参数时，会把第二个整型参数转换成字符串写配置项中。
Caps()只有一个参数时，读取配置项并转换成布尔值返回。有两个参数时，会把第二个布尔参数转换成字符串写配置项中。

#### 3.0.3 日志接口
```go
func (m *Message) Log(action string, ctx *Context, str string, arg ...interface{})
```
参数action为日志类型，可取error, info, debug...类型及相关处理由log模块自定义。
参数ctx指定模块, 可取nil。
参数str与arg与Printf参数相同。

### 3.1 context模块开发进阶
#### 3.1.0 context模块的生命周期与继承多态
context结构体封封装了三种标准接口，cache、config、command。
但这些接口只能保存与处理简单的数据。如果需要处理复杂的数据，可以自己定义类型。
自定义的类型必须实现Server接口。
```go
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}
```
模块有自己的完整的生命周期函数，像对象，像进程。

Spawn()创建新模块，类似于面向对象的new。参数m为创建模块时的消息，c为新模块，arg为创建时的参数。返回值为自定义的结构体。

Begin()初始化新模块，类似于面向对象的构造函数。

Start()启动模块服务，类似于进程启动。

Close()关闭模块服务，类似于进程结束，类似于面向对象的析构函数。
```go
package demo

import (
	"context"
)

type Demo struct {
	*ctx.Context
}

func (demo *Demo) Spawn(m *ctx.Message, c *ctx.Context, arg ...string) ctx.Server {
	c.Caches = map[string]*ctx.Cache{
		"format": &ctx.Cache{Name: "format", Value: c.Name + ": hello %s world", Help: "output string"},
	}
	c.Configs = map[string]*ctx.Config{}

	d := new(Demo)
	d.Context = c
	return d
}
func (demo *Demo) Begin(m *ctx.Message, arg ...string) ctx.Server {
	return demo
}

func (demo *Demo) Start(m *ctx.Message, arg ...string) bool {
	return false
}

func (demo *Demo) Close(m *ctx.Message, arg ...string) bool {
	return true
}

var Index = &ctx.Context{Name: "demo", Help: "example demo",
	Caches: map[string]*ctx.Cache{
		"format": &ctx.Cache{Name: "format", Value: "hello %s world", Help: "output string"},
	},
	Configs: map[string]*ctx.Config{
		"default": &ctx.Config{Name: "default", Value: "go", Help: "output string"},
	},
	Commands: map[string]*ctx.Command{
		"echo": &ctx.Command{
			Name: "echo word",
			Help: "echo something",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				m.Echo(m.Cap("format"), m.Conf("default"))
			},
		},
		"new": &ctx.Command{
			Name: "new name",
			Help: "create new module",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				sub := c.Spawn(m, arg[0], "new module")
				sub.Begin(m)
				sub.Start(m)
			},
		},
		"del": &ctx.Command{
			Name: "del",
			Help: "delete module",
			Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
				c.Close(m)
			},
		},
	},
}

func init() {
	demo := &Demo{}
	demo.Context = Index
	ctx.Index.Register(Index, demo)
}
```
新定义了一个结构体Demo，可以向其添加各种自定义的数据。ctx.Context指向绑定的模块。
并实现了四个生命周期函数。Spawn()在创建新模块时会添加一个新缓存项。创建新的自定义结构体，并与新模块绑定。
init()创建新的自定义结构体，并与静态模块绑定。
```sh
$ go install src/example/bench.go
$ bench
cli> ~demo
demo> new one
one> context demo
demo> context
demo(ctx:cli:aaa:root:root): begin() example demo
one(demo:cli:aaa:root:root): start() new module
demo> new two
two> context demo
demo> context
demo(ctx:cli:aaa:root:root): begin() example demo
one(demo:cli:aaa:root:root): start() new module
two(demo:cli:aaa:root:root): start() new module
one> context one
one> echo
one: hello go world
```
重新编译并运行bench，进入demo模块，创新名为one的新模块，自动进入新模块。调用context demo切回原模块，调用context查看新建的模块。
调用context one切换到新建的模块调用echo。
新模块中没有echo命令，则会直接调用父模块的echo。
新模块中有format缓存项，则不会调用父模块的format。
新模块中没有default配置项，则会调用父模块的配置项。
从而实现了模块间的继承与多态。
```sh
one> del
one> context demo
demo> context
demo(ctx:cli:aaa:root:root): begin() example demo
two(demo:cli:aaa:root:root): start() new module
```
在one模块下调用del，则会删除当前模块。调用context demo切换到父模块，则看到one模块已经被删除。

#### 3.1.1 message消息驱动的开发模型
为了降低模块间的信赖关系，除了定义三种标准接口外，还有模块间的通信使用消息机制。
所以大多函数中都会有参数m代表当前的消息，arg消息的请求行参数。
```go
Commands: map[string]*ctx.Command{
	"send": &ctx.Command{
		Name: "send module",
		Help: "send something",
		Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			msg := m.Find(arg[0], false)
			msg.Cmd("echo")
			m.Echo(msg.Result(0))
		},
	},
	...
```
如上，再给demo模块添加一条send命令。
m.Find()会查找当前模块的子模块，并创建一个由当前模块发送给子模块的消息。
msg.Cmd()设置请求行，并执行命令。
msg.Result()取出响应行的数据。
m.Echo()设置当前命令的执行结果，即当前命令的响应行。
```sh
$ go install src/example/bench.go
$ bench
demo> new one
one> context demo
demo> send one
one: hello go world
```
重新编译并运行bench，在demo模块下，创建一个新的模块one，并切换回demo模块向one模块发送一条消息。

* 消息创建
```go
func (m *Message) Search(key string, root ...bool) []*Message
func (m *Message) Find(name string, root ...bool) *Message
func (m *Message) Sess(key string, arg ...string) *Message
```
Search()遍历功能树，找出模块名称或模块帮助信息匹配正则表达式key的模块，并创建由当前模块发送给这些模块的消息。

Find()遍历功能树，找出模块名匹配完整路径名的模块，并创建由当前模块发送给此模块的消息。name值为点分路径名，如"ctx.cli"代表ctx模块下的cli模块。

第二个参数root代表搜索的起点，默认为从根模块开始搜索。值为false时，从当前模块开始搜索。

Sess()是对Search()与Find()进行了封装，会把搜索的结果保存下来，以便之后的多次使用。
参数key为保存名，第二个参数为模块名，第三个参数为搜索方法search或find，第四个参数为搜索起点是否为根模块。
后面三个参数都有默认值，使用方法类似于C++的变参。如果只有一个参数则会直接取出之前的查找结果并创建一个新的消息。

* 消息发送
```go
func (m *Message) Cmd(arg ...interface{}) *Message
```
消息发送给目标模块，调用对应的Command接口。参数为命令行，包括命令名，与命令参数。

* 消息读写
```go
func (m *Message) Detail(index int, arg ...interface{}) string
func (m *Message) Detaili(index int, arg ...int) int
func (m *Message) Details(index int, arg ...bool) bool

func (m *Message) Result(index int, arg ...interface{}) string
func (m *Message) Resulti(index int, arg ...int) int
func (m *Message) Results(index int, arg ...bool) bool

func (m *Message) Option(key string, arg ...interface{}) string
func (m *Message) Optioni(key string, arg ...int) int
func (m *Message) Options(key string, arg ...bool) bool

func (m *Message) Append(key string, arg ...interface{}) string
func (m *Message) Appendi(key string, arg ...int) int
func (m *Message) Appends(key string, arg ...bool) bool
```
消息所带的数据类似于http报文。Detail()读写请求行（命令行）数据，Result()读写响应行。Option()读写请求头。Append()读写响应头。

请求行与响应行是由空格隔开的多个字符串，类似于字符串数组。Detail()与Result()第一个参数指定读写位置，第二个参数为要写入的数据，各种类型会转换为字符串插入到相应的位置。如果没有第二个参数，则代表读取对应位置上的数据。

请求头与响应头存储结构为map[string][]string，类似于字符串的数组结构体。Option()与Append()第一个参数为键值，指定要读写的数组，剩下的参数为要写入数组的多个数据，返回值为数组的第一个参数。如果没有第二个参数，则返回数组的第一个数据。

s结尾的函数代表读写的数据为bool值，函数内部会进行bool与string类型的相互转换。

i结尾的函数代表读写的数据为int值，函数内部会进行int与string类型的相互转换。

```go
Commands: map[string]*ctx.Command{
	"send": &ctx.Command{
		Name: "send module",
		Help: "send something",
		Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Sess("sub", arg[0], "find", "false")

			msg := m.Sess("sub")
			msg.Detail(0, "echo")
			msg.Option("key", "hello")

			msg.Cmd()

			m.Echo(msg.Result(0))
			m.Echo(" ")
			m.Echo(msg.Result(1))
			m.Echo(" ")
			m.Echo(msg.Append("hi"))
		},
	},
	"echo": &ctx.Command{
		Name: "echo",
		Help: "echo something",
		Hand: func(m *ctx.Message, c *ctx.Context, key string, arg ...string) {
			m.Result(0, m.Option("key"), "world")
			m.Append("hi", "nice")
		},
	},
	...
```
在send命令中， Sess()在一开始就搜索并保存了所需的模块，这些操作可以放在专门的初始化。不必每次都查找模块。
之后再调用Sess()创建新的消息，并调用Detail()设置请求行（命令行），调用Option()设置请求头。
然后调用Cmd()执行命令。命令执行完成后调用Result()取出响应行，调用Append()取出响应头。

在echo命令中，调用Option取出请求头，调用Result设置响应行，调用Append设置响应头。
```sh
$ go install src/example/bench.go
$ bench
demo> new one
one> context demo
demo> send one
hello world nice
```

## 4 context核心模块详解
### 4.0 ctx模块中心
### 4.1 cli命令中心
### 4.2 lex词法中心
### 4.3 yac语法中心
### 4.4 tcp网络中心
### 4.5 nfs存储中心
### 4.6 aaa认证中心
### 4.7 web应用中心
### 4.8 log日志中心
### 4.9 gdb调试中心
### 4.a ssh集群中心
### 4.b mdb数据中心

## 5 设计理念

## 数据结构
* ARM: 寻址与指令
* Linux: 文件与进程
* HTTP: 表示与会话


## 开发流程
* 设计: 协议与流程
* 编程: 接口与框架
* 测试: 语句与表达式

## 接口设计
* 功能树: Caches Configs Commands
* 消息树: Request History Session

### Context功能树
* Cap() Conf() Cmd()
* Spawn() Begin() Start() Close()

### Message消息树
* Detail() Option() Result() Append()
* Req() His() Sess()

## 模块设计
* 应用层 ctx cli aaa web
* 控制层 lex yac log gdb
* 数据层 tcp nfs ssh mdb

### 应用层
* ctx: 模块中心
* cli: 管理中心
* aaa: 认证中心
* web: 应用中心

### 控制层
* lex: 词法中心
* yac: 语法中心
* log: 日志中心
* gdb: 调试中心

### 数据层
* tcp: 网络中心
* nfs: 存储中心
* ssh: 集群中心
* mdb: 数据中心

