## 简介

MySQL 是一个开源的关系型数据库管理系统。

- 官网: <https://www.mysql.com/>
- 源码: <https://dev.mysql.com/downloads/file/?id=480541>
- 文档: <https://dev.mysql.com/doc/refman/5.5/en>
- 开源: <https://github.com/mysql/mysql-server/tree/5.5>

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

### 配置操作
mysql启动时会从/etc/mysql/my.cnf加载配置，show variables;可以查看当前配置

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

- 查看 show tables
- 创建 create table demo(a int)
- 修改 alter table add column b int
- 查看 desc demo
- 删除 drop table demo

### 数据操作

- 查询 select * from demo
- 添加 insert into demo values(1,2)
- 修改 update demo set a=1, b=2 where a=3
- 删除 delete from demo where a=1

## 存储引擎

- 源码：<https://github.com/mysql/mysql-server/tree/5.5/storage>
- 文档：<https://dev.mysql.com/doc/refman/5.5/en/storage-engines.html>

```
mysql> show engines;
```

|engine|comment|
|------|-------|
|InnoDB| supports transactions, row-level locking, and foreign keys|
|MyISAM| |
|MEMORY|hash based, stored in memory, useful for temporary tables|
|BLACKHOLE| |
|ARCHIVE| |
|CSV| |
|MRG_MYISAM| |
|PERFORMANCE_SCHEMA| |
|FEDERATED| |

### InnoDB

InnoDB是一种支持ACID事务的存储引擎

- 源码：<https://github.com/mysql/mysql-server/tree/5.5/storage/innobase>
- 文档：<https://dev.mysql.com/doc/refman/5.5/en/innodb-storage-engine.html>

InnoDB是多线程的，通过命令可以看各线程的状态
```
mysql> show engine innodb status\G
```
|meta|task|
|------|----|
|BACKGROUND| 主线程 |
|FILE I/O| IO线程 |
|INSERT BUFFER AND HASH INDEX| 插入缓存与哈希索引 |
|SEMAPHORES| |
|TRANSACTIONS| |
|BUFFER POOL AND MEMORY| |
|ROW OPERATIONS| |
|LOG| |

#### 线程与缓存
MySQL的存储引擎是模块化，所以需要注册模块信息
```
// 注册模块 struct st_mysql_plugin {} // mysql-5.5.62/include/mysql/plugin.h:423
mysql_declare_plugin(innobase) { // mysql-5.5.62/storage/innobase/handler/ha_innodb.cc:11925
    MYSQL_STORAGE_ENGINE_PLUGIN,
    &innobase_storage_engine,
    innobase_hton_name,
    plugin_author,
    "Supports transactions, row-level locking, and foreign keys",
    PLUGIN_LICENSE_GPL,
    innobase_init, /* Plugin Init */
    NULL, /* Plugin Deinit */
    INNODB_VERSION_SHORT,
    innodb_status_variables_export,/* status variables             */
    innobase_system_variables, /* system variables */
    NULL, /* reserved */
    0,    /* flags */
}
...
mysql_declare_plugin_end;
```

InnoDB在初始化函数中，会注册各种回调函数，并启动各种工作线程
```
innobase_init() // mysql-5.5.62/storage/innobase/handler/ha_innodb.cc:2218
    // 注册各种回调函数 struct handlerton {} // mysql-5.5.62/sql/handler.h:705
    innobase_hton->state = SHOW_OPTION_YES;
    innobase_hton->db_type= DB_TYPE_INNODB;
    innobase_hton->savepoint_offset=sizeof(trx_named_savept_t);
    innobase_hton->close_connection=innobase_close_connection;
    innobase_hton->savepoint_set=innobase_savepoint;
    innobase_hton->savepoint_rollback=innobase_rollback_to_savepoint;
    innobase_hton->savepoint_release=innobase_release_savepoint;
    innobase_hton->commit=innobase_commit;
    innobase_hton->rollback=innobase_rollback;
    innobase_hton->prepare=innobase_xa_prepare;
    innobase_hton->recover=innobase_xa_recover;
    innobase_hton->commit_by_xid=innobase_commit_by_xid;
    innobase_hton->rollback_by_xid=innobase_rollback_by_xid;
    innobase_hton->create_cursor_read_view=innobase_create_cursor_view;
    innobase_hton->set_cursor_read_view=innobase_set_cursor_view;
    innobase_hton->close_cursor_read_view=innobase_close_cursor_view;
    innobase_hton->create=innobase_create_handler;
    innobase_hton->drop_database=innobase_drop_database;
    innobase_hton->panic=innobase_end;
    innobase_hton->start_consistent_snapshot=innobase_start_trx_and_assign_read_view;
    innobase_hton->flush_logs=innobase_flush_logs;
    innobase_hton->show_status=innobase_show_status;
    innobase_hton->flags=HTON_SUPPORTS_FOREIGN_KEYS;
    innobase_hton->release_temporary_latches=innobase_release_temporary_latches;

    // 启动各种工作线程
    innobase_start_or_create_for_mysql() // srv/srv0start.c:1028
        os_thread_create(io_handler_thread, NULL, NULL);
        os_thread_create(&srv_lock_time_thread, NULL, NULL);
        os_thread_create(&srv_error_monitor_thread, NULL, NULL);
        os_thread_create(&srv_monitor_thread, NULL, NULL);
        os_thread_create(&srv_master_thread, NULL, NULL);
        os_thread_create(&srv_purge_thread, NULL, NULL);
```

主线程
```
srv_master_thread() // mysql-5.5.62/storage/innobase/srv/srv0srv.c:2748
loop:
    srv_main_1_second_loops++;

    // 同步日志缓存
    srv_sync_log_buffer_in_background(); // srv/srv0srv.c:2702
        log_buffer_sync_in_background(TRUE); // log/log0log.c:1689
            log_write_up_to(lsn, LOG_NO_WAIT, flush);
                mutex_enter(&(log_sys->mutex));
                log_sys->write_lsn = log_sys->lsn;
                fil_flush(group->space_id); // fil/fil0fil.c:4755
                    mutex_enter(&fil_system->mutex);
                    os_file_flush(file); // os/os0file.c:2119
                        os_file_fsync(file);
                            fsync(file);

    // 合并插入缓存
    ibuf_contract_for_n_pages(FALSE, PCT_IO(5));

    // 刷新数据缓存
    buf_flush_list(PCT_IO(100), IB_ULONGLONG_MAX); // buf/buf0flu.c:1928
        buf_flush_batch(buf_pool, BUF_FLUSH_LRU, min_n, 0);
            buf_pool_mutex_enter(buf_pool);
            buf_flush_LRU_list_batch(buf_pool, min_n);
            buf_flush_flush_list_batch(buf_pool, min_n, lsn_limit);
            buf_flush_buffered_writes();
                mutex_enter(&(trx_doublewrite->mutex));
                buf_flush_sync_datafiles();

    // 执行清理任务
    srv_master_do_purge();
        trx_purge(srv_purge_batch_size); // trx/trx0purge.c:1139
            rw_lock_x_lock(&purge_sys->latch);
            thr = que_fork_start_command(purge_sys->query);
            que_run_threads(thr);

background_loop:
    srv_main_background_loops++;
	row_drop_tables_for_mysql_in_background(); // row/row0mysql.c:2253
        mem_free(drop->table_name);

flush_loop:
    srv_main_flush_loops++;
	log_checkpoint(TRUE, FALSE); // log/log0log.c:2077
        log_groups_write_checkpoint_info();
            log_group_checkpoint(group);

suspend_thread:
	srv_suspend_thread(slot);

```

- 插入缓存与两次写
- 自适应哈希索引
- 启动、恢复、关闭
- 引擎插件升级

#### 各种文件

配置文件，MySQL在启动时会加载配置文件，这些配置在运行时也可以随时读写

- show variables;
- show variables like 'timeout';
- set [global|session] var=val

日志文件
数据文件

### MyISAM

- 源码：<https://github.com/mysql/mysql-server/tree/5.5/storage/myisam>
- 文档：<https://dev.mysql.com/doc/refman/5.5/en/myisam-storage-engine.html>
