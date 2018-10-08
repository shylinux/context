## 《后端技术栈》简介
后端技术栈，是对后端服务器开发中所用到的技术进行组合，不断的优化框架结构与服务水平。

- Linux
- Nginx
- Python 是一种脚本语言，
- MySQL
- Redis

## 基础入门
### Linux
Linux系统的应用十分广泛，有很多流行的桌面操作系统与移动操作系统，尤其是在云计算的服务器领域和物联网的嵌入式设备中，
因为有大量免费开源的软件可以自由的下载与使用，尤其是有很多优秀的软件开发工具，可以搭建起高效的开发与测试环境，所以熟练掌握Linux是程序员必备的技能。

严格上意义上讲Linux只是一个系统内核，没法直接使用，还需要各种应用软件。
GNU是一个软件基金组织，提供了各种各样免费开源的自由软件。
Debian就是把Linux内核与GNU软件打包成一个完整的操作系统，并且提供了软件包管理工具，可以很方便的下载各种应用软件。
Ubuntu在Debian的基础上，提供了更加友好的桌面系统，降低了使用难度。
Apache也是一个著名的软件基金组织，赞助了一系列的开源软件，尤其是服务器相关的软件，所以推动了WEB繁荣发展。
Mozilla是一个软件社区，开发了浏览器Firefox，提供了一系列的Web技术与文档，推动了WEB繁荣发展。

- [Linux官网](https://www.linux.org/)
- [GNU官网](https://www.gnu.org/)
- [Debian官网](https://www.debian.org/)
- [Ubuntu官网](https://www.ubuntu.com/)
- [Apache官网](https://www.apache.org/)
- [Mozilla官网](https://developer.mozilla.org/)

电脑端常见的操作系统是MacOSX与Windows。
所以使用Linux常用的方式有：

- 购买一个开发板，如树莓派开发板，提供了一整套完整的环境，可以制作各种电子设备，软硬件结合，更加有直观的成就感。
- 租用一个云主机，如阿里云的服务器也很便宜，再申请一个域名与证书，就可以搭建一服务器，提供各种应用服务，也会很有成就感。
- 本地安装虚拟机，当然最主要的还是在自己电脑上安装一下，可以选择各种虚拟机软件安装如VMWare或VirtualBox，像应用软件一样安装操作系统。最近流行的Docker容器技术，也可以尝试一下。

#### 软件管理
Linux之所以这么流行，就是因为有大量优秀的开源软件与免费软件可以自由的获取与使用。
apt是Debian的软件管理工具，只需要一条命令即可下载所需的软件。
软件包，是开发人员将程序、文档、脚本、配置等相关打包在一起，方便软件的分发与部署。
软件源，是专门存放软件包的服务器，用户可以在线搜索、查看、下载各种软件包。
软件包之间会存在依赖关系，使用apt下载软件包时，会检测软件包的依赖关系，自动下载并安装相关的软件包。


最常用的两条命令是update与install，更新软件信息列表和安装软件包。
更新软件包信息列表，从软件源服务器上下载软件包的信息列表。安装软件时，会根据这些信息下载相关软件包。
需要经常更新一下。
```
$ sudo apt-get update
```
安装软件包，只需输入软件包名就可以自动下载并安装相关软件。如下安装vim。
```
$ sudo apt-get install vim
```
当不知道软件包的完整名字时，可以使用search命令，如下搜索docker。
```
$ apt-cache search docker
```

- 软件源 /etc/apt/sources.list
- 软件包清单 /var/lib/apt/lists
- 软件包缓存 /var/cache/apt/archives/
- 更新软件源 apt-get update
- 安装软件包 apt-get install
- 卸载软件包 apt-get remove
- 清空软件包 apt-get purge
- 清理缓存包 apt-get clean
- 清理无用包 apt-get autoclean
- 搜索软件包 apt-cache search
- 查看软件包 apt-cache show

#### 帮助信息
Linux有大量的软件与工具，每个软件都有自己的使用方法与参数。
一下子记住这么多信息是不可能的，所以需要快速找到相关的帮助信息。

最直接的帮助信息是软件自带，输入参数-h或是--help，就可以查看参数列表及相关信息。如下查看命令ps的帮助信息。
```
$ ps --help
```
另外，查看更详细的使用信息使用man命令。man手册，使用交互式查看文档，可以翻页，可以搜索。
```
$ man ps
```
使用whatis可以查看命令的简要描述信息。
```
$ whatis ps
```
查看命令所在的文件。
```
$ which ps
```
当不知道命令的完整名字时，可以使用模糊搜索。
```
$ apropos nice
```
很多软件都有自己的官网，如果还需要更多的信息，可以去官网查阅在线的文档。

#### 远程登录
很多时候运行环境并不在本机，需要去远程登录设备，如访问服务器，如连接开发板。
这时就会用到ssh工具，ssh是security shell的简写，提供加密通信的远程连接。可以远程执行各种命令，传输文件等。

ssh参数指定所需要用户名与主机地址，即可。
```
$ ssh shy@10.0.0.10
```

如果使用密码认证，每次都需要输入密码，很是麻烦，尤其是开发环境也在远程设备上，需要频繁的输入密码。
可以使用ssh-keygen命令，生成密钥对，使用密钥文件登录。
```
$ ssh-keygen
```

密钥生成后，用ssh-copy-id命令将公钥上传到远程主机。下次再用ssh登录时就不用再输入密码了。
```
$ ssh-copy-id shy@10.0.0.10
```

有时候需要上传或是下载文件，scp命令像cp命令一样简单，可以直接在本机与远程主机传输文件。
与cp不同的是，远程文件名前需要加上用户名与主机地址。

将远程主机的文件下载到本地。
```
$ scp shy@10.0.0.10:/home/shy/.vimrc vimrc
```
将本地文件上传到远程主机。
```
$ scp vimrc shy@10.0.0.10:/home/shy/.vimrc
```
当需传输更多文件时，可以使用sftp命令，交互式的访问远程目录。使用get下载文件，put上传文件。
```
$ sftp shy@10.0.0.10
```

- 私钥文件 ~/.ssh/id_rsa
- 公钥文件 ~/.ssh/id_rsa.pub
- 授权公钥 ~/.ssh/authorized_keys

#### 编译代码
计算机由硬件与软件组成，软件又分为操作系统与应用程序。

无论是Linux，还是MacOSX，还是Windows，操作系统给应用程序提供了API接口。
应用程序通过调用这些API接口，实现各种各样的功能。

如果现有的工具不能满足需求时，就需要下载工具的源码进行定制编译，或是开发一些模块。

API



##### Makefile
make命令根据Makefile文件中定义的规则，调用各种命令，来生成目标文件。
可以用来调用编译器，将项目的源码文件，编译成可执行文件。
将源码文件编译成可执行文件。
[Makefile官方文档](https://www.gnu.org/software/make/manual/make.html)
<>

#### 本地化

- locale

#### 网络管理

- hostname
- /etc/hosts
- /etc/resolv.conf

#### 用户管理

- /etc/passwd
- /etc/shadow
- /etc/group
- /etc/shells
- vipw
- vigr
- passwd
- chsh
- newgrp
- id
- addgroup
- delgroup
- groupmod
- adduser
- /etc/adduser.conf
- /etc/skel
- bash
- /etc/bash.bashrc
- /etc/profile
- ~/.bashrc
- ~/.bash_profile
- zsh
- /etc/zshrc
- /etc/zshenv

#### 时间管理

- date
- /etc/timezone

#### 权限管理

- sudo
- visudo
- /etc/sudoers

#### 磁盘管理

- /etc/fstab
- mount
- umount

#### 内核管理
#### 系统启动
system

### Nginx
nginx是一种Web服务器，尤其是反向代理与负载均衡功能在高并发的服务器上应用广泛。更多信息参考：[nginx官网](http://nginx.org/)

安装nginx
```
$ sudo apt-get install nginx
```
启动nginx
```
$ sudo nginx
```
访问nginx
```
$ curl localhost
```

nginx通过丰富的配置文件，启用各种功能。查看nginx相关的文件。
```
$ nginx -V
```

### Python
Mac上自带python，不需要安装。Ubuntu上也自带python。更多信息参考：[python官网](https://www.python.org/)

### MySQL
### Redis
## 微服务化
### es
### etcd
### kafka
### consul
### thrift
### databus
## 源码解析
### Linux
### Nginx
### Python
### MySQL
### Redis
