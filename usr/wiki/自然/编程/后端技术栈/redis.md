## 简介

Redis是一个使用ANSI C编写的开源、支持网络、基于内存、可持久性的键值对存储数据库。
Redis是最流行的键值对存储数据库。

- 官网: <https://redis.io>
- 源码: <http://download.redis.io/releases/redis-4.0.9.tar.gz>
- 文档: <https://redis.io/documentation>

## 源码安装
```
$ wget http://download.redis.io/releases/redis-4.0.11.tar.gz
$ tar xzf redis-4.0.11.tar.gz
$ cd redis-4.0.11
$ make
```
#### 启动服务端
```
$ src/redis-servce
...
```
#### 启动客户端
```
$ src/redis-cli
127.0.0.1:6379>
```
#### 基本命令
```
$ src/redis-cli
127.0.0.1:6379> set employee_name shy
OK
127.0.0.1:6379> get employee_name
"shy"
```
## 源码解析
### 目录解析

- COPYING 版权文件
- README.md 说明文档
- Makefile make文件
- deps/ 依赖库
- src/ 源码目录
- tests/ 测试脚本
- utils/ 工具脚本
- redis.conf 配置文件
- sentinel.conf 配置文件

BUGS
INSTALL
MANIFESTO
CONTRIBUTING
00-RELEASENOTES

dump.rdb
runtest
runtest-cluster
runtest-sentinel

### 代码解析
```
server.h //服务端
  redisObject:struct //数据结构
    type:unsigned
    encoding:unsigned
    lru:unsigned
    refcount:int
    ptr:void*

server.c //服务端
  server: redisServer //服务端上下文
    pid: pid_t
    configfile: char*
    executable: char*
    exec_argv: char*
    commands: dict* //命令哈希表
    db: redisDb* //数据库
      dict: dict*
      expires: dict*
      blocking_keys: dict*
      ready_keys: dict*
      watched_keys: dict*
      id: int
      avg_ttl: long long
    clients: list/client //客户端连接
      id: uint64
      fd: int
      db: redisDb*
      name: robj*
      querybuf: sds
      pending_querybuf: sds
      argc: int
      argv: robj**
      cmd: redisCommand*
      reply: list*

  redisCommandTable: redisCommand //命令列表
    "get": getCommand
    "set": setCommand
      setGenericCommand(c)
        setKey(c->db,k,v)
          lookupKeyWrite(db,k)
            lookupKey(db,k)
              dictFind(db->dict,k->ptr)
          dbAdd(db,k,v)
            dictAdd(db->dict,k->ptr,v)
          dbOverwrite(db,k,v)
            dictReplace(db->dict,k->ptr,v)
        addReply(c, o)
          prepareClientToWrite(c)
            listAddNodeHead(server.clients_pending_write, c)
          _addReplyToBuffer(c, o)
            c->buf
            c->bufpos
          _addReplyObjecToList(c, o)
            listAddNodeTail(c->reply, sdsdup(o->ptr))

  serverLog() //输出日志
    server.verbosity
    serverLogRaw()
      server.logfile
  ustime()
  mstime()


  main() //程序入口
    initServerConfig() //初始化server
      populateCommandTable() //加载命令列表
        server.commands = redisCommandTable
    loadServerConfig() //加载配置文件
    initServer() //
      aeCreateFileEvent()
    loadDataFromDisk()
    aeMain():ae.c //事件循环
      el->beforesleep()
        handleClientsWithPendingWrites() //返回命令执行结果
          writeToClient(c)
            write(c->buf)
      aeProcessEvents(el):ae.c
        aeApiPoll()
        el->aftersleep()
        fe->rfileProc()/acceptTcpHandler() //添加网络监听事件
          anetTcpAccept()
          acceptCommonHandler()
            createClient()
              aeCreateFileEvent()/readQueryFromClient(el) //添加读取数据事件
                read(c->querybuf)
                processInputBuffer(c)
                  processInlineBuffer() //解析客户端命令
                    c->argv[i]=createObject()
                  processCommand(c) //执行客户端命令
                    c->cmd=lookupCommand()
                      dictFetchValue(server.commands)
                    call(c)
                      c->cmd->proc(c)/setCommand(c)

      fe->wfileProc()
      fe->rfileProc()
      processTimeEvnts()

db.c
  setKey()
t_string.c
  setGenericCommand()
  setCommand()

t_hash.c

t_list.c
t_set.c
t_zset.c

networking.c //
  createClient()

adlist.h //双链表
aslist.c
ae.h //事件循环
ae.c
ae_epoll.c
ae_evport.c
ae_kqueue.c
ae_select.c
anet.h //网络接口
anet.c
aof.c
asciilogo.h
atomicvar.h
bio.h
bio.c
bitops.c
blocked.c
childinfo.c
cluster.h
cluster.c
config.h
config.c
crc16.c
crc64.c
crc64.h
debug.c
debugmacro.h
defrag.c
dict.h
dict.c
edianconv.c
edianconv.h
evict.c
expire.c
fmacros.h
geo.c
geo.h
geohash.h
geohash.c
geohash_helper.h
geohash_helper.c
help.h
hyperloglog.c
intset.h
intset.c
latency.h
latency.c
lazyfree.c
lzf.h
lzf_c.h
lzf_d.h
lzfP.h
memtest.c
module.c
multi.c
networking.c
notify.c
object.c
pqsort.c
pqsort.h
pubsub.c
quicklist.c
quicklist.h
rand.c
rand.h
rax.c
rax.h
rax.malloc.h
rdb.c
rdb.h
redis-benchmark.c
redis-cli.c
redisassert.h
redismodule.h
release.c
release.h
replication.c
rio.h
rio.c
scripting.c
sds.h
sds.c
sdsalloc.h
sentinel.c
setproctitle.c
sha1.h
sha1.c
siphash.h
siphash.c
sort.c
sparkline.h
sparkline.c
syncio.c
testhelp.c
util.c
util.h
version.h
ziplist.h
ziplist.c
zipmap.c
zipmap.h
zmalloc.c
zmalloc.h





dict.c //
  dict:struct
    type:dictType
	privdata:void*
    ht:dictht[2]
      table:dictEntry**
        key:void*
        v:union
        next:void*
      size:long
      sizemask:long
      used:long
    rehashidx:long
    iterators:long
  
zmalloc.c //内存管理
  zmalloc()
  zcalloc()
  zrealloc()
  zmalloc_size()
  zfree()
  zstrdup()
```

### server.h
#### struct redisServer
#### struct client
#### struct redisObject
### networking.c
### db.c
### object.c
### t_hash.c
### t_list.c
### t_set.c
### t_string.c
### t_zset.c
### ae.c 事件循环

