## context

context是一种新的编程语言与应用框架，通过模块化、集群化、自动化，实现软件的快速开发，快速共享，快速使用。

context是以群聊的形式，进行资源的共享。
用户可以创建任意的群聊，把相关人员聚集在一起，每个人可以将自己的设备，共享到群聊中，供组员使用，从而实现资源的最大利用。
每个设备上有一堆命令，用户可以将任意设备上任意命令，添加到自定义的应用界面中，按照自己的需求去组合，从面实现场景化与个性化的定制。
所以每个群聊中会有各种各样自定义的应用，所有的命令都是以群聊作为场景，进行权限的检查与分配。
这些应用，可以像文本与图片一样，在群聊里自由的流动，可以被更快分享出去，再次收藏与组合形成新的应用组件，还可以在聊天记录中直接使用。

context是以分布式的方式，进行程序的开发。
开发者，可以用脚本语言开发应用，随时随地的在自己任意设备上加载脚本，然后将动态域名分享出去，应用就可以被用户在群聊中任意的传播。
所有的代码与数据，都在自己设备上，可以进行任意的实时控制。
消灭所有中间环节，让几行代码的小函数，就可以成为独立的应用，从而实现软件的快速开发与快速传播，将任意一行代码的价值，放大成千上万倍。

## 下载安装
### 下载
在Linux或Mac上，可以直接用命令下载，
在Windows上，推荐先安装 [GitBash](https://www.git-scm.com/download/)，然后在GitBash中执行命令下载。
```
$ export ctx_dev=https://shylinux.com; curl $ctx_dev/publish/boot.sh | bash -s install context
```
*install后面的参数context，就是指定的下载目录，如不指定，会把相关文件下载到当前目录。*

*ctx_dev环境变量指定服务器地址，所以可以自行搭建服务器。*

### 启动
下载完成后，会自动启动context，
windows下的GitBash中，如果自动启动失败，则需要手动启动一下，如下命令。
```
$ cd context && bin/boot.sh
```

### 使用
启动后context，提供了一种交互式的shell，直接可以执行各种内部命令和本地命令。
如下查看当前目录与相关目录下的文件。
```
0[22:21:19]nfs> pwd
/home/homework/context

1[22:21:20]nfs> dir
time                size line path
2019-09-12 22:21:18 103  5    bin/
2019-09-12 22:20:40 72   3    etc/
2019-09-12 22:20:40 55   3    var/
2019-09-12 22:21:18 50   2    usr/

2[20:51:21]nfs> dir bin
time                size     line path
2019-09-16 20:51:14 18782016 5209 bin/bench
2019-09-16 20:51:14 2634     99   bin/boot.sh
2019-09-16 20:51:14 125      5    bin/node.sh
2019-09-16 20:51:14 96       6    bin/user.sh
2019-09-16 20:51:14 147      9    bin/zone.sh

3[20:51:22]nfs> dir etc
time                size line path
2019-09-16 20:51:14 339  11   etc/common.shy
2019-09-16 20:51:14 244  11   etc/exit.shy
2019-09-16 20:51:14 297  18   etc/init.shy

4[22:21:20]nfs>
```
- bin目录，就是各种启动脚本与命令
- etc目录，就是各种配置脚本
- var目录，就是各种输出文件，如日志与缓存文件
- usr目录，就是各种前端文件与数据，如js、css文件

*如需要自行启动context，必须在当前目录，然后运行bin/boot.sh脚本。否则会找不到相关文件。*

## 基本功能

除了命令行交互，还可以访问<http://localhost:9095>，用浏览器进行操作。
context启动后，默认监听9095端口，启动网页服务。

进入下载目录，可以看到的有八个文件。

在bin目录下，就是各种执行文件

- bin/bench，context的执行程序
- bin/boot.sh，context的启动脚本
- bin/zone.sh，启动区域节点
- bin/user.sh，启动用户节点
- bin/node.sh，启动工作节点

context内部实现了语法解析，通过自定义的脚本语言，实现功能的灵活控制。

在etc目录下，就是context执行过程中用到的脚本。

- etc/init.shy，启动时加载的脚本
- etc/exit.shy，结束时运行的脚本
- etc/common.shy，init.shy调用到的脚本

## 创建集群
context是一种分布式框架，可以运行在任意设备上，并且实现了自动组网、自动路由、自动认证。
远程命令与本地命令，无差别的运行，从而实现无限扩容的分布式计算。

context每个启动的进程都是一个独立的节点，根据网络框架中的功能作用，可以分为区域节点、用户节点、工作节点、分机节点。
这几种节点，除了网络框架中的作用外，其它的功能模块与命令都完全一样，没有差别。

个人使用，可以创建一个区域节点，下挂多个工作节点。

团队使用，需要创建一个区域节点，多个用户节点，每个用户节点下，可以挂多个工作节点。

如果用户节点或工作节点过多，可以创建分机节点，通过增加层级来降低单机负载。

### 个人使用
#### 启动区域节点
打开终端，进入context目录，执行如下命令，
```
$ bin/zone.sh
0[13:26:27]nfs>
```

#### 启动工作节点
再打开终端，进入context目录，执行如下命令，
```
$ bin/node.sh create app/hello

0[13:26:27]nfs> remote
create_time         pod type  
2019-07-30 13:26:27 com master
```
启动context后，调用remote命令，可以查看到有一个上级节点。

#### 启动工作节点
再打开终端，进入context目录，执行如下命令，
```
$ bin/node.sh create app/world

0[13:26:27]nfs> remote
create_time         pod type  
2019-07-30 13:26:27 com master
```

#### 分布式命令
启动两种节点节点后，就可以在任意节点上调用命令，也可以调用远程节点的命令。
如在区域节点上调用remote，就可以看到两个工作节点。
```
4[13:27:26]nfs> remote
create_time         pod    type  
2019-07-30 13:26:27 hello  worker
2019-07-30 13:26:30 world  worker
```

查看当前路径
```
3[13:39:29]nfs> pwd
D:\context/var
4[13:40:03]nfs> 
```

查看当时目录
```
4[13:40:03]nfs> dir
time                size line path      
2019-07-23 21:36:36 387  4    var/hi.png
2019-07-27 13:41:56 4096 4    var/log/  
2019-06-15 10:58:03 0    1    var/run/  
2019-07-30 12:55:19 4096 8    var/tmp/  
5[13:40:20]nfs> 
```

执行远程命令，只需要在命令前加上节点名与冒号。
```
6[13:41:28]nfs> hello:pwd
D:\context\hello/var
6[13:41:28]nfs> world:pwd
D:\context\world/var
```

在任意随机节点上执行命令，用百分号作节点名。
```
5[13:40:20]nfs> %:pwd
D:\context\hello/var
5[13:40:20]nfs> %:pwd
D:\context\world/var
```

在所有节点上执行命令，用星号作节点名。
```
7[13:41:36]nfs> *:pwd
D:\context\hello/var D:\context\hello/var
```

## 团队使用
context也可以支持团队协作使用，这时候就需要将区域节点部署到公共主机上。
区域节点的作用就是生成动态域名，分发路由，解决命名冲突，与权限分配等功能。

### 启动用户节点
在公共主机上启动区域节点后，每个组员就可以在自己主机上启动用户节点，但需要指定区域节点的地址。
如下命令，ip换成自己的公共主机，9095端口保留，这是context默认的web端口。
```
$ ctx_dev=http://192.168.88.102:9095 bin/user.sh
```

### 启动工作节点
同样每个用户都可以启动多用工作节点。
```
$ bin/node.sh create world
```

### 启动团队协作
当有多个用户连接到公共节点后，用户与用户之间就可以相互访问对方的所有节点。
但是默认启用了节点认证，所有命令都没有权限。所以调用对应节点上的命令，需要对方开启命令权限。

每个用户随时都可以在自己节点上，为其它用户设置任意角色，给每个角色分配任意命令。
从而实现安全快速的资源共享。


## 启动分机节点
当区域的用户节点过多，就可以启动分机节点。
启动分机节点，只需要指定上级节点即可。
用户在连接公共节点时，指定这个新节点的ip即可。
context会自动生成新的网络路由。

```
$ ctx_dev=http://192.168.88.102:9095 bin/boot.sh
```

## 创建群聊
除了命令行的使用的方式之外，context还有自己的前端框架。
不仅降低了使用难度，还提供更加场景化、自动化的应用界面。

用户可以访问区域节点或是任意用户节点的网页服务。

http://127.0.0.29:9095

输入用户名与初始密码，即可登录，如果用户与主机上的用户名相同，则是管理员权限，如果不同，则是普通用户，只有最小的功能权限。

所以任意节点都支持多用户共享使用，但只有管理用户有所有权限，进行资源的管理与分配。

打开应用界面，就可以看到context以办公聊天软件的形式提供各种丰富的功能。

左边框是用户群组列表，用户可以选择群聊或是创建新的群聊。
中间就是聊天记录与输入框，用户可以自由的聊天收发消息。

与其它聊天软件不同的是，context提供了自定义的功能列表。
右边框中，就是此群组的功能列表。
每个用户都可以将自己设备上的命令添加到这个群组的功能列表中，分享给本组员使用。
每个组员都可以根据自己的需求组合这些命令，生成自己的应用界面。

以每个群聊作为场景，进行资源的共享与应用的开发，从而实现更加场景化与个性化应用。

通过这种精细化的应用场景，进行工具化、标准化、流程化。提高各行各业的工作效率。

## 应用开发

网络框架与应用界面，已经实现了标准化与自动化，剩下就是应用的开发了。

开发者，可以在任意机器上开发自己的应用。以模块与函数为单位进行开发与上线。
一个函数，即使只有几行代码，也是一个独立完整的应用，可以随时上线，被任意用户使用。

用户还可以在任意群聊中转发此应用，更自由的传播出去。
从而将软件开发的速度提升成千上万倍，将代码的使用效率提升成千上万倍。

### 创建项目
在任意节点上，执行project命令，指定项目名，即可创建应用目录。
```
$ bin/user.sh
8[13:41:41]nfs> project hello
time                line hash     path                     
2019-07-30 14:27:09 35   eba8eda2 src/plugin/hello/index.go 
2019-07-30 14:27:09 1    b858cb28 src/plugin/hello/index.shy
2019-07-30 14:27:09 1    b858cb28 src/plugin/hello/local.shy
2019-07-30 14:27:09 4    407265b6 src/plugin/hello/index.js 
9[14:27:09]nfs> 
```
每个项目，都可以用go语言开发低层应用，用js开发前端交互。
除此，context有自己的通用语法解析器，开者完全可以随时自定义语法，定制自己的解析器。
用自己喜欢的语法开发应用。

默认的shy语法，提供了一个完整的前后端应用框架。创建项目时，自动创建的模块如下。
```
fun hello world "" "" \
	public \
	text "" \
	button "执行"
	copy pwd
end
```

这个模板就是一个完整的应用，fun关键字开关，end关键字结束。
前四行就是定义应用界面，剩下代码就是后端脚本。
public代表，这个应用是公共的，所以有人都可以访问。也可以是private，只有管理用户可以访问。

text与button，就是需要前端展示的控件。用户在前端点击此button，就会将请求发送到后端，执行此脚本。
然后将执行结果返回给前端界面。

### 加载项目
切换到cli模块，使用upgrade命令，加载新的项目应用。
```
8[13:41:41]nfs> ~cli
8[13:41:41]cli> upgrade plugin hello
3[15:55:53]cli> ~
names    ctx msg  status stream helps   
ctx          0    start  stdio  模块中心
cli      ctx 4    begin         管理中心
hello    cli 5958 start         shy     
4[15:57:00]cli> 
```

切换到hello模块，使用command命令，可以查看到hello模块下的命令列表，然后就可以调用hello命令。
```
8[13:41:41]cli> ~hello
6[15:58:42]hello> command
key   name                                  
hello hello world   public text  button 执行
7[20:27:32]nice> hello
D:\context/var
```

同时在前端界面上添加功能，即可看到此函数。

context内部实现了很多功能模块，每个模块下有很多命令，每条命令就是一种应用。

context的使用方式有很多种，

- 可以直接调用，像Shell一样，去解析一条命令
- 可以启动cli服务，像MySQL一样，交互式使用格式化命令
- 可以启动web服务，像LabView一样，可以自定义各种图形界面
- 可以自动组网，将任意台设备组合在一起，实现分布式应用
- 可以自动建群，在群聊场景中，实现多用户、多会话、多任务、多设备的使用

### 命令模式
如果只是使用一条命令，或是写在脚本文件中，可以使用这种方式。

例如，dir命令就是查看目录，
```
$ bin/bench dir
time                 size  line  filename
2019-06-16 10:35:18  324   11    common.shy
2019-06-16 10:35:18  201   9     exit.shy
2019-06-16 10:35:18  261   13    init.shy
```

还可以加更多参数，dir_deep递归查询目录，dir_type文件类型过滤，dir_sort输出表排序。
```
$ bin/bench dir ../ dir_deep dir_type file dir_sort line int_r
time                 size      line   filename
2019-06-16 10:22:52  13256968  91314  bench
2019-06-16 11:10:16  1535      66     boot.sh
2019-06-16 11:10:16  613       31     node.sh
2019-06-16 11:10:16  261       13     init.shy
2019-06-16 11:10:16  324       11     common.shy
2019-06-16 11:10:16  201       9      exit.shy

```

### 交互模式

启动服务，可以提供更丰富的命令与环境。
```
$ bin/bench
0[11:35:46]ssh> dir
time                 size  line  filename
2019-06-16 11:35:06  160   3     log/
2019-06-16 11:35:06  96    1     run/
2019-06-16 11:35:44  192   4     tmp/
1[11:35:46]ssh>
```

如果集中管理，命令越多，系统只会越复杂，学习成本越高，使用越低效，开发越困难。

所以通过模块化，分而治之，更高效的管理丰富的命令。

context命令就是用来管理模块，没有参数时，直接查看当前模块的信息。

如下，第二行是当前模块，第一行是当前模块的父模块，其它行都是当前模块的子模块。

```
1[11:39:01]ssh> context
names  ctx  msg  status  stream         helps
ctx         0    start   shy            模块中心
ssh    ctx  10   begin   ctx.nfs.file3  集群中心
```

context第一个参数，可以指定当前模块，
如下，切换到nfs模块，然后查看各种IO模块，
切换到ctx根模块，查看所有模块。
```
2[11:43:57]ssh> context nfs

3[11:43:58]nfs> context
names  ctx  msg   status  stream  helps
ctx         0     start   shy     模块中心
nfs    ctx  9     begin           存储中心
stdio  nfs  1174  start   stdio   scan stdio

4[11:44:22]ssh> context ctx

5[11:45:17]ctx> context
names    ctx  msg   status  stream    helps
ctx           0     start   shy       模块中心
aaa      ctx  3     begin             认证中心
cli      ctx  4     begin             管理中心
gdb      ctx  232   start             调试中心
lex      ctx  6     begin             词法中心
log      ctx  31    start   bench     日志中心
mdb      ctx  8     begin             数据中心
nfs      ctx  9     begin             存储中心
ssh      ctx  10    begin             集群中心
tcp      ctx  11    begin             网络中心
web      ctx  1094  start   :9094     应用中心
yac      ctx  13    begin   35,14,23  语法中心
shy      cli  1171  start   engine    shell
matrix1  lex  34    start   76,28,2   matrix
stdio    nfs  1174  start   stdio     scan stdio
chat     web  14    begin             会议中心
code     web  15    begin             代码中心
wiki     web  16    begin             文档中心
engine   yac  1173  start   stdio     parse

```

command命令，就是用来管理当前模块的命令，
```
17[11:52:02]nfs> context nfs

17[11:52:02]nfs> command
key     name
_init   _init
action  action cmd
copy    copy to from
dir     dir [path [fields...]]
export  export filename
git     git sum
hash    hash filename
import  import filename [index]
json    json str
load    load file [buf_size [pos]]
open    open file
path    path filename
printf  printf arg
prompt  prompt arg
pwd     pwd [all] | [[index] path]
read    read [buf_size [pos]]
remote  remote listen|dial args...
save    save file string...
scan    scan file name
send    send [file] args...
temp    temp data
term    term action args...
trash   trash file
write   write string [pos]

```

help子命令，查看命令帮助信息。
```
18[11:59:19]nfs> command help dir
dir: dir [path [fields...]]
    查看目录, path: 路径, fields...: 查询字段, time|type|full|path|tree|filename|size|line|hash
    dir_deep: 递归查询
    dir_type both|file|dir|all: 文件类型
    dir_reg reg: 正则表达式
    dir_sort field order: 排序
```


### 集群模式

context提供自动化集群的功能，可以自动组网、自动认证。从而快速实现多台设备的协同工作。

#### 启动服务节点
```
$ bin/boot.sh
0[11:35:12]ssh>
```

#### 启动工作节点

新打开一个终端，启动工作节点，执行remote命令，查看上级节点，
```
$ bin/boot.sh create app/demo
0[15:15:30]ssh> remote
key  type    module         create_time
mac  master  ctx.nfs.file3  2019-06-16 15:15:23
```

回到服务节点终端，执行remote命令，可以查看到所有远程节点。
```
2[15:15:31]ssh> remote
key   type    module         create_time
com   master  ctx.nfs.file4  2019-06-16 14:25:10
demo  worker  ctx.nfs.file7  2019-06-16 15:15:23
```

默认配置中，子节点信任父，所以父节点可以调用子节点的命令。还有更复杂的认证机制，可以灵活配置。

远程命令和本地命令一样，没有任何区别。如下调用demo节点的pwd命令。还支持更复杂的多节点命令，可以更快速的同时管理多台设备。
```
2[15:15:31]ssh> remote demo pwd
/Users/shaoying/context/app/demo/var
```

#### 启动分机节点

在服务节点的终端，查看服务地址
```
3[15:49:00]ssh> web.brow
index  site
0      http://192.168.199.139:9094
```

同样，在另一台设备上下载context，然后启动服务节点。
通过环境变量ctx_dev指定上级节点。
```
$ ctx_dev="http://192.168.199.139:9094" bin/boot.sh
0[15:49:00]ssh> remote
key  type    module         create_time
mac  master  ctx.nfs.file3  2019-06-16 15:15:23
```

回到服务节点终端，执行remote命令，可以查看到新添加了一个服务子节点。
```
2[15:15:31]ssh> remote
key   type    module         create_time
com   master  ctx.nfs.file4  2019-06-16 14:25:10
demo  worker  ctx.nfs.file7  2019-06-16 15:15:23
sub   server  ctx.nfs.file8  2019-06-16 16:15:23
```

同样可以远程调用命令。
```
2[15:15:31]ssh> remote sub pwd
/Users/shaoying/context/app/sub/var
```

#### 文件管理
调用本地命令，查看当前目录下的文件列表，
```
0[10:53:25]ssh> ls
total 0
drwxr-xr-x  5 shaoying  staff  160 May 10 03:57 bin
drwxr-xr-x  5 shaoying  staff  160 May 10 03:30 etc
drwxr-xr-x  2 shaoying  staff   64 May 10 03:30 usr
drwxr-xr-x  8 shaoying  staff  256 May 10 03:36 var
```

调用内部命令，查看文件列表，如下dir与ls命令用途相似，但提供了更丰富的功能，如统计文件行数
```
5[10:56:24]ssh> dir etc
time                 size  line  filename
2019-04-14 21:29:21  316   10    common.shy
2019-04-29 21:12:28  130   7     exit.shy
2019-04-29 21:12:12  191   12    init.shy
```

"%"是一个内部命令，可以对前一步命令结果进行各种处理。如下按字段line排序
```
5[10:56:24]ssh> dir etc % order line
time                 size  line  filename
2019-04-29 21:12:12  191   12    init.shy
2019-04-14 21:29:21  316   10    common.shy
2019-04-29 21:12:28  130   7     exit.shy
```

如下按字段""聚合，即可得到汇总结果，所有文件的总字节数、总行数、总文件数
```
16[11:04:30]ssh> dir etc % group ""
time                 size  line  filename    count
2019-04-14 21:29:21  637   29    common.shy  3
```

#### 时间管理
查看当前时间戳
```
18[11:11:01]ssh> time
1557457862000
```
将时间戳转换成日期
```
19[11:11:14]ssh> time 1557457862000
2019-05-10 11:11:02
```
将日期转换成时间戳
```
20[11:11:25]ssh> time "2019-05-10 11:11:02"
1557457862000
```

#### 网卡管理
```
2[10:53:25]ssh> ifconfig
index  name  ip             mask  hard
5      en0   192.168.0.106  24    c4:b3:01:cf:0b:51
```

### 完整版
命令行模式，只是context最基本的功能。context还提供了一个前端框架，
让用户可以自由的组合功能列表，满足自己的需求。

使用更丰富的功能，可以直接下载源码，
```
$ git clone https://github.com/shylinux/context.git
```
下载完源码后，如果安装了golang，就可以对源码直接进行编译，
如果没有可以去官网下载安装golang（<https://golang.org/dl/>）。
第一次make时，会自动下载各种依赖库，所以会慢一些。
```
$ cd context && make
```

#### 启动服务
bin目录下是各种可执行文件，如启动脚本boot.sh与node.sh。

node.sh用来启动单机版context，boot.sh用来启动网络版context。

如下，直接运行脚本，即可启动context。

启动context后，就可以解析执行各种命令，即可是本地的shell命令，也可以是内部模块命令。

```
$ bin/node.sh
0[03:58:43]ssh>
```

#### 网页服务

下载完整版的context，启动的服务节点，就会带有前端网页服务。

#### 知识库

wiki模块提供了知识库管理。

访问：http://localhost:9094/wiki

wiki模块会将usr/wiki目录下的md文件进行解析，生成网页文件，并自动生成目录与索引。

可以创建自己的知识库，如下创建目录与文件。

```
$ mkdir -p usr/wiki/some
$ echo "hello world" > usr/wiki/some/hi.md
```

然后在服务终端上，切换到新建的知识库，
```
0[10:53:25]ssh> ~wiki config wiki_level wiki/some
```

也可以加到启动文件中，
```
$ cat etc/local.shy
~wiki
    config wiki_level wiki/some
```

#### 用户界面

chat模块提供了完整的功能系统。

在一台公共的设备上启动服务节点，在etc/local.shy的启动脚本中加入如下两行命令。此设备就可以成为公共服务器。
```
~ssh
    work serve
```

在任意其它设备上，每个用户都可以启动自己的节点，在启动脚本etc/local.shy，添加如下两行代码，此设备就可以成为用户的主控节点。
```
~ssh
    work create
```

在启动用户节点时，只要指定服务节点，就可以实现用户注册，自动加入群组。
```
    ctx_dev=http://172.0.0.172 bin/boot.sh
```

然后登录 http://172.0.0.172:9095/chat 或 http://127.0.0.127:9095/chat ，
输入自己的用户名，与初始密码，就可以登录系统。

然后就可以自由的创建群聊、共享设备、实时聊天。

#### 模块开发
```
project hello
compile hello
publish hello
```

### 所有目录

#### 一级目录

- src
- etc
- bin
- var
- usr

#### 源码目录

- src/toolkit
- src/context
- src/example
- src/plugin
 
#### 配置文件

- etc/init.shy
- etc/common.shy
- etc/exit.shy

#### 执行文件

- bin/boot.sh
- bin/node.sh
- bin/bench
 
#### 日志文件

- var/log/boot.log
- var/log/error.log
- var/log/right.log
- var/log/bench.log
- var/log/debug.log
- var/run/bench.pid
- var/run/user/cert.pem
- var/run/user/key.pem
- var/run/node/cert.pem
- var/run/node/key.pem
- var/tmp/runtime.json
- var/tmp/auth.json
 
#### 应用目录

- usr/template
- usr/librarys
- usr/upgrade
- usr/client

## 应用开发
- Windows
- Mac
- pi
- mp
- iOS
- Android

### 应用接口

context的应用模块都是web的子模块，在web模块启动HTTP服务后，会根据模块名与命令名自动生成路由。
web模块会将所有的HTTP请求转换成context的命令调用，所以HTTP的应用接口和普通命令，除了名字必须以"/"开头，其它没有太大区别。

当web接收到HTTP请求后，可以调用单个命令如 http://shylinux.com/code/consul 就会调用code模块下的/consul命令

可以调用多个命令如 http://shylinux.com/code/?componet_group=login 就会调用web模块下的/render命令，
根据code的componet下的login组件，依次调用每个接口的命令，然后将执行结果与参数一起，调用golang的template，渲染生成HTML。

所有命令都解析完成后就可以生成一个完整的网页。当然如果Accept是application/json，则会跳过模块渲染，直接返回多条命令的执行结果。
所以componet就是接口的集合，统一提供参数配置、权限检查、命令执行、模板渲染，前端展示样式，前端初始化函数，降低内部命令与外部应用的耦合性，但又将前后端完全融合在一起。

如下，是web.code模块的应用接口定义。配置componet下定义了多个组件，每个组件下定义了多个接口。

login就是登录页面，下面定义了三个接口code、login、tail，
其中code，使用模板head生成网页头，会包括一些配置，如favicon可以指定图标文件，styles指定引用模式表。
其中tail，使用模板tail生成网页尾，会包括一些配置，如scripts指定引用脚本文件。
login就是网页组件了，生成一个网页登录的输入表单，并接收表单请求调用aaa模块的auth命令，进行用户身份的验证。
其中arguments指定了Form表单字段的列表。
```
...
var Index = &ctx.Context{Name: "code", Help: "代码中心",
	Caches: map[string]*ctx.Cache{},
	Configs: map[string]*ctx.Config{
		"skip_login": &ctx.Config{Name: "skip_login", Value: map[string]interface{}{"/consul": "true"}, Help: "免密登录"},
        "componet": &ctx.Config{Name: "componet", Value: map[string]interface{}{
            "login": []interface{}{
                map[string]interface{}{"componet_name": "code", "componet_tmpl": "head", "metas": []interface{}{
                    map[string]interface{}{"name": "viewport", "content": "width=device-width, initial-scale=0.7, user-scalable=no"},
                }, "favicon": "favicon.ico", "styles": []interface{}{"example.css", "code.css"}},

                map[string]interface{}{"componet_name": "login", "componet_help": "login", "componet_tmpl": "componet",
                    "componet_ctx": "aaa", "componet_cmd": "auth", "componet_args": []interface{}{"@sessid", "ship", "username", "@username", "password", "@password"}, "inputs": []interface{}{
                        map[string]interface{}{"type": "text", "name": "username", "value": "", "label": "username"},
                        map[string]interface{}{"type": "password", "name": "password", "value": "", "label": "password"},
                        map[string]interface{}{"type": "button", "value": "login"},
                    },
                    "display_append": "", "display_result": "",
                },

                map[string]interface{}{"componet_name": "tail", "componet_tmpl": "tail",
                    "scripts": []interface{}{"toolkit.js", "context.js", "example.js", "code.js"},
                },
            },
...
```
### 网页开发

#### 模板

usr/template 存放了网页的模板文件，context会调用golang的template接口进行后端渲染，生成html文件。
不同的应用模块都会有自己的模板目录，也有公共模板库。

- usr/template/common.tmpl 公共模板
- usr/template/code/ code模块的模板
- usr/template/wiki/ wiki模块的模板
- usr/template/chat/ chat模块的模板

#### 样式

所有的css都存放usr/librarys

- example.css
- code.css
- wiki.css
- chat.css

#### 脚本

所有的js都存放usr/librarys

- toolkit.js 工具库，主要是网页相关的操作，如AppendChild
- context.js 通信库，主要是用来与后端context进行通信
- example.js 框架库，统一定义了网页的框架，每个应用网页都会继承
- code.js 工具链应用的网页
- wiki.js 知识库应用的网页
- chat.js 信息流应用的网页

### 小程序
### 开发板

## 接口开发
### componet
### python
### java
### c

## 模块开发
### 应用模块
#### 简单模块
#### 复杂模块
#### 脚本模块
### 插件模块
#### 独立插件
#### 扩展插件
### 核心模块
#### 模块中心ctx
#### 命令中心cli
#### 认证中心aaa
#### 应用中心web
#### 网络中心tcp
#### 存储中心nfs
#### 集群中心ssh
#### 数据中心mdb

## 系统架构

| |数据流|命令流|权限流|应用流|
|---|---|---|---|---|
|应用层|ctx|cli|aaa|web|
|控制层|lex|yac|log|gdb|
|数据层|tcp|nfs|ssh|mdb|

### 应用框架
#### 模块

context内部使用模块组织功能，每个模块都可以独立编译，独立运行。
解除了代码之间的包依赖、库依赖、引用依赖、调用依赖。
通过map查找模块，通过map查找命令，通过map查找配置，从而实现完全自由的模块。

**模块定义：**
```
type Context struct { // src/contexts/ctx/ctx.go
    Name string
    Help string

    Caches   map[string]*Cache
    Configs  map[string]*Config
    Commands map[string]*Command

    ...

    contexts map[string]*Context
    context  *Context
    root     *Context

    ...
    Server
}

```
Name：模块名称，Help：模块帮助。
模糊搜索搜索时，会根据Name与Help进行匹配。

每个模块会有命令集合Commands，配置集合Configs，缓存集合Caches。通过这种形式提供功能集合。

contexts：所有子模块，context：指向父模块，root：指向根模块。
从而组成一个模块树，所以可以通过路由查找模块，
如ctx.web.code，code的父模块是web，web的父模块是ctx，ctx是根模块。
所以可以通过命令，查看到当前程序所有模块的信息。

**缓存定义：**
```
type Cache struct {
	Value string
	Name  string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}
```
Value：存放的数据，Name：变量名称，Help：变量帮助，Hand读写函数。

缓存数量是一种数据接口，用来存放一些状态量，向外部显示程序进行状态，对外部来说一般是只读的。对内部来说可读可写。
所以可以通过命令，查看到当前程序任意模块的状态数据。

如下，ncontext当前有多少个模块。nserver有多少个模块运行了守护协程。

```
"nserver":    &Cache{Name: "nserver", Value: "0", Help: "服务数量"},
"ncontext":   &Cache{Name: "ncontext", Value: "0", Help: "模块数量"},
```

**缓存读写：**
```
func (m *Message) Cap(key string, arg ...interface{}) string {}
func (m *Message) Capi(key string, arg ...interface{}) int {}
func (m *Message) Caps(key string, arg ...interface{}) bool {}
```
定义了缓存数据的三种读写接口。

m.Cap()只有一个参数时，会从当前模块查询缓存变量，如果查到则返回Value，如果没有，则依次查询父模块。
如果查找到根模块还没有查到找，变返回空字符串。

m.Cap()有两个参数时，同样会从当前模块依次查询父模块，直到查到变量，然后设置其值。

m.Capi()是对m.Cap()封装了一下，在int与str相互转换，从而实现用str存储int。
所有转换失败的数据，都会返回0。

m.Caps()，实现了str存储bool。返回false的值有"", "0", "false", "off", "no", "error: "，其它都返回true。

**配置定义：**
```
type Config struct {
	Value interface{}
	Name  string
	Help  string
	Hand  func(m *Message, x *Cache, arg ...string) string
}
```
Value：存放的数据，Name：变量名称，Help：变量帮助，Hand读写函数。
与Cache相似，只是Value的类型不再是String而是interface{}，所以可以用来存放更复杂的数据。

一般用来存放配置数据，是外部控制内部数据接口。
所以可以通过命令，实时修改当前程序任意模块的配置数据。
避免只是修改某个配置变量，就要重启整个进程，从而实现高效灵活的配置。
把每个进程当成一个生命来对待，不要轻易杀死任何一个进程。有问题可以用微创手术解决。

**配置读写：**
```
func (m *Message) Conf(key string, args ...interface{}) string {}
func (m *Message) Confi(key string, arg ...interface{}) int {}
func (m *Message) Confs(key string, arg ...interface{}) bool {}
func (m *Message) Confx(key string, args ...interface{}) string {}
func (m *Message) Confv(key string, args ...interface{}) interface{} {}
func (m *Message) Confm(key string, args ...interface{}) map[string]interface{} {}
```
与Cache相似，也定义了各种读写的接口。

因为interface{}可以是任意复合类型，所以数据嵌套很深时，查询会涉及各种类型转换，非常麻烦。
Conf()定义了键值链。内部去处理类型转换与嵌套的深入。 
如m.Conf("runtime", "user.node")、m.Conf("runtime", []string{"user", "node"})、m.Conf("runtime", []interface{}{"user", "node"})
都会查询配置runtime下的user下的node的值。

m.Confx()内部进行选择，如果m.Option(key)中取到了值，则直接返回m.Option(key)，否则返回m.Conf(key)。
把配置当成一个备用的默认值，如果命令参数设置了此参数，则用命令中的参数。

m.Confv()直接读写原始数据。
m.Confm()则定义了更丰富的接口，m就是map意思，直接返回map[string]interface{}
m另外一个意思就是magic，可以传入各种回调函数。
如下配置node类型是map[string]interface{}，m.Confm()会遍历此map，查到value也是map[string]interface{}的键值，调用回调函数。
```
    ...
    m.Confm("node", func(name string, node map[string]interface{}) {
        if kit.Format(node["type"]) != "master" {
            ps = append(ps, kit.Format(node["module"]))
        }
    })
    ...
```

**命令定义：**
```
type Command struct {
	Form map[string]int
	Name string
	Help interface{}
	Auto func(m *Message, c *Context, key string, arg ...string) (ok bool)
	Hand func(m *Message, c *Context, key string, arg ...string) (e error)
}
```
Name：命令语法，Help：命令帮助。

Hand：命令处理函数，m是调用消息，c是当前模块，key是命令名，arg是命令参数。

在命令解析时，会根据Form将[key value...]形式的参数，取出存放到m.Option中，方便用key直接查找参数。
所以arg中只剩下序列参数，通过index序号查找参数。

Auto：终端自动补全函数。在使用终端每输入一个单词时，就调用此函数输出提示信息。所以在命令执行前，这个函数会被调用多次。

如下定义了trans命令
```
    ...
    "trans": &Command{Name: "trans option [type|data|json] limit 10 [index...]", Help: "数据转换",
        Form: map[string]int{"format": 1, "fields": -1},
        Hand: func(m *Message, c *Context, key string, arg ...string) (e error) {
            ...
        }}
    ...
```

**命令调用：**
```
func (m *Message) Cmd(args ...interface{}) *Message {}
func (m *Message) Cmdx(args ...interface{}) string {}
func (m *Message) Cmds(args ...interface{}) bool {}
func (m *Message) Cmdy(args ...interface{}) *Message {}
func (m *Message) Cmdm(args ...interface{}) *Message {}
```
m.Cmd()根据第一个参数去当前查找命令，如果没有查找到，则去父模块查找。如果没有查找，则不会执行。
剩下的参数会根据Form定义来解析，存放到m.Option中。

m.Cmds()当返回值转换成bool规则同m.Caps()与m.Confs()。
m.Cmdx()当返回值转换成str。 m.Cmdy()将结果复制到当前Message。

m.Cmdm()，m同样是magic，会根据当前会话，自动定向到远程某主机某模块，远程调用其命令，当然也可能定向到本机。

#### 协程
#### 消息

context内部调用都是


### 解析引擎
#### 文件扫描
#### 词法解析
#### 语法解析
#### 执行命令
### 通信框架
#### 节点路由
每个节点在启动时，自动向上级注册，生成一个动态域名，作为本节点的地址。
如com.mac.led，led的上级节点是mac，mac的上级节点是com。
其它节点，就可以通过个这个地址查找到此节点。

在调用远程命令时，通信模块根据远程地址的第一个字段，查找子节点，查取成功后，会将剩余的地址与命令发送给查到的子节点。
子节点收到地址与命令后，继续查找子节点，直到目标节点收到命令，然后将执行结果原路返回。

在查找的过程中，如果没有查找到子节点，则会传给上级节点重新处理。

#### 节点认证

**节点加密**
每个节点都有证书与密钥。每个节点在发送命令时，都会用自己的密钥签名，目标节点都会用它的证书验签。以此保证命令来源的可靠性。

**节点类型**

- 初始节点，没有归属的节点
- 主控节点，有用户证书与密钥的节点
- 从属节点，有用户证书的节点

- 代理节点，主控节点指定的代理节点
- 共享节点，允许多个用户控制的节点
- 认证节点，专门用来存放与查询用户证书的节点

用户在某一设备上添加自己的证书与密钥，此节点即为主控节点。 在其它节点上绑定自己的证书，此即为从属节点。

主控节点就可任意控制从属节点，从属节点不能控制主控节点，从属节点之间也不能相互控制。

主控节点可以指定代理节点，代理节点可以代替主控节点控制从属节点。

共享节点，会有多个用户，可能产生冲突，所以需要认证节点协调。

访问共享节点前，需要向认证节点注册，共享节点会从认证节点取出访问用户的信息。

#### 节点权限

***角色***

每个访问用户，都会指定一个角色。
主控节点的用户默认有root权限。即拥有设备的所有控制权限。

认证节点的用户默认有tech权限。一般是部分功能。

其它节点的用户默认有void权限。即拥有最小集合的权限，一般是只读的命令。

***组件***

组件是功能的集合，远程访问至少要有remote与source组件的权限才可以执行命令。

每个角色下都会有多个组件的权限。

***命令***

命令就节点向外提供功能的最小单元。

每个组件下都会有多条命令。

***规则***

每条命令内部可以用组件与命令的权限机制自定义权限检查。

权限的分配完全由主控节点与从属节点自己配置，其它节点不许配置。

***示例***

如下配置，用户shy的角色是tech，角色tech下有两个组件remote与source，每个组件下都有命令pwd与dir，所以用户shy就可以远程调用命令dir与pwd
```
role tech user shy
role tech componet remote command pwd dir
role tech componet source command pwd dir
```

### 存储引擎
#### 配置
#### 缓存
#### 数据

