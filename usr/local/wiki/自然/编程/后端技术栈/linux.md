## GNU

- 官网：<https://www.gnu.org/>
- grub：<https://ftp.gnu.org/gnu/grub/grub-2.00.tar.gz>
- bash：<https://ftp.gnu.org/gnu/bash/bash-4.3.30.tar.gz>
- libc: <http://ftp.gnu.org/gnu/libc/glibc-2.19.tar.gz>
- gcc：<https://ftp.gnu.org/gnu/gcc/gcc-4.8.4/gcc-4.8.4.tar.gz>
- gdb：<https://ftp.gnu.org/gnu/gdb/gdb-7.7.1.tar.gz>

C99
```
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <ctype.h>
#include <math.h>
#include <time.h>

#include <errno.h>
#include <setjmp.h>
#include <assert.h>
#include <signal.h>

#include <complex.h>
#include <fenv.h>
#include <float.h>
#include <inttypes.h>
#include <iso646.h>
#include <limits.h>
#include <stdarg.h>
#include <locale.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <tgmath.h>

#include <wchar.h>
#include <wctype.h>
```
POSIX.1
```
#include <sys/types.h>
#include <unistd.h>

#include <poll.h>
#include <fcntl.h>
#include <termios.h>
#include <sys/select.h>
#include <dlfcn.h>

#include <pwd.h>
#include <grp.h>

#include <aio.h>
#include <tar.h>
#include <cpio.h>

#include <glob.h>
#include <dirent.h>
#include <fnmatch.h>
#include <sys/stat.h>
#include <sys/statvfs.h>

#include <sched.h>
#include <pthread.h>
#include <semaphore.h>
#include <sys/wait.h>

#include <netdb.h>
#include <arpa/inet.h>
#include <net/if.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <sys/socket.h>
#include <sys/un.h>

#include <iconv.h>
#include <langinfo.h>
#include <monetary>
#include <nl_types.h>
#include <regex.h>
#include <strings.h>
#include <wordexp.h>
#include <sys/mman.h>
#include <sys/times.h>
#include <sys/utsname.h>

#include <fmtmsg.h>
#include <ftw.h>
#include <libgen.h>
#include <ndbm.h>
#include <search.h>
#include <syslog.h>
#include <utmpx.h>
#include <sys/ipc.h>
#include <sys/msg.h>
#include <sys/resource.h>
#include <sys/sem.h>
#include <sys/shm.h>
#include <sys/time.h>
#include <sys/uio.h>

#include <mqueue.h>
#include <spawn.h>
```
## LINUX

- 官网：<https://www.linux.org/>
- 文档：<https://www.kernel.org/doc/html/latest/>
- 源码：<https://mirrors.edge.kernel.org/pub/linux/kernel/v3.x/linux-3.7.4.tar.gz>
- 开源：<https://github.com/torvalds/linux>

- 代理：<https://mirror.tuna.tsinghua.edu.cn/kernel/v3.x/linux-3.7.4.tar.gz>

## 系统内核
### 内核源码

lib
arch
init
kernel
include
drivers
ipc
net
mm
fs

firmware
security
crypto
sound
block
virt

tools
scripts
samples
Makefile
Kconfig
Kbuild
usr

README
COPYING
CREDITS
Documentation
MAINTAINERS
REPORTING-BUGS

### 系统启动
```
startup_32() // linux-2.6.12/arch/i386/kernel/head.S:57
start_kernel() // linux-2.6.12/init/main.c:424
    do_basic_setup()
        sock_init() // net/socket.c:2025
	sched_init()
    trap_init() // arch/i386/kernel/traps.c
	init_IRQ()
	init_timers()
	softirq_init()
    time_init() // arch/i386/kernel/time.c
        time_init_hook()
            setup_irq(0, timer_interrupt)
```
### 进程调度
```
struct task_struct { // linux-3.7.4/include/linux/sched.h:1190
    // 进程状态
    state long / TASK_RUNNING // include/linux/sched.h:108

    // 线程信息
    thread_info *struct thread_info
        task *struct task_struct
        status

    // 进程树
    parent *struct task_struct
    children struct list_head
    sibling struct list_head

    mm *struct mm_struct
        start_code
        end_code
        start_data
        end_data
        start_brk
        brk
        mmap_base
        start_stack
        arg_start
        arg_end
        env_start
        env_end

    thread struct thread_struct
        sp0
        sp
        es
        ds
        fsindex
        gsindex
        ip
        fs
        gs

    fs *struct fs_struct
    files *struct files_struct
    pid pid_t
    tgid pit_t
}

schedule()
```

### 时间管理
```
timer_interrupt() // arch/i386/kernel/time.c:292
do_timer_interrupt() // arch/i386/kernel/time.c:250
    do_timer_interrupt_hook()
        do_timer() // kernel/timer.c:925
```
### 系统中断
```
interrupt() // arch/i386/kernel/entry.S:397
do_IRQ() // arch/i386/kernel/irq.c:51
    __do_IRQ() // kernel/irq/handle.c:107
        handle_IRQ_event()

do_softirq() // arch/i386/kernel/softirq.c:158
    __do_softirq() // kernel/softirq.c:74

tasklet_schedule() // include/linux/interrupt.h
    __tasklet_schedule() // kernel/softirq.c:223

run_workqueue() // kernel/workqueue.c:145
```
### 并发同步
```
atomic_add() // include/asm-i386/atomic.h
spin_lock_irq() // include/linux/spinlock.h
down_interruptible() // include/asm-i386/semaphore.h
wait_for_completion() // kernel/sched.c
```
### 系统调用
```
system_call() // arch/i386/kernel/entry.S:225
sys_call_table[] // arch/i386/kernel/syscall_table.S
	.long sys_restart_syscall	/* 0 - old "setup()" system call, used for restarting */
	.long sys_exit() // kernel/exit.c:859
        do_exit()
        exit_mm()
        exit_sem()
        __exit_files()
        __exit_fs()
        exit_namespace()
        exit_thread()
        cpuset_exit()
        exit_keys()
        exit_notify()
        schedule()
	.long sys_fork // arch/i386/kernel/process.c:645
        do_fork() // kernel/fork.c:1192
            copy_process()
                dup_task_struct()
                copy_files()
                copy_fs()
                copy_sighand()
                copy_signal()
                copy_mm()
                sched_fork()
            wake_up_new_task()
	.long sys_read
	.long sys_write
	.long sys_open		/* 5 */
        get_unused_fd()
        filp_open()
            open_namei()
                path_lookup()
            dentry_open()
        fd_install()

	.long sys_close
    .long sys_socketcall // net/socket.c:1897
        sys_socket()
            socket_create()
                __sock_create()
                    net_families[i]->create()/inet_create()
                        inetsw_array[j]->prot->init()/tcp_v2_init_sock()
        sys_bind()
        sys_connect()
        sys_listen()
        sys_accept()
        sys_getsocketname()
        sys_getpeername()
        sys_socketpair()
        sys_send()
            sys_sendto()
        sys_sendto()
            socket_sendmsg()
                __sock_sendmsg()
        sys_recv()
            sys_recvfrom()
        sys_recvfrom()
            sock_recvmsg()
                __sock_recvmsg()
                sock->ops->recvmsg()/inetsw_array[j]->ops->recvmsg()/inet_stream_ops->recvmsg()/sock_common_recvmsg()
                    sk->sk_prot->recvmsg()/inetsw_array[j]->prot->recvmsg()/tcp_prot->recvmsg()/tcp_recvmsg()
                        sk_wait_data()
                            prepare_to_wait(sk->sk_sleep)
                            sk_wait_event()
                                release_sock()
                                    __release_sock()
                                        sk->sk_backlog_rcv()/sk->sk_prot->backlog_rcv()/inetsw_array[j]->prot->backlog_rcv()/tcp_prot->backlog_rcv()/tcp_v4_do_rcv()
        sys_shutdown()
        sys_setsocketopt()
        sys_getsocketopt()
        sys_sendmsg()
        sys_recvmsg()
        
...
```

## 网络编程
### 网络协议
MAC ARP
IP ICMP
TCP UDP SCTP
RIP BGP OSPF
DNS NAT DHCP
FTP SSH SNMP
SMTP IMAP POP3
HTTP HTML

### 网络命令
```
ifconfig
nslookup
netstat
tcpdump
telnet
ping
wget
curl

/etc/resolve.conf
/etc/hostname
/etc/hosts
```

### 网络编程
```
#include <sys/socket.h>
socket()
bind()
listen()
accept()
connect()
recv()
send()
shutdown()

#include <netinet/in.h>
socket_in: struct
    sin_len: uint8_t
    sin_family: sa_family_t
    sin_port: in_port_t
    sin_addr: struct in_addr
        s_addr: in_addr_t
htons()
ntohs()

#include <arpa/inet.h>
inet_addr()
inet_ntoa()
inet_aton()

#include <netdb.h>
gethostbyname()
gethostbyaddr()
getservbyname()
getservbyport()
```

### 系统调用
```
socket()/sys_socket() // net/socket.c:1180
	sock_create(family, type, protocol, &sock)
        __sock_create(family, type, protocol, res, 0)
            net_families[PF_INET]->create(sock, protocol)/inet_create()
                sock_init_data(sock, sk)
                    skb_queue_head_init(&sk->sk_receive_queue)
                    skb_queue_head_init(&sk->sk_write_queue)
                    skb_queue_head_init(&sk->sk_error_queue)
                    sk->sk_state_change = sock_def_wakeup
                    sk->sk_data_ready = sock_def_readable
                sk->sk_prot->init(sk)/tcp_v4_init_sock()
                    tcp_init_xmit_timers(sk)
                        tp->retransmit_timer.function = &tcp_write_timer
                        tp->delack_timer.function = &tcp_delack_timer
                        sk->sk_timer.function = &tcp_keepalive_timer
                        sk->sk_state = TCP_CLOSE
	sock_map_fd(sock)
		file->f_op = SOCK_INODE(sock)->i_fop = &socket_file_ops

bind()/sys_bind() // net/socket.c:1276
    sock->ops->bind(sock)/inet_bind()

listen()/sys_listen() // net/socket.c:1306
    sock->ops->listen(sock)/inet_listen() // net/ipv4/af_inet.c:193
        tcp_listen_start(sk)
            tp->accept_queue = tp->accept_queue_tail = NULL

accept()/sys_accept() // net/socket.c:1340
        sock->ops->accept(sock, newsock)/inet_accept() // net/ipv4/af_inet.c:590
            sk1->sk_prot->accept(sk1)/tcp_accept() // net/ipv4/tcp.c:1895
                wait_for_connect(sk, timeo)
                    prepare_to_wait_exclusive(sk->sk_sleep, &wait, TASK_INTERRUPTIBLE)
                tp->accept_queue->sk

connect()/sys_connect() // net/socket.c:1410
        sock->ops->connect()/inet_stream_connect() // net/ipv4/af_inet.c:504
            sk->sk_prot->connect(sk)/tcp_v4_connect() // net/ipv4/tcp_ipv4.c:747
                tcp_set_state(sk, TCP_SYN_SENT) // include/net/tcp.h:1607
                tcp_v4_hash_connect(sk)
                tcp_connect(sk) // net/ipv4/tcp_output.c:1481
                    __skb_queue_tail(&sk->sk_write_queue, buff)
                    tcp_transmit_skb(sk, skb_clone(buff, GFP_KERNEL))
                    tcp_reset_xmit_timer(sk, TCP_TIME_RETRANS, tp->rto)
            inet_wait_for_connect(sk, timeo)
                prepare_to_wait(sk->sk_sleep, &wait, TASK_INTERRUPTIBLE)

recv()/sys_recv() // net/socket.c:1592
    sock_recvfrom()
        sock_recvmsg()
            __sock_recvmsg()
            sock->ops->recvmsg()/inetsw_array[j]->ops->recvmsg()/inet_stream_ops->recvmsg()/sock_common_recvmsg()
                sk->sk_prot->recvmsg()/inetsw_array[j]->prot->recvmsg()/tcp_prot->recvmsg()/tcp_recvmsg()
                    sk_wait_data()
                        prepare_to_wait(sk->sk_sleep)
                        sk_wait_event()
                            release_sock()
                                __release_sock()
                                    sk->sk_backlog_rcv()/sk->sk_prot->backlog_rcv()/inetsw_array[j]->prot->backlog_rcv()/tcp_prot->backlog_rcv()/tcp_v4_do_rcv()
send()/sys_send() // net/socket.c:1541
    sys_send_to()
        sock_sendmsg(sock, &msg, len)
            __sock_sendmsg(&iocb, sock, msg, size)
                sock->ops->sendmsg(iocb, sock, msg, size)/tcp_sendmsg()
                    skb_entail(sk, tp, skb)
                        __skb_queue_tail(&sk->sk_write_queue, skb)
                        sk->sk_send_head = skb
                    tcp_push_one(sk, mss_now)
                        tcp_transmit_skb(sk, skb_clone(skb, sk->sk_allocation))

shutdown()/sys_shutdown() // net/socket.c:1660
    sock->ops->shutdown(sock, how)/inet_shutdown() // net/ipv4/af_inet.c:671
        sk->sk_prot->shutdown(sk, how)/tcp_shutdown() // net/ipv4/tcp.c:1571
            tcp_close_state()
            tcp_send_fin() // net/ipv4/tcp_output.c:1247
```

### 内核模块
```
static struct inet_protosw inetsw_array[] = { // net/ipv4/af_inet.c:857
    {
        type/SOCK_STREAM
        protocol/IPPROTO_TCP
        prot/tcp_prot // net/ipv4/tcp_ipv4.c:2593
            init/tcp_v4_init_sock
            accept/tcp_accept
            connect/tcp_v4_connect
            sendmsg/tcp_sendmsg
            recvmsg/tcp_recvmsg
            shutdown/tcp_shutdown,
        ops/inet_stream_ops // net/ipv4/af_inet.c:777
            bind/inet_bind
            listen/inet_listen
            accept/inet_accept
            connect/inet_stream_connect
            sendmsg/inet_sendmsg
            recvmsg/sock_common_recvmsg
            shutdown/inet_shutdown
    },
}

struct tcp_func ipv4_specific = { // net/ipv4/tcp_ipv4.c:2023
	syn_recv_sock/tcp_v4_syn_recv_sock
    conn_request/tcp_v4_conn_request
	send_check/tcp_v4_send_check
	queue_xmit/ip_queue_xmit
}

enum { // include/linux/tcp.h:59
  TCP_ESTABLISHED = 1,
  TCP_SYN_SENT,
  TCP_SYN_RECV,
  TCP_FIN_WAIT1,
  TCP_FIN_WAIT2,
  TCP_TIME_WAIT,
  TCP_CLOSE,
  TCP_CLOSE_WAIT,
  TCP_LAST_ACK,
  TCP_LISTEN,
  TCP_CLOSING,	 /* now a valid state */

  TCP_MAX_STATES /* Leave at the end! */
};

module_init(inet_init) net/ipv4/af_inet.c:1115
    inet_init() net/ipv4/af_inet.c:1012
        sock_register()
            net_families[PF_INET]=inet_family_ops
        proto_register(&tcp_prot, 1)
        proto_register(&udp_prot, 1)
        proto_register(&raw_prot, 1)
        arp_init()
        ip_init()
            dev_add_pack()
        tcp_v4_init()
        tcp_init()
        icmp_init()
```

### 网卡驱动
```
e1000_up() // drivers/net/e1000/e1000_main.c:297
    e1000_configure_rx()
		adapter->clean_rx = e1000_clean_rx_irq
    request_irq(e1000_intr)
e1000_intr()
    adapter->clean_rx()/e1000_clean_rx_irq()
        netif_receive_skb(skb)
            deliver_skb(skb, pt_prev)
                pt_recv->func()/ip_rcv() // net/ipv4/ip_input():360
                ip_rcv_finish()
                    dst_input()
                        skb->dst->input()/ip_local_deliver()
                            ip_local_deliver_finish()
                                ipprot->handler()/tcp_v4_rcv()
                                    tcp_v4_do_rcv()
                                        TCP_LISTEN? tcp_v4_hnd_req(sk, skb)
                                            tcp_check_req(sk, skb, req, prev)
                                                tcp_acceptq_queue(sk, req, child)
                                                    tp->accept_queue = req
                                        tcp_rcv_state_process()
tcp_rcv_state_process() // net/ipv4/tcp_input.c:4688
    TCP_LISTEN? tp->af_specific->conn_request()/tcp_func->conn_request()/tcp_v4_conn_request() // net/ipv4/tcp_ipv4.c:1396
        tcp_v4_send_synack(sk, req, dst)
    TCP_SYN_SENT? tcp_rcv_synsent_state_process(sk, skb, th, len)
        th->ack&&th->syn? tcp_ack(sk, skb, FLAG_SLOWPATH)
        th->ack&&th->syn? tcp_set_state(sk, TCP_ESTABLISHED)
			sk_wake_async(sk, 0, POLL_OUT);
    th->rst? tcp_reset()
        tcp_done()
            tcp_set_state(sk, TCP_CLOSE)
    th->ack?
        TCP_SYN_RECV? tcp_set_state(sk, TCP_ESTABLISHED)
        TCP_SYN_RECV? sk_wake_async(sk,0,POLL_OUT)
        TCP_FIN_WAIT1? tcp_set_state(sk, TCP_FIN_WAIT2)
        TCP_CLOSING? tcp_time_wait(sk, TCP_TIME_WAIT, 0)
        TCP_LAST_ACK? tcp_done()
    tcp_urg(sk, skb, th)
    TCP_ESTABLISHED? tcp_data_queue(sk, skb)
            __skb_queue_tail(&sk->sk_receive_queue, skb)
            th->fin? tcp_fin(skb, sk, th)
                TCP_SYN_RECV? tcp_set_state(sk, TCP_CLOSE_WAIT)
                TCP_ESTABLISHED? tcp_set_state(sk, TCP_CLOSE_WAIT)
                TCP_FIN_WAIT1? tcp_send_ack(sk);
                TCP_FIN_WAIT1? tcp_set_state(sk, TCP_CLOSING);
                TCP_FIN_WAIT2? tcp_send_ack(sk);
                TCP_FIN_WAIT2? tcp_set_state(sk, TCP_TIME_WAIT);
            sk->sk_data_ready()/sock_def_readable()
                wake_up_interruptible(sk->sk_sleep)
                sk_wake_async(sk,1,POLL_IN)
```

## 并发编程
### 文件
```
#include <fcntl.h>
#include <unistd.h>
fcntl()sys_fcntl() // fs/fcntl.c:333
open()/sys_open() // fs.open.c:933
    fd = get_unused_fd()
    f = filp_open()
        nd = open_namei(filename)
            path_lookup(name, nd)
                nd->mnt = mntget(current->fs->rootmnt);
                nd->dentry = dget(current->fs->root);
                nd->mnt = mntget(current->fs->pwdmnt);
                nd->dentry = dget(current->fs->pwd);
                link_path_walk(name, nd);
                    __link_path_walk(name, nd);
                        do_lookup(nd, &this, &next);
                            real_lookup(nd->dentry, name, nd);
                                dir->i_op->lookup(dir, dentry, nd)/ext3_lookup()
                                    ext3_find_entry(dentry, &de)
            __lookup_hash(&nd->last, nd->dentry, nd)
                cached_lookup(base, name, nd)
                new = d_alloc(base, name)
                inode->i_op->lookup(inode, new, nd)
            vfs_create(dir->d_inode, path.dentry, mode, nd)
                dir->i_op->create(dir, dentry, mode, nd)/ext3_create()
                    inode = ext3_new_inode(handle, dir, mode)
                        new_inode()
                            alloc_inode(sb)
                                sb->s_op->alloc_inode(sb)
                    inode->i_op = &ext3_file_inode_operations;
                    inode->i_fop = &ext3_file_operations;
        dentry_open(nd.dentry, md.mnt)
            f->f_dentry = dentry
            f->f_vfsmnt = mnt
            inode = dentry->d_inode
            f->f_op = fops_get(inode->i_fop)
            f->f_op->open(inode,f)
    fd_install(fd, f)
        files->fd[fd] = f
read()/sys_read() // fs/read_write.c:312
    file_pos_read()
    vfs_read()
        rw_verify_area(READ, file, pos, count)
            locks_mandatory_area()
                __posix_lock_file(inode, &fl)
                wait_event_interruptible(fl.fl_wait, !fl.fl_next)
        file->f_op->read(file, buf, count, pos)/do_sync_read()
            filp->f_op->aio_read(&kiocb, buf, len, kiocb.ki_pos)/generic_file_aio_read()
                __generic_file_aio_read(iocb, &local_iov, 1, &iocb->ki_pos)
                    do_generic_file_read(filp,ppos,&desc,file_read_actor)
                        do_generic_mapping_read(filp->f_mapping)
                            find_get_page(mapping, index)
                                radix_tree_lookup(&mapping->page_tree, offset)
                                page_cache_get(page)
                        mapping->a_ops->readpage(filp, page)/ext3_readpage()
                            mpage_readpage()
                                do_mpage_readpage()
                                    mpage_bio_submit(READ, bio)
                                        submit_bio(rw, bio)
                                            generic_make_request(bio)
                                                block_wait_queue_running(q)
                                                    prepare_to_wait_exclusive(&rl->drain, &wait, TASK_UNINTERRUPTIBLE)
    file_pos_write()
write()/sys_write() // fs/sys_write():330
    file_pos_read(file)
    vfs_write(file, buf, count, &pos)
        rw_verify_area(WRITE, file, pos, count)
            file->f_op->write(file, buf, count, pos)/do_sync_write()
                filp->f_op->aio_write(&kiocb, buf, len, kiocb.ki_pos)/ext3_file_write()
                    generic_file_aio_write(iocb, buf, count, pos)
                        __generic_file_aio_write_nolock(iocb, &local_iov, 1, &iocb->ki_pos)
                            generic_file_buffered_write(iocb, iov, nr_segs, pos, ppos, count, written)
                            a_ops->commit_write(file, page, offset, offset+bytes)/ext3_writeback_commit_write()
                                generic_commit_write(file, page, from, to)
                                __block_commit_write(inode,page,from,to)
                        sync_page_range(inode, mapping, pos, ret)
    file_pos_write(file, pos)
current:*struct task_struct
    fs:*struct fs_struct
        pwd:*struct dentry
        root:*struct dentry
        altroot:*struct dentry
        pwdmnt:*struct vfsmount
        rootmnt:*struct vfsmount
        altrootmnt:*struct vfsmount
    files:*struct files_struct
    fd:**struct file // include/linux/fs.h:576
        f_pos:loff_t
        f_count:atomic_t
        f_op:*struct file_operations
        f_dentry:*struct dentry // include/linux/dcache.h:83
            d_name:struct qstr
            d_inode:*struct inode // include/linux/fs.h:427
                i_ino:unsigned long
                i_size:loff_t
                i_op:*struct inode_operations/ext3_dir_inode_operations
                    lookup:ext3_lookup
                    create:ext3_create
                i_op:*struct inode_operations/ext3_file_inode_operations
                i_op:*struct inode_operations/ext3_symlink_inode_operations
                i_op:*struct inode_operations/ext3_fast_symlink_inode_operations

                i_fop:*struct file_operations/ext3_dir_operations
                i_fop:*struct file_operations/ext3_file_operations
                    open:generic_file_open
                    ioctl:ext3_ioctl
                    read:do_sync_read
                    write:do_sync_write
                    aio_read:generic_file_aio_read
                    aio_write:ext3_file_write
                    llseek:generic_file_llseek
                    release:ext3_release_file
                i_mapping:*struct address_space
                    a_ops:*struct address_space_operations/ext3_writeback_aops
                        commit_write:ext3_writeback_commit_write
                        writepage:ext3_writeback_writepage
            d_op:*struct dentry_operations
            d_sb:*struct super_block
                s_op:*struct super_operations
        f_vfsmnt:*struct vfsmount

ext3_sops:struct super_operations
	alloc_inode:ext3_alloc_inode
	destroy_inode:ext3_destroy_inode
    read_inode:ext3_read_inode
	write_inode:ext3_write_inode
module_init(init_ext3_fs)
    register_filesystem(&ext3_fs_type)
```

```
/dev/null
/dev/stdio
/dev/stdout
/dev/stderr


#include <unistd.h>
STDIN_FILENO
STDOUT_FILENO
STDERR_FILENO

read()
write()
lseek()
close()

ioctl()

sync()
fsync()
fdatasync()
```

### 信号
signal()

### 进程并发
fork()
exec()
exit()
wait()

### 线程并发
pthread_create()
pthread_self()
pthread_exit()
pthread_join()
pthread_detach()

pthread_mutex_lock()
pthread_mutex_unlock()

pthread_cond_wait()
pthread_cond_signal()

## 系统编程
### 启动
/etc/passwd
/etc/shadow
    fork()
    setsid()
### 配置
### 多路
select()
epoll()
```
```
### 日志
```
#include <errno.h>
errno:int

#include <string.h>
strerror()

#include <stdio.h>
perror()

#include <syslog.h>
syslog()
```

### 用户
```
/etc/passwd
/etc/shadow
/etc/group

getuid()
getpid()
```

### 调试
### 定时
### 延时
alarm()

### 集群
### 存储

## Compile
```
linux-3.7.4/Makefile
include $(srctree)/arch/$(SRCARCH)/Makefile # linux-3.7.4/Makefile:495
```
sudo apt-get install libncurses5-dev
make defconfig
make menuconfig
make
make tags

## GCC
gcc -E hi.c -o hi.i
gcc -S hi.c
gcc -c hi.c

ar -r hi.a hi.o he.o
gcc -shared -fPIC -o hi.so hi.c he.c
gcc -L./ -lhi main.c

nm
readelf -S hi.o
objdump -d hi.o
hexdump -C hi.o

## GRUB
- 下载：<ftp://ftp.gnu.org/gnu/grub/grub-2.00.tar.gz>
BIOS MBR GRUB

## 内核启动
```
startup_32() // arch/x86/kernel/head_32.S:88
i386_start_kernel() // arch/x86/kernel/head32.c:31
start_kernel() // init/main.c:468
    mm_init()
    sched_init()
    console_init()
    signals_init()
    rest_init()
        kernel_thread(kernel_init)
            do_fork() // kernel/fork.c:1548
                copy_process()
                wake_up_new_task() // kernel/sched/core.c:1622
                    activate_task()
                        enqueue_task(rq, p)

kernel_init() // init/main.c:805
    run_init_process()
        kernel_execve() // fs/exec.c:1710
            do_execve()
                do_execve_common()
                    open_exec()
                    sched_exec() // kernel/sched/core.c:2538
                        select_task_rq()
                    bprm_mm_init()
                    prepare_binprm()
                    search_binary_handler()
                        formats[i]->load_binary()

fair_sched_class() // kernel/sched/fair.c:5308
    select_task_rq_fair()
    pick_next_task_fair()
        pick_next_entity()
            __pick_first_entity()
                rb_entry()

formats[i]->load_binary()
    load_elf_binary() // fs/binfmt_elf.c:561
        current->mm->start_stack
        set_brk(elf_bss, elf_brk)

        start_thread() // arch/x86/kernel/process_32.c:200
formats[i]->load_binary()
    load_script()
        bprm_change_interp()
        open_exec()
        prepare_binprm()
        search_binary_handler()

```
