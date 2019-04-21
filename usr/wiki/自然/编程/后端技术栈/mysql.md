## 简介

MySQL 是一个开源的关系型数据库管理系统。

- 官网: <https://www.mysql.com/>
- 源码: <https://dev.mysql.com/downloads/file/?id=482483>
- 文档: <https://dev.mysql.com/doc/refman/5.6/en>
- 开源: <https://github.com/mysql/mysql-server>

## 基础命令

## 存储引擎

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

