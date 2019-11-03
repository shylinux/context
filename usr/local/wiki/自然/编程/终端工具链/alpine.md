## alpine

- 官网: https://www.alpinelinux.org/
- 文档: https://wiki.alpinelinux.org/
- 源码: https://github.com/alpinelinux/docker-alpine
- 博客: https://blog.csdn.net/zl1zl2zl3/article/details/80118001

## 安装

```
docker run alpine -it pwd
```

## 安装
### 主机名
```
$ echo myos > /etc/hostname
$ hostname -F /etc/hostname
$ sed -i -r 's#127.0.0.1.*#127.0.0.1 myos#g' /etc/hosts
$ /etc/resolv.conf
```

### 网络配置
```
$ ip
$ ping
$ udhcpc
$ ifconfig
$ /etc/network/interfaces
auto lo
iface lo inet lookback
auto eth0
iface eth0 inet dhcp
iface eth1 inet static
iface eth1 inet static
    address 192.168.1.21
    netmask 255.255.0.0
    gateway 192.168.1.1

$ wpa_supplicant
$ /etc/wpa_supplicant/wpa_supplicant.conf
```

### 包工具
```
/etc/apk/repositories
/var/cache/apk
/etc/apk/world
$ apk
$ apk update
$ apk search
$ apk cache
$ apk info
$ apk add
$ apk del
```

### 开发环境

```
$ apk update
$ apk add build-base
```

```
/* vi hi.c */
#include<stdio.h>

int main(int argc, char *argv[]) {
    println("hello c world!");
}
```

```
$ gcc hi.c -o hi
$ ./hi
hello c world!
```

