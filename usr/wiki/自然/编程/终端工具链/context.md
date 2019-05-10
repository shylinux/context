## context

context is not only context

### 下载安装
#### 简易版
在Mac或Linux上，可以直接用脚本下载
```
$ mkdir context && cd context
$ curl https://shylinux.com/code/upgrade/boot_sh | bash -s install
```

#### 完整版
如果对源码有兴趣，使用更丰富的功能，可以直接下载源码，
```
$ git clone https://github.com/shylinux/context.git
```
下载完源码后，如果安装了golang，就可以对源码直接进行编译，
如果没有可以去官网下载安装golang（<https://golang.org/dl/>）。
第一次make时，会自动下载各种依赖库，所以会慢一些。
```
$ cd context && make
```

### 启动服务
bin目录下是各种可执行文件，如启动脚本boot.sh与node.sh。

node.sh用来启动单机版context，boot.sh用来启动网络版context。

如下，直接运行脚本，即可启动context。

启动context后，就可以解析执行各种命令，即可是本地的shell命令，也可以是内部模块命令。

```
$ bin/node.sh
0[03:58:43]ssh>
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

### 组建集群

context不仅只是一个shell，还可以用来组建集群。

#### 启动服务节点

启动服务节点，使用脚本boot.sh，
与node.sh不同的是，boot.sh启动的context，
会启动web模块监听9094端口，会启动ssh模块监听9090端口。
```
$ cd context
$ bin/boot.sh
0[11:23:03]ssh>
```

#### 启动工作节点
启动工作节点，使用脚本node.sh，
它启动的context，会主动连接本地9090端口，向服务节点注册自己。

如下，新打开一个终端，调用boot.sh，创建并启动服务节点demo。
```
$ cd context
$ bin/node.sh create app/demo
0[11:23:03]ssh>
```

如下，再打开一个终端，调用boot.sh，创建并启动服务节点led。
```
$ cd context
$ bin/node.sh create app/led
0[11:23:03]ssh>
```

#### 调用远程命令
如下回到服务节点终端，执行remote命令，可以查看到所有远程节点。
```
22[11:35:12]ssh> remote
key   create_time          module          name   type
com   2019-05-09 20:57:21  ctx.nfs.file4   com    master
led   2019-05-09 20:59:28  ctx.nfs.file5   led    worker
demo  2019-05-09 20:59:28  ctx.nfs.file5   demo   worker
```

远程命令只需要在命令前加上节点名与冒号即可。

如下，远程调用led节点的命令。
```
24[11:41:15]ssh> led:pwd
/Users/shaoying/context/app/led/var
```

如下，远程调用demo节点的命令。
```
24[11:41:15]ssh> demo:pwd
/Users/shaoying/context/app/demo/var
```

如下，远程调用所有子节点的命令。
```
29[11:44:09]ssh> *:pwd
/Users/shaoying/context/app/led/var

/Users/shaoying/context/app/demo/var
```

#### 启动分机服务

boot.sh不仅可以用来启动本地服务，还可以将不同的主机组建在一起。

在另外一台计算机上，重新下载安装一下context，然后启动服务节点。

其中环境变量ctx_dev，用来指定上级服务节点。

```
$ cd context
$ ctx_dev=http://192.168.0.106:9094 boot.sh
0[11:53:11]ssh>
```

回到原主机的服务节点终端，
使用remote命令，可以查看到新加的从机节点。
```
30[11:55:38]ssh> remote
key   create_time          module          name   type
com   2019-05-09 20:57:21  ctx.nfs.file4   com    master
mac   2019-05-10 10:53:00  ctx.nfs.file13  mac    server
led   2019-05-09 20:59:28  ctx.nfs.file5   led    worker
demo  2019-05-09 20:59:28  ctx.nfs.file5   demo   worker
```

当然，也可以在本地启动多个服务节点，根据ctx_dev指定不同的上级节点，可以级联，也可以并联。
```
$ cd context
$ ctx_dev=http://localhost:9094 boot.sh create app/sub
```
***注意，不指定ctx_dev时，默认连接 https://shylinux.com ，如果不信任此主机，记得设置ctx_dev***

### 网页服务

下载完整版的context，启动的服务节点，就会带有前端网页服务。

#### 工具链

code模块提供了工具链管理。

访问服务节点：http://localhost:9094/code

因为工具链，可以直接调用各种命令，为了安全，所以需要用户认证才能授权。

可以在服务节点的终端，直接添加用户认证信息。

如下，添加一个角色为root的，用户名为who，用户密码为ppp。
```
25[11:54:44]ssh> ~aaa role root user who ppp
```

也可以在启动文件中，加入这条配置，记得重启一下服务节点
```
$ cat etc/local.shy
~aaa
    role root user who who
```

然后，在登录网页上输入用户名与密码，即可登入网页工作台。

页面工作台，只是对命令进行了一层封装，但提供了更流畅的交互。

以fieldset的形式，将各种命令以弱耦合的形式，组织在一起。

每一个fielset就是一个命令，默认的工具链有

- buffer，tmux粘贴板管理
- upload，文件上传
- dir，文件列表管理
- pod，节点列表管理
- ctx，模块列表管理
- cmd，命令行

使用pod，选择任意远程节点， 然后使用ctx，选择任意模块，
然后在cmd执行的命令，都会发送给指定的节点与模块。

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

#### 信息流

chat模块提供了信息管理。

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

### 小程序
### 开发板

## 接口开发
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
#### 协程
#### 消息
### 解析引擎
#### 文件扫描
#### 词法解析
#### 语法解析
#### 执行命令
### 通信框架
#### 节点路由
#### 节点认证
#### 节点服务
### 存储引擎
#### 配置
#### 缓存
#### 数据

