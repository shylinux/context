## 简介

MySQL 是一个开源的关系型数据库管理系统。

- 官网: <https://www.mysql.com/>
- 源码: <https://dev.mysql.com/downloads/file/?id=482483>
- 文档: <https://dev.mysql.com/doc/refman/5.6/en>
- 开源: <https://github.com/mysql/mysql-server>

## 下载安装
### Ubuntu安装MySQL
安装服务器与客户端
```
sudo apt-get install mysql-server mysql-client
```

相关目录与文件

- 配置目录 /etc/mysql/
- 运行状态 /var/run/mysqld/
- 日志目录 /var/log/mysql/
- 数据目录 /var/lib/mysql/
- 其它文件 /var/lib/mysql-files/
- 动态插件 /usr/lib/mysql/plugin/

## 基础命令
初次登录，直接用安装时设置的密码连接
```
$ mysql -u root -p
Enter password
mysql> 
```
查看帮助信息
```
mysql> help
...
```
其中常用的命令有

- status 查看当前状态
- system 调用系统shell命令
- source 加载并执行sql文件
- delimiter 设置行分隔符
- connect 重新连接服务器
- use 切换数据库

### 数据库操作

- 查看 show databases
- 切换 use demo
- 创建 create database demo
- 删除 drop database demo

查看数据库列表
```
mysql> show databases;
```
创建数据库
```
mysql> create database demo;
```
切换数据库
```
mysql> use demo;
```
删除数据库
```
mysql> drop database demo;
```

### 关系表操作

- 创建 create table demo(a int)
- 修改 alter table add column b int
- 删除 drop table demo

### 数据操作

- 查询 select * from demo
- 添加 insert into demo values(1,2)
- 修改 update demo set a=1, b=2 where a=3
- 删除 delete from demo where a=1

## 存储引擎
### InnoDB

### MyISAM

- 下载: <https://dev.mysql.com/downloads/mysql/5.6.html#downloads>

- 博客: <http://blog.codinglabs.org/articles/theory-of-mysql-index.html>

变量的定义与引用: <https://www.cnblogs.com/EasonJim/p/7966918.html>

show engines
show engine innodb status
show variables

show databases
create database demo
drop database demo
use demo

show tables
create table t(a int unsigned not null, b char(10), primary key(a))
drop table t

select * from t;
insert into t values()
update t set b='1234'
delete from t

