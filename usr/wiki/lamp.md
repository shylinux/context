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
debian就是把Linux内核与GNU软件打包成一个完整的操作系统，并且提供了软件包管理工具，可以很方便的下载各种应用软件。
ubuntu在debian的基础上，提供了更加友好的桌面系统，降低了使用难度。
apache也是一个著名的软件基金组织，赞助了一系列的开源软件，尤其是服务器相关的软件，所以推动了WEB繁荣发展。

- [Linux官网](https://www.linux.org/)
- [GNU官网](https://www.gnu.org/)
- [debian官网](https://www.debian.org/)
- [Ubuntu官网](https://www.ubuntu.com/)
- [Apache官网](https://www.apache.org/)

电脑端常见的操作系统是MacOSX与Windows。
所以使用Linux常用的方式有：

- 购买一个开发板，如树莓派开发板，提供了一整套完整的环境，可以制作各种电子设备，软硬件结合，更加有直观的成就感。
- 租用一个云主机，如阿里云的服务器也很便宜，再申请一个域名与证书，就可以搭建一服务器，提供各种应用服务，也会很有成就感。
- 本地安装虚拟机，当然最主要的还是在自己电脑上安装一下，可以选择各种虚拟机软件安装如VMWare或VirtualBox，像应用软件一样安装操作系统。最近流行的Docker容器技术，也可以尝试一下。

#### 软件管理
apt是debian与ubuntu的软件管理工具，Linux之所以这么流行，就是因为有大量优秀的开源软件与免费软件可以自由的获取与使用。

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

- man
- whatis
- apropos

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
#### 远程登录

- ssh
- scp
- sftp
- ssh-keygen -t rsa
- ~/.ssh/id_rsa
- ~/.ssh/id_rsa.pub
- ssh-copy-id
- ~/.ssh/authorized_keys



### Nginx
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
